package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := make(chan struct{})
		panicChan := make(chan interface{})

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case p := <-panicChan:
			panic(p)
		case <-time.After(timeout):
			IncrementTimeoutRejections()
			c.Abort()
			c.JSON(http.StatusGatewayTimeout, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "REQUEST_TIMEOUT",
					Message: "Request took too long to process",
					Details: []string{
						"Timeout: " + timeout.String(),
					},
				},
			})
		}
	}
}
