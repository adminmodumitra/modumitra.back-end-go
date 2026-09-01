package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func deviceRouter(t *testing.T) (http.Handler, *identity.RecordingMailer) {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, mail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, _ := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orgs, cfg.EntitlementSigningKey)
	return NewRouter(cfg, fakePinger{}, svc, orgs, devs), mail
}

func TestActivateDevice201(t *testing.T) {
	h, mail := deviceRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-dev-1")
	w := doJSON(t, h, http.MethodPost, "/v1/devices/activate", map[string]any{
		"name":     "Studio iMac",
		"platform": "macos",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "dev-act-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	dev, _ := body["device"].(map[string]any)
	if dev["platform"] != "macos" || dev["name"] != "Studio iMac" {
		t.Fatalf("device=%v", dev)
	}
	snap, _ := body["entitlement_snapshot"].(map[string]any)
	if snap["signature"] == nil || snap["signature"] == "" {
		t.Fatalf("snap=%v", snap)
	}
	if snap["plan"] != "none" {
		t.Fatalf("plan=%v", snap["plan"])
	}
}

func TestListDevicesActiveCount(t *testing.T) {
	h, mail := deviceRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-dev-list")
	w := doJSON(t, h, http.MethodPost, "/v1/devices/activate", map[string]any{
		"name": "A", "platform": "macos",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "dev-list-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("activate=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodGet, "/v1/devices", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("list=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["active_count"] != float64(1) || body["device_limit"] != float64(3) {
		t.Fatalf("body=%v", body)
	}
	items, _ := body["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items=%v", body["items"])
	}
}

func TestFourthActivate409(t *testing.T) {
	h, mail := deviceRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-dev-lim")
	for i := 0; i < 3; i++ {
		w := doJSON(t, h, http.MethodPost, "/v1/devices/activate", map[string]any{
			"name": "D", "platform": "linux",
		}, map[string]string{
			"Authorization":   "Bearer " + access,
			"Idempotency-Key": "dev-lim-" + string(rune('a'+i)),
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("activate %d=%d %s", i, w.Code, w.Body.String())
		}
	}
	w := doJSON(t, h, http.MethodPost, "/v1/devices/activate", map[string]any{
		"name": "Overflow", "platform": "windows",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "dev-lim-overflow",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("status=%d %s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "device_limit_reached" {
		t.Fatalf("code=%s", errorCode(t, w))
	}
}

func TestOrgEntitlementsSigned(t *testing.T) {
	h, mail := deviceRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-ent")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{
		"name": "Lovelace Cabinets",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "org-ent-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodGet, "/v1/organizations/"+orgID+"/entitlements", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("entitlements=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	snap, _ := body["entitlement_snapshot"].(map[string]any)
	if snap["plan"] != "trial" || snap["subscription_status"] != "trialing" {
		t.Fatalf("snap=%v", snap)
	}
	if snap["signature"] == nil || snap["signature"] == "" {
		t.Fatal("missing signature")
	}
}

func TestMeCapabilities(t *testing.T) {
	h, mail := deviceRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-caps")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{
		"name": "Lovelace Cabinets",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "org-caps-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodGet, "/v1/me/capabilities?organization_id="+orgID, nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("capabilities=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["role"] != "owner" || body["plan"] != "trial" {
		t.Fatalf("body=%v", body)
	}
	if _, has := body["signature"]; has {
		t.Fatal("capabilities must not be signed")
	}
	caps, _ := body["capabilities"].(map[string]any)
	if caps["cloud_projects"] != true {
		t.Fatalf("caps=%v", caps)
	}
}
