package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCSRFTokenIsPerSession(t *testing.T) {
	key := CSRFKey("a-secret")
	if CSRFToken(key, "ses_1") != CSRFToken(key, "ses_1") {
		t.Fatal("token not stable")
	}
	if CSRFToken(key, "ses_1") == CSRFToken(key, "ses_2") {
		t.Fatal("token not bound to session")
	}
	if CSRFToken(CSRFKey("other"), "ses_1") == CSRFToken(key, "ses_1") {
		t.Fatal("token not bound to secret")
	}
}

func TestCORSCredentialsOnlyForSessionOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := "https://app.modumitrasoftware.com"
	www := "https://www.modumitrasoftware.com"
	r := gin.New()
	r.Use(CORS([]string{www}, app))
	r.POST("/v1/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := corsHit(r, http.MethodOptions, app, true)
	if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("app preflight=%d creds=%q", w.Code, w.Header().Get("Access-Control-Allow-Credentials"))
	}
	w = corsHit(r, http.MethodPost, www, false)
	if w.Header().Get("Access-Control-Allow-Origin") != www || w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("www must not get credentials: %v", w.Header())
	}
}
