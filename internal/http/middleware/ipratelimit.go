package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func IPRateLimit(rps rate.Limit, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]*rate.Limiter{}

	limiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		lim, ok := buckets[ip]
		if !ok {
			lim = rate.NewLimiter(rps, burst)
			buckets[ip] = lim
		}
		return lim
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" {
			c.Next()
			return
		}
		lim := limiter(c.ClientIP())
		if !lim.Allow() {
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))
			WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many requests", nil)
			return
		}
		c.Next()
	}
}
