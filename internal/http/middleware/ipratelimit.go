package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimit limits each caller IP. The IP comes from clientIP, so behind
// Render/Cloudflare every visitor gets their own bucket instead of sharing the
// proxy's address.
func IPRateLimit(rps rate.Limit, burst int) gin.HandlerFunc {
	lim := newKeyedLimiter(rps, burst)
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" || strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}
		if !lim.allow(clientIP(c)) {
			rejectRateLimited(c, burst)
			return
		}
		c.Next()
	}
}
