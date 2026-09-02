package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"

	"google.golang.org/protobuf/proto"
)

func orgRouter(t *testing.T) (http.Handler, *identity.RecordingMailer, *org.RecordingMailer) {
	t.Helper()
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, imail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, omail, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orgs, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret})
	return NewRouter(cfg, fakePinger{}, svc, orgs, devs, bills, memorySSO(svc, cfg)), imail, omail
}

func registerVerifyLogin(t *testing.T, h http.Handler, mail *identity.RecordingMailer, email, key string) (access, userID string) {
	t.Helper()
	w := doJSON(t, h, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":           email,
		"password":        "correct-horse-battery-staple-19",
		"display_name":    "Ada Lovelace",
		"terms_version":   "2026-06-01",
		"privacy_version": "2026-06-01",
	}, map[string]string{"Idempotency-Key": key})
	if w.Code != http.StatusCreated {
		t.Fatalf("register=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	user, _ := created["user"].(map[string]any)
	userID, _ = user["id"].(string)
	call, ok := mail.LastVerification()
	if !ok {
		t.Fatal("missing verification mail")
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/verify-email", map[string]any{"token": call.Token}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("verify=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": "correct-horse-battery-staple-19",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login=%d %s", w.Code, w.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	access, _ = login["access_token"].(string)
	if access == "" {
		t.Fatal("missing access")
	}
	return access, userID
}

func TestCreateAndListOrganization(t *testing.T) {
	h, mail, _ := orgRouter(t)
	access, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-org-1")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{
		"name": "Lovelace Cabinets",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "org-create-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	orgObj, _ := created["organization"].(map[string]any)
	id, _ := orgObj["id"].(string)
	if id == "" {
		t.Fatalf("body=%v", created)
	}
	w = doJSON(t, h, http.MethodGet, "/v1/organizations", nil, map[string]string{
		"Authorization": "Bearer " + access,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("list=%d %s", w.Code, w.Body.String())
	}
	var listed map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	items, _ := listed["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items=%v", listed["items"])
	}
}

func TestGetOrgOtherUser404(t *testing.T) {
	h, mail, _ := orgRouter(t)
	accessA, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-a")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Ada Org"}, map[string]string{
		"Authorization":   "Bearer " + accessA,
		"Idempotency-Key": "org-a",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	accessB, _ := registerVerifyLogin(t, h, mail, "charles@example.com", "reg-b")
	w = doJSON(t, h, http.MethodGet, "/v1/organizations/"+orgID, nil, map[string]string{
		"Authorization": "Bearer " + accessB,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d %s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "not_found" {
		t.Fatalf("code=%s", errorCode(t, w))
	}
}

func TestViewerForbiddenOnPatch(t *testing.T) {
	h, mail, omail := orgRouter(t)
	accessA, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-own")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Ada Org"}, map[string]string{
		"Authorization":   "Bearer " + accessA,
		"Idempotency-Key": "org-own",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodPost, "/v1/organizations/"+orgID+"/invitations", map[string]any{
		"email": "viewer@example.com",
		"role":  "viewer",
	}, map[string]string{
		"Authorization":   "Bearer " + accessA,
		"Idempotency-Key": "inv-v",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("invite=%d %s", w.Code, w.Body.String())
	}
	inv, ok := omail.LastInvite()
	if !ok {
		t.Fatal("missing invite mail")
	}
	accessV, _ := registerVerifyLogin(t, h, mail, "viewer@example.com", "reg-v")
	w = doJSON(t, h, http.MethodPost, "/v1/invitations/"+inv.Token+"/accept", nil, map[string]string{
		"Authorization":   "Bearer " + accessV,
		"Idempotency-Key": "acc-v",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("accept=%d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, h, http.MethodPatch, "/v1/organizations/"+orgID, map[string]any{"name": "Nope"}, map[string]string{
		"Authorization": "Bearer " + accessV,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("patch=%d %s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "forbidden" {
		t.Fatalf("code=%s", errorCode(t, w))
	}
}

func TestInvitationPreviewPublicAndAccept(t *testing.T) {
	h, mail, omail := orgRouter(t)
	accessA, _ := registerVerifyLogin(t, h, mail, "ada@example.com", "reg-inv")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{"name": "Lovelace Cabinets"}, map[string]string{
		"Authorization":   "Bearer " + accessA,
		"Idempotency-Key": "org-inv",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	w = doJSON(t, h, http.MethodPost, "/v1/organizations/"+orgID+"/invitations", map[string]any{
		"email": "charles@example.com",
		"role":  "sales_designer",
	}, map[string]string{
		"Authorization":   "Bearer " + accessA,
		"Idempotency-Key": "inv-c",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("invite=%d %s", w.Code, w.Body.String())
	}
	inv, ok := omail.LastInvite()
	if !ok {
		t.Fatal("missing invite")
	}
	w = doJSON(t, h, http.MethodGet, "/v1/invitations/"+inv.Token, nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("preview=%d %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate=%q", got)
	}
	var preview map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &preview)
	if preview["organization_name"] != "Lovelace Cabinets" || preview["role"] != "sales_designer" {
		t.Fatalf("preview=%v", preview)
	}
	if _, has := preview["organization_id"]; has {
		t.Fatal("organization_id leaked")
	}
	accessC, _ := registerVerifyLogin(t, h, mail, "charles@example.com", "reg-c")
	w = doJSON(t, h, http.MethodPost, "/v1/invitations/"+inv.Token+"/accept", nil, map[string]string{
		"Authorization":   "Bearer " + accessC,
		"Idempotency-Key": "acc-c",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("accept=%d %s", w.Code, w.Body.String())
	}
}

func TestMeStill401WithoutToken(t *testing.T) {
	h, _, _ := orgRouter(t)
	w := doJSON(t, h, http.MethodGet, "/v1/me", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d %s", w.Code, w.Body.String())
	}
	if errorCode(t, w) != "unauthenticated" {
		t.Fatalf("code=%s", errorCode(t, w))
	}
}

func TestCreateOrganizationOutboxProto(t *testing.T) {
	cfg := config.Load()
	cfg.ServiceName = "cabinet-studio-api-test"
	svc, mail, mem := identity.NewMemoryService(testSettings(cfg))
	orgs, _, orgMem := org.NewMemoryService(mem.Users(), cfg.TrialDays)
	devs, _ := device.NewMemoryService(orgs, cfg.EntitlementSigningKey)
	bills, _, _ := billing.NewMemoryService(orgMem.Orgs(), orgMem.Members(), orgMem.Invites(), mem.Users(), billing.Settings{Provider: billing.ProviderFake, WebhookSecret: cfg.BillingWebhookSecret})
	h := NewRouter(cfg, fakePinger{}, svc, orgs, devs, bills, memorySSO(svc, cfg))
	access, _ := registerVerifyLogin(t, h, mail, "ada-outbox@example.com", "reg-org-outbox")
	w := doJSON(t, h, http.MethodPost, "/v1/organizations", map[string]any{
		"name": "Outbox Cabinets",
	}, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "org-create-outbox",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	orgID, _ := created["organization"].(map[string]any)["id"].(string)
	mw, ok := orgs.Outbox().(*outbox.MemoryWriter)
	if !ok {
		t.Fatal("expected memory outbox")
	}
	var found *pb.Envelope
	for _, row := range mw.Rows() {
		var env pb.Envelope
		if err := proto.Unmarshal(row.Payload, &env); err != nil {
			t.Fatalf("not protobuf: %v", err)
		}
		if env.EventType == outbox.EventOrgCreated {
			found = &env
		}
	}
	if found == nil {
		t.Fatal("missing org.created proto in outbox")
	}
	if found.GetOrgCreated().GetOrganizationId() != orgID {
		t.Fatalf("org_id=%s want %s", found.GetOrgCreated().GetOrganizationId(), orgID)
	}
}
