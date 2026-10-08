package infrastructure

import (
	"database/sql"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MySQL connection pool metrics
	mysqlOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_open_connections",
			Help: "Number of open MySQL connections",
		},
	)

	mysqlInUseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_in_use_connections",
			Help: "Number of MySQL connections currently in use",
		},
	)

	mysqlIdleConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_idle_connections",
			Help: "Number of idle MySQL connections",
		},
	)

	mysqlMaxOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_max_open_connections",
			Help: "Maximum number of open MySQL connections",
		},
	)

	mysqlWaitCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_wait_count_total",
			Help: "Total number of connections waited for",
		},
	)

	mysqlWaitDuration = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_wait_duration_seconds_total",
			Help: "Total time waited for connections in seconds",
		},
	)

	// Go runtime / GC metrics
	goGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_goroutines_count",
			Help: "Number of goroutines",
		},
	)

	goGCPauseSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_gc_pause_seconds_last",
			Help: "Duration of last GC pause in seconds",
		},
	)

	goGCPauseTotalSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_gc_pause_seconds_total",
			Help: "Total GC pause time in seconds",
		},
	)

	goGCCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_gc_count_total",
			Help: "Total number of GC cycles",
		},
	)

	goMemoryAllocBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_memory_alloc_bytes",
			Help: "Current memory allocation in bytes",
		},
	)

	goMemorySysBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_memory_sys_bytes",
			Help: "Total memory obtained from system in bytes",
		},
	)

	goMemoryHeapInuseBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_memory_heap_inuse_bytes",
			Help: "Heap memory in use in bytes",
		},
	)
)

// MySQLMetricsCollector collects MySQL connection pool metrics
type MySQLMetricsCollector struct {
	db *sql.DB
}

// NewMySQLMetricsCollector creates a new MySQL metrics collector
func NewMySQLMetricsCollector(db *sql.DB) *MySQLMetricsCollector {
	return &MySQLMetricsCollector{db: db}
}

// Collect updates all MySQL connection pool metrics
func (c *MySQLMetricsCollector) Collect() {
	if c.db == nil {
		return
	}

	stats := c.db.Stats()

	mysqlOpenConnections.Set(float64(stats.OpenConnections))
	mysqlInUseConnections.Set(float64(stats.InUse))
	mysqlIdleConnections.Set(float64(stats.Idle))
	mysqlMaxOpenConnections.Set(float64(stats.MaxOpenConnections))
	mysqlWaitCount.Set(float64(stats.WaitCount))
	mysqlWaitDuration.Set(stats.WaitDuration.Seconds())
}

// RuntimeMetricsCollector collects Go runtime and GC metrics
type RuntimeMetricsCollector struct{}

// NewRuntimeMetricsCollector creates a new runtime metrics collector
func NewRuntimeMetricsCollector() *RuntimeMetricsCollector {
	return &RuntimeMetricsCollector{}
}

// Collect updates all runtime and GC metrics
func (c *RuntimeMetricsCollector) Collect() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	goGoroutines.Set(float64(runtime.NumGoroutine()))
	goGCCount.Set(float64(memStats.NumGC))
	goMemoryAllocBytes.Set(float64(memStats.Alloc))
	goMemorySysBytes.Set(float64(memStats.Sys))
	goMemoryHeapInuseBytes.Set(float64(memStats.HeapInuse))

	// Total GC pause time
	goGCPauseTotalSeconds.Set(float64(memStats.PauseTotalNs) / 1e9)

	// Last GC pause duration
	if memStats.NumGC > 0 {
		lastPauseIdx := (memStats.NumGC + 255) % 256
		goGCPauseSeconds.Set(float64(memStats.PauseNs[lastPauseIdx]) / 1e9)
	}
}
