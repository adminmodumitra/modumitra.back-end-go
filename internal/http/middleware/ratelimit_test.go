package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func limitedRouter(t *testing.T, mw ...gin.HandlerFunc) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	r.Use(mw...)
	r.Any("/*any", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func limitedHit(r *gin.Engine, path, remote, cfIP string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = remote
	if cfIP != "" {
		req.Header.Set("CF-Connecting-IP", cfIP)
	}
	r.ServeHTTP(w, req)
	return w.Code
}

func TestIPRateLimit(t *testing.T) {
	t.Run("blocks after burst", func(t *testing.T) {
		r := limitedRouter(t, IPRateLimit(rate.Every(time.Hour), 2))
		for i := 0; i < 2; i++ {
			if code := limitedHit(r, "/v1/auth/login", "203.0.113.9:1", ""); code != http.StatusOK {
				t.Fatalf("hit %d status=%d", i, code)
			}
		}
		if code := limitedHit(r, "/v1/auth/login", "203.0.113.9:1", ""); code != http.StatusTooManyRequests {
			t.Fatalf("over burst status=%d", code)
		}
	})

	t.Run("visitors behind the Render proxy get their own bucket", func(t *testing.T) {
		r := limitedRouter(t, IPRateLimit(rate.Every(time.Hour), 1))
		if code := limitedHit(r, "/v1/auth/login", "10.0.0.2:1", "198.51.100.1"); code != http.StatusOK {
			t.Fatalf("first visitor status=%d", code)
		}
		if code := limitedHit(r, "/v1/auth/login", "10.0.0.2:1", "198.51.100.2"); code != http.StatusOK {
			t.Fatalf("second visitor shares the proxy bucket: status=%d", code)
		}
		if code := limitedHit(r, "/v1/auth/login", "10.0.0.2:1", "198.51.100.1"); code != http.StatusTooManyRequests {
			t.Fatalf("first visitor again status=%d", code)
		}
	})

	t.Run("health checks are exempt", func(t *testing.T) {
		r := limitedRouter(t, IPRateLimit(rate.Every(time.Hour), 1))
		for i := 0; i < 3; i++ {
			if code := limitedHit(r, "/healthz", "203.0.113.9:1", ""); code != http.StatusOK {
				t.Fatalf("healthz %d status=%d", i, code)
			}
		}
	})
}

func TestUserRateLimit(t *testing.T) {
	setUser := func(c *gin.Context) {
		if uid := c.GetHeader("X-Test-User"); uid != "" {
			c.Set(utils.CtxUserID, uid)
		}
		c.Next()
	}
	r := limitedRouter(t, setUser, UserRateLimit(rate.Every(time.Hour), 1))
	send := func(uid, remote string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		req.RemoteAddr = remote
		req.Header.Set("X-Test-User", uid)
		r.ServeHTTP(w, req)
		return w.Code
	}
	if code := send("usr_a", "203.0.113.1:1"); code != http.StatusOK {
		t.Fatalf("first status=%d", code)
	}
	if code := send("usr_a", "203.0.113.2:1"); code != http.StatusTooManyRequests {
		t.Fatalf("same user, new ip status=%d", code)
	}
	if code := send("usr_b", "203.0.113.1:1"); code != http.StatusOK {
		t.Fatalf("other user status=%d", code)
	}
	if code := send("", "203.0.113.1:1"); code != http.StatusOK {
		t.Fatalf("no user status=%d", code)
	}
}

func TestKeyedLimiterForgetsIdleKeys(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	k := newKeyedLimiter(rate.Limit(1), 1)
	k.now = func() time.Time { return now }
	k.allow("a")
	k.allow("b")
	now = now.Add(limiterIdleTTL + time.Second)
	k.allow("c")
	if got := k.size(); got != 1 {
		t.Fatalf("buckets=%d want 1", got)
	}
}
