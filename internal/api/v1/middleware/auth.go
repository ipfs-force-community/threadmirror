package v1

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/ipfs-force-community/threadmirror/internal/api/v1"
)

func Authentication(m gin.HandlerFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		if _, ok := c.Get(v1.BearerAuthScopes); ok {
			m(c)
		} else {
			c.Next()
		}
	}
}
