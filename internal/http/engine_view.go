package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Engine protection Phase 5b: what the app shows on screen (report, cut list,
// machining preview, review gates) and the review actions, computed on the
// private engine for the protected build. One route per op; the op must be on
// the engine's display list (engine.viewOps), else 404 before any network
// call. Same auth, 4 MB body, timeout and error mapping as the other engine
// routes; the display quota bucket (ENGINE_VIEW_QUOTA_*). Priced ops get the
// org book injected; the others have any client "priceBook" removed.

// engineView godoc
// @Summary Display result (report, cut list, review gate)
// @Description Runs one display op on the private engine with the app's arguments (`{"args": [...]}`); priced ops use the org price book. Display quota.
// @Tags engine
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param op path string true "Display op, e.g. createProjectReport"
// @Param body body dto.EngineViewBody true "The op's arguments"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/view/{op} [post]
func (h *handlers) engineView(c *gin.Context) {
	logEnter(c, "engine_view")
	op := engine.ViewPrefix + c.Param("op")
	if spec, ok := engine.SpecFor(op); !ok || !spec.View {
		middleware.Abort(c, utils.E(utils.CodeNotFound))
		return
	}
	h.proxyEngine(c, op)
}

// engineProductionPacketPDF godoc
// @Summary Production packet PDF
// @Description Forwards the cabinet project, room and packet inputs to the private engine. Returns the production packet PDF (base64), watermarked with the org id. No price book (it prices from the project's own preferences, as the app does). Document quota.
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineProductionPacketBody true "Project, room and packet inputs"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/production-packet-pdf [post]
func (h *handlers) engineProductionPacketPDF(c *gin.Context) {
	logEnter(c, "engine_production_packet_pdf")
	h.proxyEngine(c, engine.OpProductionPacketPDF)
}
