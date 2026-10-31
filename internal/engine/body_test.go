package engine

import (
	"encoding/json"
	"testing"
)

func TestBuildBodyInjectsPriceBook(t *testing.T) {
	out, err := BuildBody([]byte(`{"document":{"rooms":[]},"now":"2026-10-31T00:00:00Z"}`), json.RawMessage(`{"rate":450}`))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if string(got["priceBook"]) != `{"rate":450}` || string(got["document"]) != `{"rooms":[]}` || string(got["now"]) != `"2026-10-31T00:00:00Z"` {
		t.Fatalf("%s", out)
	}
}

func TestBuildBodyNullBookAndClientBookReplaced(t *testing.T) {
	out, err := BuildBody([]byte(`{"document":{},"priceBook":{"rate":1}}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if string(got["priceBook"]) != "null" {
		t.Fatalf("client rates leaked through: %s", out)
	}
}

func TestBuildBodyRejects(t *testing.T) {
	for _, raw := range []string{``, `[]`, `null`, `{"now":"x"}`, `{"document":null}`, `{"document":`} {
		if _, err := BuildBody([]byte(raw), nil); err == nil {
			t.Errorf("%q: want error", raw)
		}
	}
}
