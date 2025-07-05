package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHealth returns service health status
func (h *V1Handler) GetHealth(c *gin.Context) {
	// In a real application, you might want to check:
	// - Database connectivity
	// - External service dependencies
	// - Resource usage
	// For now, we'll just return a simple OK status

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}
