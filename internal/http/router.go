package httpapi

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func NewRouter(cfg config.Config, pinger store.Pinger) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)

	chain := []gin.HandlerFunc{
		middleware.RequestID(),
		middleware.Recover(),
	}
	chain = append(chain, middleware.OTEL(cfg.ServiceName)...)
	chain = append(chain,
		middleware.BBR(cfg.BBR),
		middleware.Timeout(cfg.RequestTimeout),
		middleware.BodyLimit(cfg.BodyLimitBytes),
		middleware.IPRateLimit(rate.Limit(30), 60),
		middleware.UserRateLimit(),
		middleware.JWTAuth(),
		middleware.Tenant(),
		middleware.RBAC(),
		middleware.Idempotency(),
	)
	r.Use(chain...)

	h := &handlers{cfg: cfg, pinger: pinger}
	r.GET("/healthz", h.healthz)
	r.GET("/readyz", h.readyz)
	r.GET("/v1/public/config", h.publicConfig)
	r.NoRoute(h.noRoute)

	return r
}
