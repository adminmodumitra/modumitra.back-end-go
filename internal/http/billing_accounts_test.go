package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRegisterCreatesPersonalTrial(t *testing.T) {
	h, mail, _ := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-ba-1")
	w := doJSON(t, h, http.MethodGet, "/v1/billing/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	sub := body["subscription"].(map[string]any)
	if sub["status"] != "trialing" || sub["plan"] != "trial" {
		t.Fatalf("sub=%v", sub)
	}
	id := sub["id"].(string)
	w = doJSON(t, h, http.MethodGet, "/v1/billing/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	again := body["subscription"].(map[string]any)["id"].(string)
	if again != id {
		t.Fatalf("duplicate personal sub %s vs %s", id, again)
	}
}

func TestCreateOrgKeepsPersonalSubscription(t *testing.T) {
	h, mail, _ := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-ba-2")
	w := doJSON(t, h, http.MethodGet, "/v1/billing/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	var personal map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &personal)
	personalID := personal["subscription"].(map[string]any)["id"].(string)

	w = doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Lovelace Cabinets"}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "org-ba-2",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID := created["organization"].(map[string]any)["id"].(string)

	w = doJSON(t, h, http.MethodGet, "/v1/billing/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access, "X-Organization-Id": orgID,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("org sub=%d %s", w.Code, w.Body.String())
	}
	var orgBody map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &orgBody)
	orgSub := orgBody["subscription"].(map[string]any)
	if orgSub["id"] == personalID {
		t.Fatalf("org sub stole personal id %s", personalID)
	}
	if orgSub["status"] != "trialing" || orgSub["organization_id"] != orgID {
		t.Fatalf("org sub=%v", orgSub)
	}

	w = doJSON(t, h, http.MethodGet, "/v1/billing/subscription", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	_ = json.Unmarshal(w.Body.Bytes(), &personal)
	if personal["subscription"].(map[string]any)["id"] != personalID {
		t.Fatalf("personal sub changed: %v", personal)
	}
}

func TestCapabilitiesPersonalAndUnion(t *testing.T) {
	h, mail, _ := billingRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-ba-caps")
	w := doJSON(t, h, http.MethodGet, "/v1/me/capabilities", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("personal caps=%d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["plan"] != "trial" {
		t.Fatalf("personal plan=%v", body)
	}
	caps := body["capabilities"].(map[string]any)
	if caps["cloud_projects"] != true {
		t.Fatalf("personal caps=%v", caps)
	}

	w = doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Lovelace Cabinets"}, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "org-ba-caps",
	})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodGet, "/v1/me/capabilities", nil, map[string]string{
		"Authorization": "Bearer " + access, "X-Organization-Id": orgID,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("org caps=%d %s", w.Code, w.Body.String())
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["plan"] != "trial" || body["role"] != "owner" {
		t.Fatalf("union=%v", body)
	}
}
