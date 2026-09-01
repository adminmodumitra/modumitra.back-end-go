package httpapi

import (
	"bytes"
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
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func billingRouter(t *testing.T) (http.Handler, *identity.RecordingMailer, config.Config) {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	if cfg.BillingWebhookSecret == "" {
		cfg.BillingWebhookSecret = "cabinet-studio-billing-webhook-local-dev-only"
	}
	svc, mail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orgs, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{
		Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret,
	})
	return NewRouter(cfg, fakePinger{}, svc, orgs, devs, bills), mail, cfg
}

func TestGetBillingPlansAuthedNoOrg(t *testing.T) {
	h, mail, _ := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-plans")
	w := doJSON(t, h, http.MethodGet, "/v1/billing/plans", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	items, _ := body["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items=%v", body["items"])
	}
	for _, it := range items {
		m := it.(map[string]any)
		if m["plan"] == "trial" {
			t.Fatal("trial in catalog")
		}
		if _, ok := m["included_seats"]; !ok {
			t.Fatalf("missing included_seats: %v", m)
		}
	}
}

func TestCheckoutSessionOwner(t *testing.T) {
	h, mail, _ := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-co")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Lovelace Cabinets"}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "org-co-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodPost, "/v1/organizations/"+orgID+"/checkout-session", map[string]any{
		"plan_id":     "plan_team_monthly",
		"success_url": "cabinetstudio://billing/success",
		"cancel_url":  "cabinetstudio://billing/cancel",
	}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "co-1",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("checkout=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	cs, _ := body["checkout_session_id"].(string)
	url, _ := body["url"].(string)
	if !strings.HasPrefix(cs, "cs_") || !strings.Contains(url, cs) {
		t.Fatalf("body=%v", body)
	}
}

func TestWebhookInvalidHMAC(t *testing.T) {
	h, _, _ := billingRouter(t)
	body := []byte(`{"id":"evt_1","type":"invoice.paid","data":{"object":{}}}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/billing", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "invalid_signature" {
		t.Fatalf("code=%s body=%s", errorCode(t, w), w.Body.String())
	}
}

func TestWebhookValidHMACActivatesSubscription(t *testing.T) {
	h, mail, cfg := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-wh")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Lovelace Cabinets"}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "org-wh-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodPost, "/v1/organizations/"+orgID+"/checkout-session", map[string]any{
		"plan_id":     "plan_team_monthly",
		"success_url": "https://app.example/ok",
		"cancel_url":  "https://app.example/cancel",
	}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "wh-co-1",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("checkout=%d %s", w.Code, w.Body.String())
	}
	var co map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &co)
	cs := co["checkout_session_id"].(string)

	payload, err := json.Marshal(map[string]any{
		"id":   "evt_http_1",
		"type": "checkout.session.completed",
		"data": map[string]any{"object": map[string]any{
			"id": cs, "organization_id": orgID, "plan_id": "plan_team_monthly",
			"customer": "cus_fake_http", "subscription": "sub_fake_http",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	sig := billing.SignBody(cfg.BillingWebhookSecret, payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/billing", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(utils.HeaderBillingSignature, sig)
	ww := httptest.NewRecorder()
	h.ServeHTTP(ww, req)
	if ww.Code != http.StatusOK {
		t.Fatalf("webhook=%d %s", ww.Code, ww.Body.String())
	}
	var rec map[string]any
	_ = json.Unmarshal(ww.Body.Bytes(), &rec)
	if rec["received"] != true {
		t.Fatalf("body=%v", rec)
	}

	w = doJSON(t, h, http.MethodGet, "/v1/organizations/"+orgID+"/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("sub=%d %s", w.Code, w.Body.String())
	}
	var subBody map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &subBody)
	sub := subBody["subscription"].(map[string]any)
	if sub["status"] != "active" {
		t.Fatalf("subscription=%v", sub)
	}
}
