package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/orglookup"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
)

const testEngineKey = "k3Q9vZp1Lr8xT2mWc7YbN4sHd6FgJ0aE-engine"

// fakeEngine records what the proxy sent and answers with status/body.
type fakeEngine struct {
	mu      sync.Mutex
	status  int
	reply   string
	path    string
	headers http.Header
	body    map[string]json.RawMessage
	calls   int
	gzip    bool // answer with Content-Encoding: gzip (Phase 4 passthrough)
}

func (f *fakeEngine) handler(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.calls++
	f.path = r.URL.Path
	f.headers = r.Header.Clone()
	f.body = map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &f.body)
	status, reply := f.status, f.reply
	f.mu.Unlock()
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	if f.gzip {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(status)
		zw := gzip.NewWriter(w)
		_, _ = io.WriteString(zw, reply)
		_ = zw.Close()
		return
	}
	w.WriteHeader(status)
	_, _ = io.WriteString(w, reply)
}

type engineSeen struct {
	path    string
	headers http.Header
	body    map[string]json.RawMessage
	calls   int
}

// seen reads what the engine got under the lock (the server runs on its own goroutine).
func (f *fakeEngine) seen() engineSeen {
	f.mu.Lock()
	defer f.mu.Unlock()
	return engineSeen{path: f.path, headers: f.headers, body: f.body, calls: f.calls}
}

type engineHarness struct {
	h     http.Handler
	mail  *identity.RecordingMailer
	omail *org.RecordingMailer
	books *pricebook.Memory
}

// engineRouter builds the API with a memory price-book store. tune edits the
// config (engine URL, timeout, quotas) before the router is built.
func engineRouter(t *testing.T, tune func(*config.Config)) *engineHarness {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	cfg.Engine.SharedSecret = testEngineKey
	if tune != nil {
		tune(&cfg)
	}
	svc, imail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, omail, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orglookup.Adapter{Service: orgs}, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret})
	books := pricebook.NewMemory()
	h := NewRouter(cfg, fakePinger{}, svc, orgs, devs, bills, memorySSO(svc, cfg), WithPriceBooks(books))
	return &engineHarness{h: h, mail: imail, omail: omail, books: books}
}

// startEngine serves f and points the router config at it.
func startEngine(t *testing.T, f *fakeEngine) func(*config.Config) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(srv.Close)
	return func(cfg *config.Config) { cfg.Engine.URL = srv.URL }
}

func bearer(access string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + access}
}

func createOrgAs(t *testing.T, h http.Handler, access, key string) string {
	t.Helper()
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Engine Cabinets"}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": key,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create org=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id, _ := created["organization"].(map[string]any)["id"].(string)
	return id
}

// joinAs invites email with role into orgID and returns the new member's access.
func joinAs(t *testing.T, hr *engineHarness, ownerAccess, orgID, email, role string) string {
	t.Helper()
	w := doJSON(t, hr.h, http.MethodPost, "/v1/organizations/"+orgID+"/invitations", map[string]any{
		"email": email, "role": role,
	}, map[string]string{"Authorization": "Bearer " + ownerAccess, "Idempotency-Key": "inv-" + email})
	if w.Code != http.StatusCreated {
		t.Fatalf("invite=%d %s", w.Code, w.Body.String())
	}
	inv, ok := hr.omail.LastInvite()
	if !ok {
		t.Fatal("missing invite mail")
	}
	access, _ := registerVerifyLogin(t, hr.h, hr.mail, email, "reg-"+email)
	w = doJSON(t, hr.h, http.MethodPost, "/v1/invitations/"+inv.Token+"/accept", nil, map[string]string{
		"Authorization": "Bearer " + access, "Idempotency-Key": "acc-" + email,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("accept=%d %s", w.Code, w.Body.String())
	}
	return access
}
