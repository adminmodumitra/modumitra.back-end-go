package httpapi

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// engineQuote godoc
// @Summary Price a design (live quote)
// @Description Forwards the design to the private pricing engine with the org price book. Quota: per user and per org per minute, plus a daily per-org cap.
// @Tags pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineQuoteBody true "Design document"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/quote [post]
func (h *handlers) engineQuote(c *gin.Context) {
	logEnter(c, "engine_quote")
	h.proxyEngine(c, engine.OpQuote)
}

// engineEstimate godoc
// @Summary Estimate a design (summary)
// @Description Forwards the design to the private pricing engine with the org price book. Shares the quote quota.
// @Tags pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineEstimateBody true "Design document"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/estimate [post]
func (h *handlers) engineEstimate(c *gin.Context) {
	logEnter(c, "engine_estimate")
	h.proxyEngine(c, engine.OpEstimate)
}

// proxyEngine: read body → resolve org → quota → price book (ops that price)
// → engine call → copy the answer back. Logs request id, status and latency,
// never bodies.
func (h *handlers) proxyEngine(c *gin.Context, op string) {
	spec, known := engine.SpecFor(op)
	if h.engine == nil || !known {
		middleware.Abort(c, utils.E(utils.CodeEngineUnavailable))
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		middleware.Abort(c, bodyReadError(err))
		return
	}
	orgID, _, err := h.callerOrg(c)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	uid := middleware.UserIDFrom(c)
	quotaKey := orgID
	if quotaKey == "" {
		quotaKey = "usr:" + uid // personal account: its own org budget
	}
	quota := h.quota
	switch {
	case spec.Manufacturing:
		quota = h.mfgQuota
	case spec.Document:
		quota = h.docQuota
	case spec.View:
		quota = h.viewQuota
	}
	if ok, wait := quota.Allow(uid, quotaKey); !ok {
		middleware.RejectQuota(c, wait)
		return
	}
	ctx := c.Request.Context()
	var book json.RawMessage
	if spec.NeedsBook {
		book, err = h.orgPriceBook(ctx, orgID)
		if err != nil {
			middleware.Abort(c, err) // never price with defaults when the book failed to load
			return
		}
	}
	body, err := engine.BuildBodyFor(raw, spec, book)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	rid := middleware.RequestIDFrom(c)
	start := time.Now()
	resp, err := h.engine.Do(ctx, op, body, engine.Caller{OrgID: orgID, UserID: uid, RequestID: rid})
	latency := time.Since(start)
	log := obs.Logger()
	lg := log.Info().Str("op", "engine_"+op).Str("request_id", rid).Str("org_id", orgID).Int64("latency_ms", latency.Milliseconds())
	if err != nil {
		timeout := engine.IsTimeout(err)
		lg.Bool("timeout", timeout).Msg("engine call failed")
		if timeout {
			middleware.Abort(c, utils.E(utils.CodeEngineTimeout))
		} else {
			middleware.Abort(c, utils.E(utils.CodeEngineError))
		}
		return
	}
	defer resp.Body.Close()
	lg.Int("engine_status", resp.StatusCode).Msg("engine call")
	relayEngine(c, resp, engine.MaxResponseFor(op))
}
