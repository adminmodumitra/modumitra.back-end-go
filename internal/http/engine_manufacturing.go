package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
)

// Engine protection Phase 4: manufacturing goes server-side. Same group, auth,
// body limit and error mapping as quote/estimate; metered in the separate
// manufacturing bucket (ENGINE_MFG_QUOTA_*).

// engineCutlist godoc
// @Summary Production cut list (rows + CSV)
// @Description Forwards the cabinet project to the private engine: the cut-list rows and the CSV file, byte-identical to the app's local export. Manufacturing quota.
// @Tags manufacturing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineManufacturingBody true "Cabinet project and optional room scope"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/cutlist [post]
func (h *handlers) engineCutlist(c *gin.Context) {
	logEnter(c, "engine_cutlist")
	h.proxyEngine(c, engine.OpCutlist)
}

// engineCNC godoc
// @Summary CNC machine files
// @Description Forwards the cabinet project to the private engine: every implemented machine file (name + text contents). Manufacturing quota.
// @Tags manufacturing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineManufacturingBody true "Cabinet project and optional room scope"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/cnc [post]
func (h *handlers) engineCNC(c *gin.Context) {
	logEnter(c, "engine_cnc")
	h.proxyEngine(c, engine.OpCNC)
}

// engineProductionPackage godoc
// @Summary Production outputs package
// @Description Forwards the cabinet project and room to the private engine with the org price book (its costing and quote files price): the package files. Manufacturing quota.
// @Tags manufacturing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineProductionPackageBody true "Cabinet project, room and optional room scope"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/production-package [post]
func (h *handlers) engineProductionPackage(c *gin.Context) {
	logEnter(c, "engine_production_package")
	h.proxyEngine(c, engine.OpProductionPackage)
}
