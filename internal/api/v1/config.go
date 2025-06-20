package v1

// Config module error codes: 10400-10499
// 10400-10499: Reserved for config-related errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
)

// Config module error codes: 11000-11999
var ErrCodeConfig = v1errors.NewErrorCode(v1errors.CheckCode(11000), "config error")

// GetConfigSupabase implements the GET /config/supabase endpoint
func (h *V1Handler) GetConfigSupabase(c *gin.Context) {
	response := SupabaseConfig{
		ProjectReference: h.supabaseConfig.ProjectReference,
		ApiAnnoKey:       h.supabaseConfig.ApiAnnoKey,
		BucketNames: struct {
			PostImages string `json:"post_images"`
		}{
			PostImages: h.supabaseConfig.BucketNames.PostImages,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
