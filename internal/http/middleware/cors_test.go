package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func corsRouter(origins []string) *gin.Engine {
	r := gin.New()
	r.Use(CORS(origins))
	r.POST("/v1/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })
	return r
}

func corsHit(r *gin.Engine, method, origin string, preflight bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/v1/auth/login", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if preflight {
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "content-type")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	site := "https://modumitrasoftware.com"

	t.Run("preflight from allowed origin is 204 with headers", func(t *testing.T) {
		w := corsHit(corsRouter([]string{site + "/"}), http.MethodOptions, site, true)
		if w.Code != http.StatusNoContent {
			t.Fatalf("status=%d", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != site {
			t.Fatalf("allow-origin=%q", got)
		}
		if w.Header().Get("Access-Control-Allow-Methods") == "" || w.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Fatalf("missing preflight headers: %v", w.Header())
		}
		if w.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Fatal("credentials must stay off")
		}
	})

	t.Run("actual request from allowed origin gets allow-origin", func(t *testing.T) {
		w := corsHit(corsRouter([]string{site}), http.MethodPost, site, false)
		if w.Code != http.StatusOK || w.Header().Get("Access-Control-Allow-Origin") != site {
			t.Fatalf("status=%d allow-origin=%q", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
		}
		if w.Header().Get("Vary") != "Origin" {
			t.Fatalf("vary=%q", w.Header().Get("Vary"))
		}
	})

	t.Run("other origin gets no allow-origin", func(t *testing.T) {
		w := corsHit(corsRouter([]string{site}), http.MethodOptions, "https://evil.example", true)
		if w.Header().Get("Access-Control-Allow-Origin") != "" || w.Code == http.StatusNoContent {
			t.Fatalf("status=%d headers=%v", w.Code, w.Header())
		}
	})

	t.Run("wildcard allows any origin", func(t *testing.T) {
		w := corsHit(corsRouter([]string{"*"}), http.MethodOptions, "http://localhost:1420", true)
		if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:1420" {
			t.Fatalf("status=%d headers=%v", w.Code, w.Header())
		}
	})

	t.Run("empty list is off", func(t *testing.T) {
		w := corsHit(corsRouter(nil), http.MethodOptions, site, true)
		if w.Header().Get("Access-Control-Allow-Origin") != "" || w.Code != http.StatusNotFound {
			t.Fatalf("status=%d headers=%v", w.Code, w.Header())
		}
	})

	t.Run("no origin header passes through", func(t *testing.T) {
		w := corsHit(corsRouter([]string{site}), http.MethodPost, "", false)
		if w.Code != http.StatusOK || w.Header().Get("Vary") != "" {
			t.Fatalf("status=%d headers=%v", w.Code, w.Header())
		}
	})
}
