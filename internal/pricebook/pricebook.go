// Package pricebook stores one price book (rates JSON) per organization.
// The engine proxy injects it into quote/estimate calls, so rates live in the
// database and never ship in the app bundle.
package pricebook

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// MaxBookBytes caps PUT /v1/org/price-book bodies.
const MaxBookBytes = 1 << 20

type Book struct {
	OrgID     string
	Book      json.RawMessage // always a JSON object
	UpdatedAt time.Time
	UpdatedBy string
}

// Repo is the storage port. Get returns (nil, nil) when the org has no book,
// which means "engine defaults".
type Repo interface {
	Get(ctx context.Context, orgID string) (*Book, error)
	Put(ctx context.Context, b Book) error
}

// ValidateBook accepts only a JSON object. Arrays, scalars and null are a
// validation error; the engine expects an object of rates.
func ValidateBook(raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return utils.E(utils.CodeValidation).WithDetails(map[string]any{
			"field":  "book",
			"reason": "must be a JSON object",
		})
	}
	return nil
}
