package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
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
	_, span := otel.Tracer("cabinet.store").Start(ctx, "postgres.ping")
	defer span.End()
	return f.err
}

func testRouter(t *testing.T, p fakePinger) http.Handler {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	return NewRouter(cfg, p)
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
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", nil)
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
