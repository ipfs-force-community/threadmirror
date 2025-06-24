package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBotStats returns bot statistics
func (h *V1Handler) GetBotStats(c *gin.Context) {
	rawStats := h.twitterBot.GetStats()

	// Convert to API response format
	stats := BotStats{
		Enabled:        rawStats["enabled"].(bool),
		CheckInterval:  rawStats["check_interval"].(string),
		ProcessedCount: rawStats["processed_count"].(int),
	}

	if username, ok := rawStats["username"].(string); ok && username != "" {
		stats.Username = &username
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
