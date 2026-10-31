package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

var quoteBody = map[string]any{
	"document":  map[string]any{"rooms": []any{}},
	"now":       "2026-10-31T10:00:00Z",
	"priceBook": map[string]any{"rate": 1}, // must never reach the engine
}

func TestEngineQuoteForwardsOrgPriceBook(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"total":123}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, userID := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-1")
	orgID := createOrgAs(t, hr.h, access, "org-eng-1")
	w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	if w.Code != http.StatusOK {
		t.Fatalf("put book=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
	if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != `{"result":{"total":123}}` {
		t.Fatalf("quote=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if got.path != "/v1/ops/quote" {
		t.Fatalf("path=%s", got.path)
	}
	if got.headers.Get("X-Engine-Key") != testEngineKey || got.headers.Get("X-Org-Id") != orgID ||
		got.headers.Get("X-User-Id") != userID || got.headers.Get("X-Request-Id") == "" {
		t.Fatalf("headers=%v", got.headers)
	}
	if string(got.body["priceBook"]) != `{"carcass":450}` || string(got.body["now"]) != `"2026-10-31T10:00:00Z"` {
		t.Fatalf("body=%v", got.body)
	}
}

func TestEngineEstimatePersonalGetsNullBook(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"low":1,"high":2}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-2")
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/estimate", map[string]any{"document": map[string]any{}}, bearer(access))
	if w.Code != http.StatusOK {
		t.Fatalf("estimate=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if got.path != "/v1/ops/estimate" || string(got.body["priceBook"]) != "null" || got.headers.Get("X-Org-Id") != "" {
		t.Fatalf("path=%s body=%v org=%q", got.path, got.body, got.headers.Get("X-Org-Id"))
	}
}

func TestEngineUnavailableWithoutURL(t *testing.T) {
	hr := engineRouter(t, func(cfg *config.Config) { cfg.Engine.URL = "" })
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-3")
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
	if w.Code != http.StatusServiceUnavailable || errorCode(t, w) != "engine_unavailable" {
		t.Fatalf("quote=%d %s", w.Code, w.Body.String())
	}
}

func TestEngineSignedOut401(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	hr := engineRouter(t, startEngine(t, f))
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, nil)
	if w.Code != http.StatusUnauthorized || f.seen().calls != 0 {
		t.Fatalf("quote=%d calls=%d", w.Code, f.seen().calls)
	}
}

func TestEngineErrorsMapToGateway(t *testing.T) {
	cases := []struct {
		status int
		reply  string
		want   int
		code   string
	}{
		{http.StatusInternalServerError, `{"error":{"code":"boom","message":"stack trace"}}`, http.StatusBadGateway, "engine_error"},
		{http.StatusUnauthorized, `{"error":{"code":"bad_key","message":"no"}}`, http.StatusBadGateway, "engine_error"},
		{http.StatusUnprocessableEntity, `{"error":{"code":"invalid_document","message":"rooms missing"}}`, http.StatusUnprocessableEntity, "invalid_document"},
	}
	for _, tc := range cases {
		f := &fakeEngine{status: tc.status, reply: tc.reply}
		hr := engineRouter(t, startEngine(t, f))
		access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-err")
		w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
		if w.Code != tc.want || errorCode(t, w) != tc.code || strings.Contains(w.Body.String(), "stack trace") {
			t.Fatalf("engine %d → %d %s", tc.status, w.Code, w.Body.String())
		}
	}
}

func TestEngineDown502(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close() // connection refused
	hr := engineRouter(t, func(cfg *config.Config) { cfg.Engine.URL = url })
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-down")
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
	if w.Code != http.StatusBadGateway || errorCode(t, w) != "engine_error" {
		t.Fatalf("quote=%d %s", w.Code, w.Body.String())
	}
}

func TestEngineTimeout504(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(400 * time.Millisecond)
		_, _ = w.Write([]byte(`{"result":{}}`))
	}))
	t.Cleanup(slow.Close)
	hr := engineRouter(t, func(cfg *config.Config) {
		cfg.Engine.URL = slow.URL
		cfg.Engine.Timeout = 50 * time.Millisecond
	})
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-slow")
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
	if w.Code != http.StatusGatewayTimeout || errorCode(t, w) != "engine_timeout" {
		t.Fatalf("quote=%d %s", w.Code, w.Body.String())
	}
}

func TestEngineBodyLimits(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-big")
	// 2 MB passes (the general 1 MB limit does not apply here).
	mid := `{"document":{"blob":"` + strings.Repeat("a", 2<<20) + `"}}`
	if w := rawPost(hr, access, mid); w.Code != http.StatusOK {
		t.Fatalf("2MB=%d %s", w.Code, w.Body.String())
	}
	big := `{"document":{"blob":"` + strings.Repeat("a", 4<<20) + `"}}`
	if w := rawPost(hr, access, big); w.Code != http.StatusRequestEntityTooLarge || errorCode(t, w) != "payload_too_large" {
		t.Fatalf("5MB=%d %s", w.Code, w.Body.String())
	}
	if w := rawPost(hr, access, `{"now":"x"}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("no document=%d %s", w.Code, w.Body.String())
	}
}

func rawPost(hr *engineHarness, access, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/engine/quote", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access)
	w := httptest.NewRecorder()
	hr.h.ServeHTTP(w, req)
	return w
}
