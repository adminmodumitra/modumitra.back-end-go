package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
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
		c.Header(utils.HeaderWWWAuth, "Bearer")
		Abort(c, utils.E(utils.CodeUnauthenticated))
	}
}
