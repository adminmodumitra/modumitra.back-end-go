package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := obs.Logger()
	obs.SetLogger(zerolog.New(buf))
	t.Cleanup(func() { obs.SetLogger(prev) })
	return buf
}

func loginContext(t *testing.T, w *httptest.ResponseRecorder, password string) *gin.Context {
	t.Helper()
	c, _ := gin.CreateTestContext(w)
	body := `{"email":"ada@example.com","password":"` + password + `"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("request_id", "req_httplog_test")
	c.Set(utils.CtxUserID, "usr_ada")
	return c
}

func parseLogLines(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log json: %v line=%s", err, line)
		}
		out = append(out, m)
	}
	return out
}

func lineByPhase(t *testing.T, lines []map[string]any, phase string) map[string]any {
	t.Helper()
	for _, m := range lines {
		if m["phase"] == phase {
			return m
		}
	}
	t.Fatalf("missing phase=%s in %#v", phase, lines)
	return nil
}

func TestLogEnterWriteJSONSuccess(t *testing.T) {
	buf := captureLogs(t)
	password := "super-secret-password"
	w := httptest.NewRecorder()
	c := loginContext(t, w, password)

	logEnter(c, "login")
	middleware.WriteJSON(c, utils.StatusOK, gin.H{"ok": true})

	raw := buf.String()
	if strings.Contains(raw, password) {
		t.Fatalf("password leaked into logs: %s", raw)
	}
	lines := parseLogLines(t, raw)
	enter := lineByPhase(t, lines, "enter")
	success := lineByPhase(t, lines, "success")
	if enter["op"] != "login" {
		t.Fatalf("enter op=%v", enter["op"])
	}
	if enter["user_id"] != "usr_ada" {
		t.Fatalf("enter user_id=%v", enter["user_id"])
	}
	if enter["request_id"] != "req_httplog_test" {
		t.Fatalf("enter request_id=%v", enter["request_id"])
	}
	if success["op"] != "login" || success["user_id"] != "usr_ada" || success["request_id"] != "req_httplog_test" {
		t.Fatalf("success=%v", success)
	}
	if success["status"] != float64(200) {
		t.Fatalf("success status=%v", success["status"])
	}
	if _, ok := success["error_code"]; ok {
		t.Fatalf("success should not have error_code: %v", success)
	}
	if strings.Contains(raw, `"stack"`) || strings.Contains(raw, "goroutine ") {
		t.Fatalf("stack in success/enter logs: %s", raw)
	}
	for _, m := range lines {
		for k, v := range m {
			if s, ok := v.(string); ok && s == "-" {
				t.Fatalf("empty placeholder dash for %s", k)
			}
		}
	}
}

func TestLogEnterAbortExit(t *testing.T) {
	buf := captureLogs(t)
	password := "super-secret-password"
	w := httptest.NewRecorder()
	c := loginContext(t, w, password)

	logEnter(c, "login")
	middleware.Abort(c, utils.E(utils.CodeInvalidCredentials))

	raw := buf.String()
	if strings.Contains(raw, password) {
		t.Fatalf("password leaked into logs: %s", raw)
	}
	if strings.Contains(raw, `"stack"`) || strings.Contains(raw, "goroutine ") {
		t.Fatalf("stack dumped on abort: %s", raw)
	}
	if strings.Contains(raw, "Invalid email or password") {
		t.Fatalf("error string leaked: %s", raw)
	}
	lines := parseLogLines(t, raw)
	exit := lineByPhase(t, lines, "exit")
	if exit["op"] != "login" {
		t.Fatalf("exit op=%v", exit["op"])
	}
	if exit["error_code"] != "invalid_credentials" {
		t.Fatalf("error_code=%v", exit["error_code"])
	}
	if exit["status"] != float64(401) {
		t.Fatalf("status=%v", exit["status"])
	}
	if exit["user_id"] != "usr_ada" || exit["request_id"] != "req_httplog_test" {
		t.Fatalf("exit ids=%v", exit)
	}
	if _, ok := lineByPhaseMaybe(lines, "success"); ok {
		t.Fatalf("unexpected success line: %s", raw)
	}
}

func lineByPhaseMaybe(lines []map[string]any, phase string) (map[string]any, bool) {
	for _, m := range lines {
		if m["phase"] == phase {
			return m, true
		}
	}
	return nil, false
}
