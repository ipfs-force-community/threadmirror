package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
)

// ErrorHandler captures errors and returns a consistent JSON error response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last().Err
			if apiErr, ok := lastErr.(*errors.APIError); ok {
				c.JSON(apiErr.Status, gin.H{
					"code":    apiErr.ErrorCode.Code,
					"message": apiErr.ErrorCode.Message,
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    errors.ErrCodeInternalError.Code,
				"message": errors.ErrCodeInternalError.Message,
			})
		}
	}
}
