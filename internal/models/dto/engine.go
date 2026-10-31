package dto

import (
	"encoding/json"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// EngineQuoteBody is POST /v1/engine/quote. document is the InteriorProject
// JSON. The server picks the price book; a client "priceBook" is replaced.
type EngineQuoteBody struct {
	Document      json.RawMessage `json:"document" swaggertype:"object"`
	Now           string          `json:"now" example:"2026-10-31T10:00:00Z"`
	StaleOverride *bool           `json:"staleOverride,omitempty"`
}

// EngineEstimateBody is POST /v1/engine/estimate.
type EngineEstimateBody struct {
	Document json.RawMessage `json:"document" swaggertype:"object"`
}

// EngineManufacturingBody is POST /v1/engine/cutlist and /v1/engine/cnc
// (Phase 4). project is the app's CabinetProject JSON; scope is a room id or
// "whole-project", omitted for the app's unscoped export.
type EngineManufacturingBody struct {
	Project json.RawMessage `json:"project" swaggertype:"object"`
	Scope   string          `json:"scope,omitempty" example:"room-1"`
}

// EngineProductionPackageBody is POST /v1/engine/production-package. room is
// the RoomConfig the app reports with; the server picks the price book.
type EngineProductionPackageBody struct {
	Project json.RawMessage `json:"project" swaggertype:"object"`
	Room    json.RawMessage `json:"room" swaggertype:"object"`
	Scope   string          `json:"scope,omitempty" example:"room-1"`
}

// EngineProposalPDFBody is POST /v1/engine/proposal-pdf (Phase 5a). frames
// are the proposal's view stills ({cameraId, viewName, dataUrl}); the server
// picks the price book.
type EngineProposalPDFBody struct {
	Document      json.RawMessage `json:"document" swaggertype:"object"`
	Frames        json.RawMessage `json:"frames,omitempty" swaggertype:"array,object"`
	Now           string          `json:"now" example:"2026-10-31T10:00:00Z"`
	StaleOverride *bool           `json:"staleOverride,omitempty"`
}

// EngineRevisionSummaryBody is POST /v1/engine/revision-summary-pdf.
type EngineRevisionSummaryBody struct {
	Project json.RawMessage `json:"project" swaggertype:"object"`
}

// EngineQuoteExportBody is POST /v1/engine/quote-export. The server picks the
// price book.
type EngineQuoteExportBody struct {
	Project    json.RawMessage `json:"project" swaggertype:"object"`
	Room       json.RawMessage `json:"room" swaggertype:"object"`
	ExportedAt string          `json:"exportedAt,omitempty" example:"2026-10-31T10:00:00Z"`
}

// EngineResultOK is the engine answer, copied through unchanged: a
// LiveInteriorQuote (quote), InteriorEstimateSummary (estimate), or a
// manufacturing result (cut list rows + CSV, machine files, package files),
// or a document (5a): {fileName?, mimeType, bytes, base64} for a PDF, the
// export bundle's text files for the quote export.
type EngineResultOK struct {
	Result json.RawMessage `json:"result" swaggertype:"object"`
}

// PriceBookBody is PUT /v1/org/price-book. book must be a JSON object.
type PriceBookBody struct {
	Book json.RawMessage `json:"book" swaggertype:"object"`
}

// PriceBookOK is GET/PUT /v1/org/price-book. book is null when the org has
// none (the engine then uses its defaults).
type PriceBookOK struct {
	Meta utils.Meta      `json:"meta"`
	Book json.RawMessage `json:"book" swaggertype:"object"`
}
