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
// Document routes (5a) lift both server deadlines for their own connection.
const engineGrace = 2 * time.Second

// docUploadWindow is how long a document request may take to arrive (a 32 MB
// body on a slow uplink), on top of the engine time.
const docUploadWindow = 60 * time.Second

// mountEngine registers /v1/engine/* in its own signed-in group: the same auth,
// CSRF/Origin guard and rate-limit buckets as the authed group (passed in),
// but a 4 MB body limit for design documents and the engine timeout. Phase 4
// adds the manufacturing routes (their own quota bucket, same limits).
func mountEngine(r *gin.Engine, h *handlers, bbr gin.HandlerFunc, auth ...gin.HandlerFunc) {
	g := r.Group("/")
	g.Use(bbr)
	g.Use(middleware.Timeout(h.cfg.Engine.Timeout + engineGrace))
	g.Use(middleware.BodyLimit(engine.MaxRequestBytes))
	g.Use(auth...)
	g.POST("/v1/engine/quote", h.engineQuote)
	g.POST("/v1/engine/estimate", h.engineEstimate)
	g.POST("/v1/engine/cutlist", h.engineCutlist)
	g.POST("/v1/engine/cnc", h.engineCNC)
	g.POST("/v1/engine/production-package", h.engineProductionPackage)

	// Phase 5a: documents built on the server. Same auth and guards; a 32 MB
	// body (proposal stills), ENGINE_DOC_TIMEOUT, and this connection's read
	// and write deadlines moved past the server-wide 10s/15s.
	docTimeout := h.cfg.Engine.DocTimeout
	if docTimeout < h.cfg.Engine.Timeout {
		docTimeout = h.cfg.Engine.Timeout
	}
	d := r.Group("/")
	d.Use(middleware.ExtendDeadlines(docUploadWindow, docUploadWindow+docTimeout+engineGrace))
	d.Use(bbr)
	d.Use(middleware.Timeout(docUploadWindow + docTimeout + engineGrace))
	d.Use(middleware.BodyLimit(engine.MaxDocumentRequestBytes))
	d.Use(auth...)
	d.POST("/v1/engine/proposal-pdf", h.engineProposalPDF)
	d.POST("/v1/engine/revision-summary-pdf", h.engineRevisionSummaryPDF)
	d.POST("/v1/engine/quote-export", h.engineQuoteExport)
}
