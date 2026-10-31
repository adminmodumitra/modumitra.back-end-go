package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

var proposalBody = map[string]any{
	"document":  map[string]any{"rooms": []any{}},
	"frames":    []any{map[string]any{"cameraId": "c1", "viewName": "Living", "dataUrl": "data:image/jpeg;base64,AAAA"}},
	"now":       "2026-10-31T10:00:00Z",
	"priceBook": map[string]any{"rate": 1}, // replaced by the org book
}

func TestEngineProposalPDFInjectsBookAndOrg(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"fileName":"p.pdf","mimeType":"application/pdf","bytes":4,"base64":"JVBERg=="}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-doc-1")
	orgID := createOrgAs(t, hr.h, access, "org-doc-1")
	doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/proposal-pdf", proposalBody, bearer(access))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"base64":"JVBERg=="`) {
		t.Fatalf("proposal=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if got.path != "/v1/ops/proposal-pdf" || string(got.body["priceBook"]) != `{"carcass":450}` || !strings.Contains(string(got.body["frames"]), `"cameraId":"c1"`) {
		t.Fatalf("path=%s body=%v", got.path, got.body)
	}
	// The engine watermarks with this org id.
	if got.headers.Get("X-Org-Id") != orgID || got.headers.Get("X-Request-Id") == "" {
		t.Fatalf("headers %v", got.headers)
	}
}

func TestEngineRevisionSummaryAndQuoteExport(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-doc-2")
	createOrgAs(t, hr.h, access, "org-doc-2")
	doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/revision-summary-pdf", map[string]any{"project": map[string]any{}, "priceBook": map[string]any{"rate": 1}}, bearer(access))
	if w.Code != http.StatusOK {
		t.Fatalf("summary=%d %s", w.Code, w.Body.String())
	}
	if got := f.seen(); got.path != "/v1/ops/revision-summary-pdf" || got.body["priceBook"] != nil {
		t.Fatalf("summary: path=%s body=%v", got.path, got.body)
	}
	body := map[string]any{"project": map[string]any{}, "room": map[string]any{"dimensions": map[string]any{}}, "exportedAt": "2026-10-31T10:00:00Z"}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote-export", body, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("quote export=%d %s", w.Code, w.Body.String())
	}
	if got := f.seen(); got.path != "/v1/ops/quote-export" || string(got.body["priceBook"]) != `{"carcass":450}` {
		t.Fatalf("quote export: path=%s body=%v", got.path, got.body)
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/proposal-pdf", map[string]any{"project": map[string]any{}}, bearer(access)); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("proposal without document=%d %s", w.Code, w.Body.String())
	}
}

func TestEngineDocumentsTakeLargerBodies(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-doc-3")
	still := "data:image/png;base64," + strings.Repeat("A", 6<<20) // over the 4 MB design limit
	big := map[string]any{"document": map[string]any{}, "now": "2026-10-31T10:00:00Z",
		"frames": []any{map[string]any{"cameraId": "c", "viewName": "v", "dataUrl": still}}}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/proposal-pdf", big, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("6 MB proposal=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", big, bearer(access)); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("6 MB quote=%d, want 413", w.Code)
	}
}

func TestEngineDocumentsHaveTheirOwnQuota(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	start := startEngine(t, f)
	hr := engineRouter(t, func(cfg *config.Config) {
		start(cfg)
		cfg.Engine.DocUserPerMinute = 1
	})
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-doc-quota")
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/revision-summary-pdf", map[string]any{"project": map[string]any{}}, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("first=%d %s", w.Code, w.Body.String())
	}
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/proposal-pdf", proposalBody, bearer(access))
	if w.Code != http.StatusTooManyRequests || errorCode(t, w) != "quota_exceeded" {
		t.Fatalf("over doc quota=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/cutlist", cutlistBody, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("manufacturing must not share the document bucket: %d %s", w.Code, w.Body.String())
	}
}
