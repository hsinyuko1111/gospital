package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/hsinyuko1111/gospital/internal/infrastructure"
)

// RootHandler handles root-level requests (health, metrics, swagger, etc.)
type RootHandler struct {
	mysqlClient    *infrastructure.MySQLClient
	dynamoClient   *infrastructure.DynamoDBClient
	swaggerHandler gin.HandlerFunc
}

// NewRootHandler creates a new root handler
func NewRootHandler(mysqlClient *infrastructure.MySQLClient, dynamoClient *infrastructure.DynamoDBClient) *RootHandler {
	return &RootHandler{
		mysqlClient:    mysqlClient,
		dynamoClient:   dynamoClient,
		swaggerHandler: ginSwagger.WrapHandler(swaggerFiles.Handler),
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Check the health status of the API and its dependencies
// @Tags Root
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,mysql=string,dynamodb=string}
// @Failure 503 {object} object{status=string,mysql=string,mysql_error=string,dynamodb=string}
// @Router /health [get]
func (h *RootHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	health := gin.H{
		"status":   "healthy",
		"mysql":    "unknown",
		"dynamodb": "unknown",
	}

	// Check MySQL connection
	if err := h.mysqlClient.PingContext(ctx); err != nil {
		health["mysql"] = "unhealthy"
		health["mysql_error"] = err.Error()
		health["status"] = "degraded"
	} else {
		health["mysql"] = "healthy"
	}

	// Note: DynamoDB client doesn't have a simple ping method
	// The circuit breaker state can indicate health
	// For now, just mark as healthy if client exists
	if h.dynamoClient != nil {
		health["dynamodb"] = "healthy"
	}

	// Return appropriate status code
	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// Metrics godoc
// @Summary Prometheus metrics
// @Description Expose Prometheus metrics for scraping
// @Tags Root
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *RootHandler) Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// Swagger godoc
// @Summary Swagger UI
// @Description Serve Swagger UI for API documentation
// @Tags Root
// @Produce html
// @Success 200 {string} string "Swagger UI"
// @Router /swagger/{any} [get]
func (h *RootHandler) Swagger(c *gin.Context) {
	h.swaggerHandler(c)
}
