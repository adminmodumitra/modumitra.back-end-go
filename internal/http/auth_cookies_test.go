package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/orglookup"
)

const (
	appOrigin = "https://app.modumitrasoftware.com"
	wwwOrigin = "https://www.modumitrasoftware.com"
)

func sessionRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()
	cfg.Session = config.Session{CookieDomain: ".modumitrasoftware.com", Origins: []string{appOrigin}, GateKey: "gate-test-key"}
	svc, _, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orglookup.Adapter{Service: orgs}, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret})
	return NewRouter(cfg, fakePinger{}, svc, orgs, devs, bills, memorySSO(svc, cfg))
}

func registerAndLogin(t *testing.T, h http.Handler, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email": "cookie@example.com", "password": "correct-horse-battery-staple-19",
		"display_name": "Cookie", "terms_version": "2026-06-01", "privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-cookie"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/login", map[string]any{
		"email": "cookie@example.com", "password": "correct-horse-battery-staple-19",
	}, headers)
	if w.Code != http.StatusOK {
		t.Fatalf("login=%d %s", w.Code, w.Body.String())
	}
	return w
}

func cookieFrom(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s cookie in %v", name, w.Header().Values("Set-Cookie"))
	return nil
}

func TestCookieSessionLogin(t *testing.T) {
	h := sessionRouter(t)
	w := registerAndLogin(t, h, map[string]string{"Origin": appOrigin})
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	csrf, _ := body["csrf_token"].(string)
	if csrf == "" || body["refresh_token"] != "" {
		t.Fatalf("body csrf=%q refresh=%v", csrf, body["refresh_token"])
	}
	access := cookieFrom(t, w, "ms_access")
	if access.Domain != "modumitrasoftware.com" || !access.HttpOnly || !access.Secure || access.SameSite != http.SameSiteLaxMode {
		t.Fatalf("access cookie=%+v", access)
	}
	refresh := cookieFrom(t, w, "ms_refresh")
	if refresh.Domain != "" || refresh.Path != "/v1/auth" || refresh.SameSite != http.SameSiteStrictMode {
		t.Fatalf("refresh cookie=%+v", refresh)
	}
	if gate := cookieFrom(t, w, "ms_gate"); gate.Domain != "modumitrasoftware.com" || !gate.HttpOnly || !strings.HasPrefix(gate.Value, "v1.") {
		t.Fatalf("gate cookie=%+v", gate)
	}
	jar := "ms_access=" + access.Value

	if w := doJSON(t, h, http.MethodGet, "/v1/me", nil, map[string]string{"Cookie": jar}); w.Code != http.StatusOK {
		t.Fatalf("me via cookie=%d %s", w.Code, w.Body.String())
	}
	for name, hdr := range map[string]map[string]string{
		"no csrf":     {"Cookie": jar, "Origin": appOrigin},
		"www origin":  {"Cookie": jar, "Origin": wwwOrigin, "X-CSRF-Token": csrf},
		"wrong token": {"Cookie": jar, "Origin": appOrigin, "X-CSRF-Token": strings.Repeat("0", 64)},
	} {
		if w := doJSON(t, h, http.MethodPost, "/v1/auth/logout-all", nil, hdr); w.Code != http.StatusForbidden {
			t.Fatalf("%s: status=%d", name, w.Code)
		}
	}

	rjar := map[string]string{"Cookie": "ms_refresh=" + refresh.Value}
	rjar["Origin"] = wwwOrigin
	if w := doJSON(t, h, http.MethodPost, "/v1/auth/refresh", nil, rjar); w.Code != http.StatusForbidden {
		t.Fatalf("refresh from www=%d", w.Code)
	}
	rjar["Origin"] = appOrigin
	w = doJSON(t, h, http.MethodPost, "/v1/auth/refresh", nil, rjar)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh via cookie=%d %s", w.Code, w.Body.String())
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	fresh, _ := body["csrf_token"].(string)
	if fresh == "" || fresh == csrf {
		t.Fatalf("refresh must return the rotated session's csrf: %q", fresh)
	}
	jar = "ms_access=" + cookieFrom(t, w, "ms_access").Value
	stale := map[string]string{"Cookie": jar, "Origin": appOrigin, "X-CSRF-Token": csrf}
	if w := doJSON(t, h, http.MethodPost, "/v1/auth/logout-all", nil, stale); w.Code != http.StatusForbidden {
		t.Fatalf("pre-refresh csrf accepted: %d", w.Code)
	}

	ok := map[string]string{"Cookie": jar, "Origin": appOrigin, "X-CSRF-Token": fresh}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/logout-all", nil, ok)
	if w.Code != http.StatusOK {
		t.Fatalf("logout-all with csrf=%d %s", w.Code, w.Body.String())
	}
	if c := cookieFrom(t, w, "ms_access"); c.MaxAge >= 0 {
		t.Fatalf("access cookie not cleared: %+v", c)
	}
}

func TestBearerLoginUnchangedWithSessionsOn(t *testing.T) {
	w := registerAndLogin(t, sessionRouter(t), nil) // desktop/API client: no Origin
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if rt, _ := body["refresh_token"].(string); rt == "" || body["csrf_token"] != nil {
		t.Fatalf("bearer body=%v", body)
	}
	if len(w.Result().Cookies()) != 0 {
		t.Fatalf("cookies set for bearer login: %v", w.Header().Values("Set-Cookie"))
	}
}

func TestCookieRefusedWhenSessionsOff(t *testing.T) {
	h := identRouter(t)
	w := registerAndLogin(t, h, map[string]string{"Origin": appOrigin})
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	access, _ := body["access_token"].(string)
	w = doJSON(t, h, http.MethodGet, "/v1/me", nil, map[string]string{"Cookie": "ms_access=" + access})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("cookie with sessions off=%d", w.Code)
	}
}
