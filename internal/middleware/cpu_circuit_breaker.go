package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
)

type ECSTaskMetadata struct {
	Limits struct {
		CPU    float64 `json:"CPU"`
		Memory int     `json:"Memory"`
	} `json:"Limits"`
}

type ECSTaskStats struct {
	Read    time.Time    `json:"read"`
	Preread time.Time    `json:"preread"`
	CPU     *ECSCPUStats `json:"cpu_stats"`
	PreCPU  *ECSCPUStats `json:"precpu_stats"`
}

type ECSCPUStats struct {
	CPUUsage *ECSCPUUsage `json:"cpu_usage"`
}

type ECSCPUUsage struct {
	TotalUsage uint64 `json:"total_usage"`
}

type CPUCircuitBreaker struct {
	mu                sync.RWMutex
	isOpen            bool
	currentCPU        float64
	overloadThreshold float64
	recoveryThreshold float64
	checkInterval     time.Duration
	metadataURI       string
	httpClient        *http.Client
	taskCPULimit      float64
}

func NewCPUCircuitBreaker(overloadThreshold, recoveryThreshold float64) *CPUCircuitBreaker {
	metadataURI := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if metadataURI == "" {
		return &CPUCircuitBreaker{
			isOpen:      false,
			metadataURI: "",
		}
	}

	cb := &CPUCircuitBreaker{
		overloadThreshold: overloadThreshold,
		recoveryThreshold: recoveryThreshold,
		checkInterval:     100 * time.Millisecond,
		isOpen:            false,
		currentCPU:        0.0,
		metadataURI:       metadataURI,
		httpClient: &http.Client{
			Timeout: 500 * time.Millisecond,
		},
	}

	if err := cb.fetchTaskCPULimit(); err != nil {
		return &CPUCircuitBreaker{
			isOpen:      false,
			metadataURI: "",
		}
	}

	if runtime.GOMAXPROCS(0) < 2 {
		runtime.GOMAXPROCS(2)
	}

	go func() {
		runtime.LockOSThread()
		cb.monitorCPU()
	}()

	log.Printf("[CPU Circuit Breaker] Initialized (overload: %.1f%%, recovery: %.1f%%, limit: %.2f vCPU, interval: %v, GOMAXPROCS: %d)",
		overloadThreshold, recoveryThreshold, cb.taskCPULimit, cb.checkInterval, runtime.GOMAXPROCS(0))

	return cb
}

func (cb *CPUCircuitBreaker) fetchTaskCPULimit() error {
	resp, err := cb.httpClient.Get(cb.metadataURI + "/task")
	if err != nil {
		return fmt.Errorf("failed to call /task endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("/task endpoint returned status %d", resp.StatusCode)
	}

	var taskMeta ECSTaskMetadata
	if err := json.NewDecoder(resp.Body).Decode(&taskMeta); err != nil {
		return fmt.Errorf("failed to parse task metadata: %w", err)
	}

	if taskMeta.Limits.CPU <= 0 {
		return fmt.Errorf("invalid CPU limit: %f", taskMeta.Limits.CPU)
	}

	cb.taskCPULimit = taskMeta.Limits.CPU
	return nil
}

func (cb *CPUCircuitBreaker) monitorCPU() {
	if cb.metadataURI == "" {
		return
	}

	ticker := time.NewTicker(cb.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		cpuUsage, err := cb.getContainerCPU()
		if err != nil {
			continue
		}

		cb.mu.Lock()
		cb.currentCPU = cpuUsage

		shouldOpen := !cb.isOpen && cpuUsage >= cb.overloadThreshold
		shouldClose := cb.isOpen && cpuUsage < cb.recoveryThreshold

		if shouldOpen {
			cb.isOpen = true
		} else if shouldClose {
			cb.isOpen = false
		}

		UpdateCPUCircuitBreakerMetrics(cb.isOpen, cb.currentCPU)

		cb.mu.Unlock()

		if shouldOpen {
			log.Printf("[CPU Circuit Breaker] Circuit OPENED - CPU: %.1f%% (threshold: %.1f%%)",
				cpuUsage, cb.overloadThreshold)
		}
		if shouldClose {
			log.Printf("[CPU Circuit Breaker] Circuit CLOSED - CPU: %.1f%% (threshold: %.1f%%)",
				cpuUsage, cb.recoveryThreshold)
		}
	}
}

func (cb *CPUCircuitBreaker) getContainerCPU() (float64, error) {
	resp, err := cb.httpClient.Get(cb.metadataURI + "/task/stats")
	if err != nil {
		return 0, fmt.Errorf("failed to call /task/stats endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("/task/stats endpoint returned status %d", resp.StatusCode)
	}

	var statsMap map[string]*ECSTaskStats
	if err := json.NewDecoder(resp.Body).Decode(&statsMap); err != nil {
		return 0, fmt.Errorf("failed to parse stats response: %w", err)
	}

	var totalVCPUsUsed float64
	validStatsFound := false

	for _, stats := range statsMap {
		if stats == nil || stats.CPU == nil || stats.PreCPU == nil {
			continue
		}
		if stats.CPU.CPUUsage == nil || stats.PreCPU.CPUUsage == nil {
			continue
		}

		timeDelta := stats.Read.Sub(stats.Preread).Nanoseconds()
		if timeDelta <= 0 {
			continue
		}

		cpuDelta := stats.CPU.CPUUsage.TotalUsage - stats.PreCPU.CPUUsage.TotalUsage
		vCPUs := float64(cpuDelta) / float64(timeDelta)
		totalVCPUsUsed += vCPUs
		validStatsFound = true
	}

	if !validStatsFound {
		return 0.0, nil
	}

	cpuPercent := (totalVCPUsUsed / cb.taskCPULimit) * 100.0
	if cpuPercent < 0 {
		cpuPercent = 0
	}

	return cpuPercent, nil
}

func (cb *CPUCircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.isOpen
}

func (cb *CPUCircuitBreaker) GetCurrentCPU() float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentCPU
}

func CPUCircuitBreakerMiddleware(cb *CPUCircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cb.IsOpen() {
			IncrementCPURejections()

			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "SERVICE_OVERLOADED",
					Message: "Service temporarily overloaded due to high CPU usage, please retry later",
					Details: []string{
						fmt.Sprintf("Current CPU: %.1f%%", cb.GetCurrentCPU()),
					},
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
