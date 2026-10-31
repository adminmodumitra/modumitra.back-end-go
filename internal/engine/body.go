package engine

import (
	"bytes"
	"encoding/json"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// priceBookKey is the field the proxy owns. A client-sent value is replaced:
// the server picks the rates, never the browser.
const priceBookKey = "priceBook"

// BuildBody checks the client body is a JSON object with a "document" and
// returns it with "priceBook" set to book (nil = JSON null = engine defaults).
func BuildBody(raw []byte, book json.RawMessage) ([]byte, error) {
	return BuildBodyFor(raw, Spec{Field: "document", NeedsBook: true}, book)
}

// BuildBodyFor checks the client body is a JSON object with spec.Field set.
// Ops that price get "priceBook" set to book (nil = JSON null); the others get
// any client "priceBook" removed, so a browser value never reaches the engine.
func BuildBodyFor(raw []byte, spec Spec, book json.RawMessage) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, invalid("body", "must be a JSON object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, invalid("body", "must be a JSON object")
	}
	value, ok := fields[spec.Field]
	if !ok || len(bytes.TrimSpace(value)) == 0 || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, invalid(spec.Field, "required")
	}
	if !spec.NeedsBook {
		delete(fields, priceBookKey)
		return json.Marshal(fields)
	}
	if len(book) == 0 {
		book = json.RawMessage("null")
	}
	fields[priceBookKey] = book
	return json.Marshal(fields)
}

func invalid(field, reason string) error {
	return utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": field, "reason": reason})
}
