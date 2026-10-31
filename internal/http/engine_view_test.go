package httpapi

import (
	"net/http"
	"testing"
)

func TestEngineViewInjectsBookForPricedOps(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"summary":{}}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-view-1")
	createOrgAs(t, hr.h, access, "org-view-1")
	doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	body := map[string]any{"args": []any{map[string]any{"cabinets": []any{}}}, "priceBook": map[string]any{"rate": 1}}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/view/createProjectReport", body, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("report=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if got.path != "/v1/ops/view/createProjectReport" || string(got.body["priceBook"]) != `{"carcass":450}` || string(got.body["args"]) != `[{"cabinets":[]}]` {
		t.Fatalf("path=%s body=%v", got.path, got.body)
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/view/buildProductionReadinessGate", body, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("gate=%d %s", w.Code, w.Body.String())
	}
	if _, leaked := f.seen().body["priceBook"]; leaked {
		t.Fatal("a price book reached the engine for an unpriced op")
	}
}

func TestEngineViewRefusesUnknownOps(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-view-2")
	for _, op := range []string{"buildLiveInteriorQuote", "quote", "nope"} {
		if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/view/"+op, map[string]any{"args": []any{}}, bearer(access)); w.Code != http.StatusNotFound {
			t.Fatalf("%s=%d %s", op, w.Code, w.Body.String())
		}
	}
	if f.seen().calls != 0 {
		t.Fatal("an unknown op reached the engine")
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/view/createProjectReport", map[string]any{"args": []any{}}, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("signed out=%d", w.Code)
	}
}

func TestEngineProductionPacketPDFForwardsWithoutBook(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"mimeType":"application/pdf","bytes":3,"base64":"AAAA"}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-view-3")
	body := map[string]any{"project": map[string]any{"cabinets": []any{}}, "room": map[string]any{"dimensions": map[string]any{}}, "projectName": "S", "priceBook": map[string]any{"rate": 1}}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/production-packet-pdf", body, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("packet=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if _, leaked := got.body["priceBook"]; leaked || got.path != "/v1/ops/production-packet-pdf" {
		t.Fatalf("path=%s body=%v", got.path, got.body)
	}
}
