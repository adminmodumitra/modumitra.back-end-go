package httpapi

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
)

// Option wires optional dependencies into NewRouter.
type Option func(*handlers)

// WithPriceBooks sets the org price-book store (Postgres in cmd/api).
func WithPriceBooks(repo pricebook.Repo) Option {
	return func(h *handlers) { h.priceBooks = repo }
}

// engineGrace lets the engine client's own timeout fire first, so a slow
// engine answers 504 engine_timeout rather than the group's generic 503.
// Keep ENGINE_TIMEOUT + engineGrace under the server WriteTimeout (15s).
const engineGrace = 2 * time.Second

// mountEngine registers /v1/engine/* in its own signed-in group: the same auth,
// CSRF/Origin guard and rate-limit buckets as the authed group (passed in),
// but a 4 MB body limit for design documents and the engine timeout.
func mountEngine(r *gin.Engine, h *handlers, bbr gin.HandlerFunc, auth ...gin.HandlerFunc) {
	g := r.Group("/")
	g.Use(bbr)
	g.Use(middleware.Timeout(h.cfg.Engine.Timeout + engineGrace))
	g.Use(middleware.BodyLimit(engine.MaxRequestBytes))
	g.Use(auth...)
	g.POST("/v1/engine/quote", h.engineQuote)
	g.POST("/v1/engine/estimate", h.engineEstimate)
}
