package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func fixedQuota(userPerMin, orgPerMin, orgPerDay int, at *time.Time) *EngineQuota {
	q := NewEngineQuota(userPerMin, orgPerMin, orgPerDay)
	q.now = func() time.Time { return *at }
	return q
}

func TestEngineQuotaPerUser(t *testing.T) {
	now := time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)
	q := fixedQuota(2, 100, 1000, &now)
	for i := 0; i < 2; i++ {
		if ok, _ := q.Allow("usr_a", "org_1"); !ok {
			t.Fatalf("hit %d refused", i)
		}
	}
	ok, wait := q.Allow("usr_a", "org_1")
	if ok || wait != 30*time.Second {
		t.Fatalf("over budget ok=%v wait=%v", ok, wait)
	}
	if ok, _ := q.Allow("usr_b", "org_1"); !ok {
		t.Fatal("another user shares the bucket")
	}
	now = now.Add(30 * time.Second)
	if ok, _ := q.Allow("usr_a", "org_1"); !ok {
		t.Fatal("token did not refill")
	}
}

func TestEngineQuotaPerOrg(t *testing.T) {
	now := time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)
	q := fixedQuota(100, 3, 1000, &now)
	users := []string{"usr_a", "usr_b", "usr_c"}
	for _, u := range users {
		if ok, _ := q.Allow(u, "org_1"); !ok {
			t.Fatalf("%s refused", u)
		}
	}
	if ok, wait := q.Allow("usr_d", "org_1"); ok || wait != 20*time.Second {
		t.Fatalf("org over budget ok=%v wait=%v", ok, wait)
	}
	if ok, _ := q.Allow("usr_d", "org_2"); !ok {
		t.Fatal("another org shares the bucket")
	}
}

func TestEngineQuotaDailyCap(t *testing.T) {
	now := time.Date(2026, 10, 31, 22, 0, 0, 0, time.UTC)
	q := fixedQuota(1000, 1000, 2, &now)
	for i := 0; i < 2; i++ {
		if ok, _ := q.Allow("usr_a", "org_1"); !ok {
			t.Fatalf("hit %d refused", i)
		}
	}
	ok, wait := q.Allow("usr_a", "org_1")
	if ok || wait != 2*time.Hour {
		t.Fatalf("daily cap ok=%v wait=%v", ok, wait)
	}
	now = now.Add(2 * time.Hour) // UTC midnight resets the day
	if ok, _ := q.Allow("usr_a", "org_1"); !ok {
		t.Fatal("daily cap did not reset")
	}
}

func TestRejectQuotaRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/engine/quote", nil)
	RejectQuota(c, 1500*time.Millisecond)
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") != "2" {
		t.Fatalf("status=%d retry=%q", w.Code, w.Header().Get("Retry-After"))
	}
}
