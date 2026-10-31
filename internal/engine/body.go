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
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, invalid("body", "must be a JSON object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, invalid("body", "must be a JSON object")
	}
	doc, ok := fields["document"]
	if !ok || len(bytes.TrimSpace(doc)) == 0 || bytes.Equal(bytes.TrimSpace(doc), []byte("null")) {
		return nil, invalid("document", "required")
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
