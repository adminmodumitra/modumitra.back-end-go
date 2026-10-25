package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods  = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders  = "Accept, Authorization, Content-Type, Idempotency-Key, X-Request-Id"
	corsExposeHeaders = "X-Request-Id, Retry-After, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset"
	corsMaxAge        = "600"
)

// CORS lets browser apps on the listed origins call the API.
// An empty list adds no headers (check off). Origins match exactly, without a
// trailing slash; "*" allows any origin. Auth is a bearer token, not a cookie,
// so credentials are never allowed. A preflight from an allowed origin returns
// 204 here, before the IP allowlist and the route groups, which do not
// register OPTIONS.
func CORS(origins []string) gin.HandlerFunc {
	allowed, anyOrigin := parseOrigins(origins)
	if len(allowed) == 0 && !anyOrigin {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		c.Writer.Header().Add("Vary", "Origin")
		if _, ok := allowed[strings.TrimRight(origin, "/")]; !ok && !anyOrigin {
			c.Next()
			return
		}
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Expose-Headers", corsExposeHeaders)
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			h.Set("Access-Control-Allow-Methods", corsAllowMethods)
			h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
			h.Set("Access-Control-Max-Age", corsMaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func parseOrigins(origins []string) (map[string]struct{}, bool) {
	allowed := map[string]struct{}{}
	anyOrigin := false
	for _, raw := range origins {
		entry := strings.TrimRight(strings.TrimSpace(raw), "/")
		switch entry {
		case "":
		case "*":
			anyOrigin = true
		default:
			allowed[entry] = struct{}{}
		}
	}
	return allowed, anyOrigin
}
