package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	obs.InitTest()
	m.Run()
}

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(ctx context.Context) error {
	_, span := otel.Tracer("cabinet.store").Start(ctx, "store.ping")
	defer span.End()
	return f.err
}

func testSettings(cfg config.Config) identity.Settings {
	return identity.Settings{
		TermsVersion:   cfg.TermsVersion,
		PrivacyVersion: cfg.PrivacyVersion,
		AccessTTL:      cfg.AccessTTL,
		RefreshTTL:     cfg.RefreshTTL,
		EmailTokenTTL:  24 * time.Hour,
		JWTSecret:      cfg.JWTSecret,
		Issuer:         identity.Issuer,
	}
}

func testRouter(t *testing.T, p fakePinger) http.Handler {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, _, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, _ := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	return NewRouter(cfg, p, svc, orgs)
}

func TestHealthzOK(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field=%v", body["status"])
	}
	if body["request_id"] == "" || body["trace_id"] == "" {
		t.Fatalf("missing ids: %+v", body)
	}
	if w.Header().Get("X-Request-Id") == "" || w.Header().Get("X-Trace-Id") == "" {
		t.Fatalf("missing headers: %v", w.Header())
	}
	if w.Header().Get("traceparent") == "" {
		t.Fatalf("missing traceparent")
	}
}

func TestReadyzOK(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ready" {
		t.Fatalf("status field=%v", body["status"])
	}
	checks, _ := body["checks"].(map[string]any)
	if checks["store"] != "ok" && checks["postgres"] != "ok" {
		t.Fatalf("checks=%v want store=ok", body["checks"])
	}
}

func TestReadyzUnavailable(t *testing.T) {
	r := testRouter(t, fakePinger{err: context.DeadlineExceeded})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "service_unavailable" {
		t.Fatalf("error=%v", body)
	}
	if errObj["trace_id"] == "" || errObj["request_id"] == "" {
		t.Fatalf("error ids=%v", errObj)
	}
}

func TestPublicConfig(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/v1/public/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["trial_days"] != float64(14) {
		t.Fatalf("trial_days=%v", body["trial_days"])
	}
	provs, _ := body["oauth_providers"].([]any)
	if len(provs) != 3 {
		t.Fatalf("oauth_providers=%v", body["oauth_providers"])
	}
}

func TestJWTUnauthenticated(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "unauthenticated" {
		t.Fatalf("error=%v", body)
	}
	if w.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("WWW-Authenticate=%q", w.Header().Get("WWW-Authenticate"))
	}
}

func TestPublicRouteWithoutAuthorization(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/v1/public/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate=%q; JWT must not run on public routes", got)
	}
}

func TestNoRouteNotJWT(t *testing.T) {
	r := testRouter(t, fakePinger{})
	req := httptest.NewRequest(http.MethodGet, "/v1/does-not-exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "not_found" {
		t.Fatalf("error=%v", body)
	}
	if got := w.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate=%q; NoRoute must not run JWT", got)
	}
}

func TestTraceparentEcho(t *testing.T) {
	r := testRouter(t, fakePinger{})
	const wantTrace = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	parent := "00-" + wantTrace + "-bbbbbbbbbbbbbbbb-01"
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("traceparent", parent)
	req.Header.Set("X-Request-Id", "req_client_trace_echo")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("X-Trace-Id"); got != wantTrace {
		t.Fatalf("X-Trace-Id=%q want %q headers=%v body=%s", got, wantTrace, w.Header(), w.Body.String())
	}
	if got := w.Header().Get("X-Request-Id"); got != "req_client_trace_echo" {
		t.Fatalf("X-Request-Id=%q", got)
	}
	tp := w.Header().Get("traceparent")
	if tp == "" {
		t.Fatalf("missing response traceparent; headers=%v", w.Header())
	}
	if !strings.HasPrefix(tp, "00-"+wantTrace+"-") {
		t.Fatalf("traceparent=%q did not continue trace %s", tp, wantTrace)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["trace_id"] != wantTrace {
		t.Fatalf("body.trace_id=%v", body["trace_id"])
	}
}

func TestReadyzReportsActiveDriver(t *testing.T) {
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, _, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, _ := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	r := NewRouter(cfg, fakePinger{}, svc, orgs)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	checks, _ := body["checks"].(map[string]any)
	if checks["store"] != "ok" {
		t.Fatalf("checks=%v want store=ok (fake pinger)", body["checks"])
	}
}
