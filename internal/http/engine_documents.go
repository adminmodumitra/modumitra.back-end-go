package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
)

// Engine protection Phase 5a: documents built on the server, watermarked with
// the org id (X-Org-Id) by the engine. Same auth and error mapping as the other
// engine routes; a 32 MB body, ENGINE_DOC_TIMEOUT and the document quota
// bucket (ENGINE_DOC_QUOTA_*). PDFs come back as base64 in the JSON result.

// engineProposalPDF godoc
// @Summary Client proposal PDF
// @Description Forwards the design, view stills and stale override to the private engine with the org price book. Returns the proposal PDF (base64) and its file name, watermarked with the org id. Document quota.
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineProposalPDFBody true "Design, stills and options"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/proposal-pdf [post]
func (h *handlers) engineProposalPDF(c *gin.Context) {
	logEnter(c, "engine_proposal_pdf")
	h.proxyEngine(c, engine.OpProposalPDF)
}

// engineRevisionSummaryPDF godoc
// @Summary Revision and approval summary PDF
// @Description Forwards the cabinet project to the private engine. Returns the revision summary PDF (base64), watermarked with the org id. No price book (it prices from the project's own preferences, as the app does). Document quota.
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineRevisionSummaryBody true "Cabinet project"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/revision-summary-pdf [post]
func (h *handlers) engineRevisionSummaryPDF(c *gin.Context) {
	logEnter(c, "engine_revision_summary_pdf")
	h.proxyEngine(c, engine.OpRevisionSummaryPDF)
}

// engineQuoteExport godoc
// @Summary BOQ / quote export bundle
// @Description Forwards the cabinet project and room to the private engine with the org price book. Returns the quote and BOQ CSVs, the issued-quote CSV, the Excel 2003 workbook and quote/invoice JSON, each watermarked with the org id. Document quota.
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Organization-Id header string false "Organization ID (default: caller's primary org)"
// @Param body body dto.EngineQuoteExportBody true "Cabinet project and room"
// @Success 200 {object} dto.EngineResultOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Failure 503 {object} utils.ErrorResponse
// @Failure 504 {object} utils.ErrorResponse
// @Router /v1/engine/quote-export [post]
func (h *handlers) engineQuoteExport(c *gin.Context) {
	logEnter(c, "engine_quote_export")
	h.proxyEngine(c, engine.OpQuoteExport)
}
