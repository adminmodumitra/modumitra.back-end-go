package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func skipJWT(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/v1/public/config":
		return true
	default:
		return false
	}
}

// JWTAuth skips ops/public config. Slice 0 does not issue or validate JWTs:
// every other route is 401 unauthenticated.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipJWT(c.Request.URL.Path) {
			c.Next()
			return
		}
		c.Header("WWW-Authenticate", "Bearer")
		WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication required", nil)
	}
}
