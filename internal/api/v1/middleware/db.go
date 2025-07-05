package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

// DBInjector 返回一个gin中间件，将*sql.DB注入到gin.Context的context中
func DBInjector(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := sql.WithDBToContext(c.Request.Context(), db)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
