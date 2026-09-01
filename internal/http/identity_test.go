package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

func identRouter(t *testing.T) http.Handler {
	t.Helper()
	h, _ := identRouterWithMailer(t)
	return h
}

func identRouterWithMailer(t *testing.T) (http.Handler, *identity.RecordingMailer) {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, mail, _ := identity.NewMemoryService(testSettings(cfg))
	return NewRouter(cfg, fakePinger{}, svc), mail
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, w.Body.String())
	}
	errObj, _ := body["error"].(map[string]any)
	code, _ := errObj["code"].(string)
	return code
}

func TestRegister201(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-1"})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	user, _ := body["user"].(map[string]any)
	if user["email"] != "ada@example.com" {
		t.Fatalf("user=%v", user)
	}
	if user["email_verified"] != false {
		t.Fatalf("verified=%v", user["email_verified"])
	}
	if _, has := user["password_hash"]; has {
		t.Fatal("password_hash leaked")
	}
}

func TestRegisterMissingIdempotencyKey(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, nil)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "validation_error" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestLogin401Unknown(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "nobody@example.com",
		"password": "correct-horse-battery-staple-19",
	}, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "invalid_credentials" {
		t.Fatalf("code=%s", errorCode(t, w))
	}
}

func TestLoginThenMe(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-me"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "ada@example.com",
		"password": "correct-horse-battery-staple-19",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	if login["organization"] != nil {
		t.Fatalf("organization=%v", login["organization"])
	}
	if login["expires_in"] != float64(900) || login["token_type"] != "Bearer" {
		t.Fatalf("tokens meta=%v", login)
	}
	access, _ := login["access_token"].(string)
	if access == "" {
		t.Fatal("missing access")
	}
	w = doJSON(t, h, http.MethodGet, "/v1/me", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("me=%d %s", w.Code, w.Body.String())
	}
	var me map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	user, _ := me["user"].(map[string]any)
	if user["email"] != "ada@example.com" {
		t.Fatalf("me user=%v", user)
	}
}

func TestRefreshAndReuse(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-rt"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "ada@example.com",
		"password": "correct-horse-battery-staple-19",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	rt, _ := login["refresh_token"].(string)
	w = doJSON(t, h, http.MethodPost, "/v1/auth/refresh", map[string]any{
		"refresh_token": rt,
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh=%d %s", w.Code, w.Body.String())
	}
	var rotated map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &rotated); err != nil {
		t.Fatal(err)
	}
	if rotated["refresh_token"] == rt {
		t.Fatal("refresh did not rotate")
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/refresh", map[string]any{
		"refresh_token": rt,
	}, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("reuse status=%d %s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "token_reuse_detected" {
		t.Fatalf("reuse code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestOTPRequestAccepted(t *testing.T) {
	h := identRouter(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/otp/request", map[string]any{
		"email": "ada@example.com",
	}, nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "accepted" {
		t.Fatalf("body=%v", body)
	}
}

func TestOTPRegisterRequestVerify(t *testing.T) {
	h, mail := identRouterWithMailer(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-otp"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/otp/request", map[string]any{
		"email": "ada@example.com",
	}, nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("otp request=%d %s", w.Code, w.Body.String())
	}
	call, ok := mail.LastOTP()
	if !ok || len(call.Token) != 6 {
		t.Fatalf("otp mail=%+v ok=%v", call, ok)
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/otp/verify", map[string]any{
		"email": "ada@example.com",
		"otp":   call.Token,
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("verify=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	access, _ := login["access_token"].(string)
	if access == "" {
		t.Fatalf("missing access_token body=%v", login)
	}
	if login["expires_in"] != float64(900) || login["token_type"] != "Bearer" {
		t.Fatalf("tokens meta=%v", login)
	}
	if login["organization"] != nil {
		t.Fatalf("organization=%v", login["organization"])
	}
}

func TestOTPVerifyBadCode(t *testing.T) {
	h, mail := identRouterWithMailer(t)
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           "ada@example.com",
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": "reg-otp-bad"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/otp/request", map[string]any{
		"email": "ada@example.com",
	}, nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("otp request=%d %s", w.Code, w.Body.String())
	}
	call, _ := mail.LastOTP()
	bad := "000000"
	if call.Token == bad {
		bad = "000001"
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/otp/verify", map[string]any{
		"email": "ada@example.com",
		"otp":   bad,
	}, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "invalid_credentials" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}
