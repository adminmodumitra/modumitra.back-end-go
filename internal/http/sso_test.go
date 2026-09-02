package httpapi

import (
	"context"
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
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

type ssoHarness struct {
	h     http.Handler
	ident *identity.Service
	sso   *sso.Service
	fakes map[string]*sso.FakeProvider
	mail  *identity.RecordingMailer
	cfg   config.Config
}

func newSSOHarness(t *testing.T) *ssoHarness {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	ident, mail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orgs, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret})
	ssoSvc, _, fakes := sso.NewMemoryService(ident, sso.Settings{
		TermsVersion:    cfg.TermsVersion,
		PrivacyVersion:  cfg.PrivacyVersion,
		PKCESecret:      cfg.JWTSecret,
		VerificationURI: "https://app.cabinetstudio.example/device",
	})
	h := NewRouter(cfg, fakePinger{}, ident, orgs, devs, bills, ssoSvc)
	return &ssoHarness{h: h, ident: ident, sso: ssoSvc, fakes: fakes, mail: mail, cfg: cfg}
}

func TestOAuthProvidersListsConfigured(t *testing.T) {
	hr := newSSOHarness(t)
	w := doJSON(t, hr.h, http.MethodGet, "/v1/auth/oauth/providers", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	items, _ := body["providers"].([]any)
	if len(items) != 3 {
		t.Fatalf("providers=%v", body["providers"])
	}
	ids := map[string]bool{}
	for _, raw := range items {
		p, _ := raw.(map[string]any)
		ids[p["id"].(string)] = true
		if p["id"] == "google" && p["device_code_supported"] != true {
			t.Fatalf("google device_code_supported=%v", p["device_code_supported"])
		}
	}
	for _, want := range []string{"google", "facebook", "x"} {
		if !ids[want] {
			t.Fatalf("missing %s in %v", want, ids)
		}
	}
}

func TestOAuthStartLoginReturnsState(t *testing.T) {
	hr := newSSOHarness(t)
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/google/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "login",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	state, _ := body["state"].(string)
	urlStr, _ := body["authorization_url"].(string)
	if !strings.HasPrefix(state, "st_") {
		t.Fatalf("state=%q", state)
	}
	if !strings.Contains(urlStr, state) {
		t.Fatalf("authorization_url=%q missing state", urlStr)
	}
}

func TestOAuthUnknownProvider422(t *testing.T) {
	hr := newSSOHarness(t)
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/yahoo/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "login",
	}, nil)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "validation_error" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestOAuthCallbackProvisionsUser(t *testing.T) {
	hr := newSSOHarness(t)
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/google/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "login",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("start=%d %s", w.Code, w.Body.String())
	}
	var start map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	state, _ := start["state"].(string)
	hr.fakes["google"].Set("code-ada", sso.Profile{
		ProviderUserID: "gid_ada",
		Email:          "ada@example.com",
		EmailVerified:  true,
		DisplayName:    "Ada Lovelace",
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/oauth/google/callback?code=code-ada&state="+state, nil)
	w = httptest.NewRecorder()
	hr.h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("callback=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	user, _ := body["user"].(map[string]any)
	if user["email"] != "ada@example.com" {
		t.Fatalf("user=%v", user)
	}
	if user["email_verified"] != true {
		t.Fatalf("email_verified=%v", user["email_verified"])
	}
	if body["access_token"] == "" || body["refresh_token"] == "" {
		t.Fatalf("tokens=%v", body)
	}
	if body["linked_provider"] != "google" {
		t.Fatalf("linked_provider=%v", body["linked_provider"])
	}
	if _, has := user["password_hash"]; has {
		t.Fatal("password_hash leaked")
	}
}

func TestOAuthSecondLoginSameSubject(t *testing.T) {
	hr := newSSOHarness(t)
	id1 := oauthLogin(t, hr, "google", "code-1", sso.Profile{ProviderUserID: "gid_ada", Email: "ada@example.com", EmailVerified: true, DisplayName: "Ada"})
	id2 := oauthLogin(t, hr, "google", "code-2", sso.Profile{ProviderUserID: "gid_ada", Email: "ada@example.com", EmailVerified: true, DisplayName: "Ada"})
	if id1 != id2 {
		t.Fatalf("duplicated user %s vs %s", id1, id2)
	}
}

func TestOAuthLinkSecondProvider(t *testing.T) {
	hr := newSSOHarness(t)
	access, userID := oauthLoginAccess(t, hr, "google", "code-g", sso.Profile{ProviderUserID: "gid_ada", Email: "ada@example.com", EmailVerified: true, DisplayName: "Ada"})
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/facebook/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "link",
	}, map[string]string{"Authorization": "Bearer " + access})
	if w.Code != http.StatusOK {
		t.Fatalf("start link=%d %s", w.Code, w.Body.String())
	}
	var start map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	state, _ := start["state"].(string)
	hr.fakes["facebook"].Set("code-fb", sso.Profile{ProviderUserID: "fb_ada", Email: "ada@example.com", DisplayName: "Ada"})
	w = doJSON(t, hr.h, http.MethodPost, "/v1/me/identities/facebook/link", map[string]any{
		"authorization_code": "code-fb",
		"state":              state,
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "link-fb-1",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("link=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, hr.h, http.MethodGet, "/v1/me/identities", nil, map[string]string{"Authorization": "Bearer " + access})
	if w.Code != http.StatusOK {
		t.Fatalf("list=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	items, _ := body["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items=%v user=%s", body["items"], userID)
	}
}

func TestOAuthUnlinkLastMethod(t *testing.T) {
	hr := newSSOHarness(t)
	access, _ := oauthLoginAccess(t, hr, "google", "code-g", sso.Profile{ProviderUserID: "gid_ada", Email: "ada@example.com", EmailVerified: true, DisplayName: "Ada"})
	w := doJSON(t, hr.h, http.MethodDelete, "/v1/me/identities/google", nil, map[string]string{"Authorization": "Bearer " + access})
	if w.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "last_login_method" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestOAuthIdentityAlreadyLinked(t *testing.T) {
	hr := newSSOHarness(t)
	oauthLogin(t, hr, "google", "code-a", sso.Profile{ProviderUserID: "gid_shared", Email: "ada@example.com", EmailVerified: true, DisplayName: "Ada"})
	accessB, _ := registerVerifyLogin(t, hr.h, hr.mail, "bob@example.com", "reg-bob-sso")
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/google/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "link",
	}, map[string]string{"Authorization": "Bearer " + accessB})
	if w.Code != http.StatusOK {
		t.Fatalf("start=%d %s", w.Code, w.Body.String())
	}
	var start map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	state, _ := start["state"].(string)
	hr.fakes["google"].Set("code-b", sso.Profile{ProviderUserID: "gid_shared", Email: "ada@example.com", DisplayName: "Ada"})
	w = doJSON(t, hr.h, http.MethodPost, "/v1/me/identities/google/link", map[string]any{
		"authorization_code": "code-b",
		"state":              state,
	}, map[string]string{
		"Authorization":   "Bearer " + accessB,
		"Idempotency-Key": "link-stolen",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "identity_already_linked" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestDeviceCodePendingThenTokens(t *testing.T) {
	hr := newSSOHarness(t)
	_, userID := oauthLoginAccess(t, hr, "google", "code-dev", sso.Profile{ProviderUserID: "gid_dev", Email: "dev@example.com", EmailVerified: true, DisplayName: "Dev"})
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/code", map[string]any{
		"client_id": "cabinet-studio-desktop",
		"provider":  "google",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
	var issued map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}
	deviceCode, _ := issued["device_code"].(string)
	userCode, _ := issued["user_code"].(string)
	if !strings.HasPrefix(deviceCode, "dc_") || !strings.Contains(userCode, "-") {
		t.Fatalf("issued=%v", issued)
	}
	if issued["expires_in"] != float64(600) || issued["interval"] != float64(5) {
		t.Fatalf("ttl=%v", issued)
	}
	w = doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/token", map[string]any{"device_code": deviceCode}, nil)
	if w.Code != 428 {
		t.Fatalf("pending status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "authorization_pending" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
	if err := hr.sso.AuthorizeDevice(context.Background(), userCode, userID); err != nil {
		t.Fatalf("authorize: %v", err)
	}
	w = doJSON(t, hr.h, http.MethodPost, "/v1/auth/device/token", map[string]any{"device_code": deviceCode}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("token=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	if login["access_token"] == "" || login["refresh_token"] == "" {
		t.Fatalf("login=%v", login)
	}
	user, _ := login["user"].(map[string]any)
	if user["id"] != userID {
		t.Fatalf("user=%v want %s", user, userID)
	}
}

func oauthLogin(t *testing.T, hr *ssoHarness, provider, code string, p sso.Profile) string {
	t.Helper()
	_, id := oauthLoginAccess(t, hr, provider, code, p)
	return id
}

func oauthLoginAccess(t *testing.T, hr *ssoHarness, provider, code string, p sso.Profile) (access, userID string) {
	t.Helper()
	w := doJSON(t, hr.h, http.MethodPost, "/v1/auth/oauth/"+provider+"/start", map[string]any{
		"redirect_uri": "cabinetstudio://oauth/callback",
		"intent":       "login",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("start=%d %s", w.Code, w.Body.String())
	}
	var start map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	state, _ := start["state"].(string)
	hr.fakes[provider].Set(code, p)
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/oauth/"+provider+"/callback?code="+code+"&state="+state, nil)
	w = httptest.NewRecorder()
	hr.h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("callback=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	access, _ = body["access_token"].(string)
	user, _ := body["user"].(map[string]any)
	userID, _ = user["id"].(string)
	if access == "" || userID == "" {
		t.Fatalf("login body=%v", body)
	}
	return access, userID
}
