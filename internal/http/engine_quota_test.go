package httpapi

import (
	"net/http"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

func TestEngineQuota429RetryAfter(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	start := startEngine(t, f)
	hr := engineRouter(t, func(cfg *config.Config) {
		start(cfg)
		cfg.Engine.UserPerMinute = 2
	})
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-quota")
	for i := 0; i < 2; i++ {
		if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access)); w.Code != http.StatusOK {
			t.Fatalf("hit %d=%d %s", i, w.Code, w.Body.String())
		}
	}
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/estimate", quoteBody, bearer(access))
	if w.Code != http.StatusTooManyRequests || errorCode(t, w) != "quota_exceeded" || w.Header().Get("Retry-After") != "30" {
		t.Fatalf("over quota=%d retry=%q %s", w.Code, w.Header().Get("Retry-After"), w.Body.String())
	}
	if n := f.seen().calls; n != 2 {
		t.Fatalf("engine calls=%d, refused call reached the engine", n)
	}
}

func TestEngineDailyCap429(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	start := startEngine(t, f)
	hr := engineRouter(t, func(cfg *config.Config) {
		start(cfg)
		cfg.Engine.OrgPerDay = 1
	})
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-eng-daily")
	createOrgAs(t, hr.h, access, "org-eng-daily")
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("first=%d %s", w.Code, w.Body.String())
	}
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access))
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatalf("daily cap=%d %s", w.Code, w.Body.String())
	}
}
