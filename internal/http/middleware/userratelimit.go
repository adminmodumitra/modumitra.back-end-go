package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// UserRateLimit limits each signed-in user across all their IPs and devices.
// Mount after JWTAuth; requests without a user id pass through.
func UserRateLimit(rps rate.Limit, burst int) gin.HandlerFunc {
	lim := newKeyedLimiter(rps, burst)
	return func(c *gin.Context) {
		uid := UserIDFrom(c)
		if uid != "" && !lim.allow(uid) {
			rejectRateLimited(c, burst)
			return
		}
		c.Next()
	}
}
