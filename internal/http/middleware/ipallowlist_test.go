package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIPAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("empty allows every ip", func(t *testing.T) {
		r := allowlistRouter(t, nil)
		if code := hit(r, "/v1/public/config", "203.0.113.9:1234"); code != http.StatusOK {
			t.Fatalf("status=%d", code)
		}
	})

	t.Run("exact and cidr allowed", func(t *testing.T) {
		r := allowlistRouter(t, []string{"10.0.0.8", "192.168.1.0/24", "2001:db8::/32"})
		if code := hit(r, "/v1/public/config", "10.0.0.8:443"); code != http.StatusOK {
			t.Fatalf("exact status=%d", code)
		}
		if code := hit(r, "/v1/webhooks/billing", "192.168.1.40:9"); code != http.StatusOK {
			t.Fatalf("cidr status=%d", code)
		}
		if code := hit(r, "/v1/me", "[2001:db8::1]:443"); code != http.StatusOK {
			t.Fatalf("ipv6 status=%d", code)
		}
	})

	t.Run("mismatch is 401", func(t *testing.T) {
		r := allowlistRouter(t, []string{"10.0.0.8"})
		if code := hit(r, "/v1/me", "203.0.113.9:1234"); code != http.StatusUnauthorized {
			t.Fatalf("status=%d", code)
		}
	})

	t.Run("render uses cloudflare client ip", func(t *testing.T) {
		r := allowlistRouter(t, []string{"106.219.151.223"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		req.RemoteAddr = "10.21.157.68:443"
		req.Header.Set("CF-Connecting-IP", "106.219.151.223")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("public peer cannot spoof cloudflare header", func(t *testing.T) {
		r := allowlistRouter(t, []string{"10.0.0.8"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		req.RemoteAddr = "203.0.113.9:1234"
		req.Header.Set("CF-Connecting-IP", "10.0.0.8")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("probes and swagger stay open", func(t *testing.T) {
		r := allowlistRouter(t, []string{"10.0.0.8"})
		if code := hit(r, "/healthz", "203.0.113.9:1234"); code != http.StatusOK {
			t.Fatalf("healthz status=%d", code)
		}
		if code := hit(r, "/readyz", "203.0.113.9:1234"); code != http.StatusOK {
			t.Fatalf("readyz status=%d", code)
		}
		if code := hit(r, "/swagger/index.html", "203.0.113.9:1234"); code != http.StatusOK {
			t.Fatalf("swagger status=%d", code)
		}
	})

	t.Run("login is open to every ip, other auth routes are not", func(t *testing.T) {
		r := allowlistRouter(t, []string{"10.0.0.8"})
		if code := hit(r, "/v1/auth/login", "203.0.113.9:1234"); code != http.StatusOK {
			t.Fatalf("login status=%d", code)
		}
		if code := hit(r, "/v1/auth/register", "203.0.113.9:1234"); code != http.StatusUnauthorized {
			t.Fatalf("register status=%d", code)
		}
		if code := hit(r, "/v1/auth/login/extra", "203.0.113.9:1234"); code != http.StatusUnauthorized {
			t.Fatalf("login prefix status=%d", code)
		}
	})
}

func allowlistRouter(t *testing.T, allowed []string) *gin.Engine {
	t.Helper()
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	r.Use(IPAllowlist(allowed))
	r.Any("/*any", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func hit(r *gin.Engine, path, remote string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = remote
	r.ServeHTTP(w, req)
	return w.Code
}
