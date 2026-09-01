package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	krlimit "github.com/go-kratos/aegis/ratelimit"
	"github.com/go-kratos/aegis/ratelimit/bbr"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

// BBR is process-wide adaptive load-shedding (Kratos aegis BBR).
// It is not an IP quota. Per-IP limits stay on IPRateLimit.
// /healthz and /readyz are never shed so probes stay up.
// When cfg.Enabled is false the limiter is a pass-through.
func BBR(cfg config.BBR) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	limiter := bbr.NewLimiter(
		bbr.WithWindow(cfg.WindowSize.Duration()),
		bbr.WithBucket(cfg.Buckets),
		bbr.WithCPUThreshold(cfg.CpuThreshold),
	)
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" {
			c.Next()
			return
		}
		done, err := limiter.Allow()
		if err != nil {
			WriteError(c, http.StatusTooManyRequests, "rate_limited", "Service is overloaded", gin.H{
				"limiter": "bbr",
			})
			return
		}
		defer done(krlimit.DoneInfo{})
		c.Next()
	}
}
