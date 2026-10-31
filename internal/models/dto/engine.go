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

// EngineResultOK is the engine answer, copied through unchanged: a
// LiveInteriorQuote (quote) or InteriorEstimateSummary (estimate).
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
