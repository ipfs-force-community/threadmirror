package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const authInfoKey = "auth.auth_info"

func Middleware(
	v JWTVerifier,
	errorHandler func(c *gin.Context, statusCode int),
) gin.HandlerFunc {
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, statusCode int) {
			c.AbortWithStatusJSON(statusCode, gin.H{
				"error": http.StatusText(statusCode),
			})
		}
	}

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			errorHandler(c, http.StatusUnauthorized)
			return
		}

		jwtString := strings.TrimPrefix(authorization, "Bearer ")
		if ai, err := v.Verify(jwtString); err != nil {
			fmt.Println("[auth.middleware] verify token error:", err)
			errorHandler(c, http.StatusForbidden)
		} else {
			SetAuthInfo(c, ai)
			c.Next()
		}
	}
}

func SetAuthInfo(c *gin.Context, ai *AuthInfo) {
	c.Set(authInfoKey, ai)
}

func GetAuthInfo(c *gin.Context) (ai *AuthInfo, ok bool) {
	var val any
	val, ok = c.Get(authInfoKey)
	if !ok {
		return
	}
	ai, ok = val.(*AuthInfo)
	return
}

func MustAuthInfo(c *gin.Context) (ai *AuthInfo) {
	var ok bool
	ai, ok = GetAuthInfo(c)
	if !ok {
		panic("auth info not found")
	}
	return
}

func CurrentUserID(c *gin.Context) string {
	ai, ok := GetAuthInfo(c)
	if !ok {
		return ""
	}
	return ai.UserID
}
