package httpapi

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

var cutlistBody = map[string]any{
	"project":   map[string]any{"cabinets": []any{}},
	"scope":     "room-1",
	"priceBook": map[string]any{"rate": 1}, // must never reach the engine
}

func TestEngineCutlistForwardsProjectWithoutBook(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"cutlist":[],"csv":{"ok":true,"contents":"a"}}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-mfg-1")
	createOrgAs(t, hr.h, access, "org-mfg-1")
	if w := doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("put book=%d %s", w.Code, w.Body.String())
	}
	for _, route := range []string{"cutlist", "cnc"} {
		w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/"+route, cutlistBody, bearer(access))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"cutlist":[]`) {
			t.Fatalf("%s=%d %s", route, w.Code, w.Body.String())
		}
		got := f.seen()
		if got.path != "/v1/ops/"+route || string(got.body["scope"]) != `"room-1"` || string(got.body["project"]) != `{"cabinets":[]}` {
			t.Fatalf("%s: path=%s body=%v", route, got.path, got.body)
		}
		if _, leaked := got.body["priceBook"]; leaked {
			t.Fatalf("%s: a price book reached the engine: %s", route, got.body["priceBook"])
		}
	}
}

func TestEngineProductionPackageInjectsOrgBook(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{"slug":"s","files":[],"serialized":""}}`}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-mfg-2")
	createOrgAs(t, hr.h, access, "org-mfg-2")
	doJSON(t, hr.h, http.MethodPut, "/v1/org/price-book", map[string]any{"book": map[string]any{"carcass": 450}}, bearer(access))
	body := map[string]any{"project": map[string]any{}, "room": map[string]any{"dimensions": map[string]any{}}, "priceBook": map[string]any{"rate": 1}}
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/production-package", body, bearer(access))
	if w.Code != http.StatusOK {
		t.Fatalf("package=%d %s", w.Code, w.Body.String())
	}
	got := f.seen()
	if got.path != "/v1/ops/production-package" || string(got.body["priceBook"]) != `{"carcass":450}` || string(got.body["room"]) != `{"dimensions":{}}` {
		t.Fatalf("path=%s body=%v", got.path, got.body)
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/cnc", map[string]any{"document": map[string]any{}}, bearer(access)); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("cnc without project=%d %s", w.Code, w.Body.String())
	}
}

func TestEngineManufacturingHasItsOwnQuota(t *testing.T) {
	f := &fakeEngine{reply: `{"result":{}}`}
	start := startEngine(t, f)
	hr := engineRouter(t, func(cfg *config.Config) {
		start(cfg)
		cfg.Engine.MfgUserPerMinute = 1
	})
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-mfg-quota")
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/cutlist", cutlistBody, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("first=%d %s", w.Code, w.Body.String())
	}
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/cnc", cutlistBody, bearer(access))
	if w.Code != http.StatusTooManyRequests || errorCode(t, w) != "quota_exceeded" || w.Header().Get("Retry-After") == "" {
		t.Fatalf("over mfg quota=%d %s", w.Code, w.Body.String())
	}
	if w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/quote", quoteBody, bearer(access)); w.Code != http.StatusOK {
		t.Fatalf("quote must not share the manufacturing bucket: %d %s", w.Code, w.Body.String())
	}
}

func TestEngineGzipPassthrough(t *testing.T) {
	reply := `{"result":{"files":[{"name":"cabinet-machine-preview.json"}]}}`
	f := &fakeEngine{reply: reply, gzip: true}
	hr := engineRouter(t, startEngine(t, f))
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, "ada@example.com", "reg-mfg-gzip")
	headers := bearer(access)
	headers["Accept-Encoding"] = "gzip, br"
	w := doJSON(t, hr.h, http.MethodPost, "/v1/engine/cnc", cutlistBody, headers)
	if w.Code != http.StatusOK || w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("gzip client=%d enc=%q", w.Code, w.Header().Get("Content-Encoding"))
	}
	if f.seen().headers.Get("Accept-Encoding") != "gzip" {
		t.Fatalf("proxy did not ask the engine for gzip")
	}
	zr, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if plain, _ := io.ReadAll(zr); string(plain) != reply {
		t.Fatalf("unpacked=%s", plain)
	}
	w = doJSON(t, hr.h, http.MethodPost, "/v1/engine/cnc", cutlistBody, bearer(access))
	if w.Code != http.StatusOK || w.Header().Get("Content-Encoding") != "" || strings.TrimSpace(w.Body.String()) != reply {
		t.Fatalf("plain client=%d enc=%q %s", w.Code, w.Header().Get("Content-Encoding"), w.Body.String())
	}
}
