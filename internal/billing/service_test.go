package billing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func setup(t *testing.T) (*Service, *identity.Service, *identity.RecordingMailer, *org.Service, *org.Memory, *FakeProvider) {
	t.Helper()
	icfg := identity.Settings{
		TermsVersion: "2026-06-01", PrivacyVersion: "2026-06-01",
		AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		EmailTokenTTL: 24 * time.Hour, JWTSecret: "test-secret-for-cabinet-studio-hs256",
		Issuer: identity.Issuer,
	}
	ident, imail, imem := identity.NewMemoryService(icfg)
	orgs, _, omem := org.NewMemoryService(imem.Users(), 14)
	svc, _, fp := NewMemoryService(omem.Orgs(), omem.Members(), omem.Invites(), imem.Users(), Settings{
		Provider: ProviderFake, WebhookSecret: "test-webhook-secret",
	})
	return svc, ident, imail, orgs, omem, fp
}

func codeOf(err error) utils.Code {
	var e *utils.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func mustUser(t *testing.T, ident *identity.Service, mail *identity.RecordingMailer, email string) *identity.User {
	t.Helper()
	u, err := ident.Register(context.Background(), identity.RegisterInput{
		Email: email, Password: "correct-horse-battery-staple-19", DisplayName: "Ada Lovelace",
		TermsVersion: "2026-06-01", PrivacyVersion: "2026-06-01",
	})
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	call, ok := mail.LastVerification()
	if !ok {
		t.Fatal("missing verification mail")
	}
	u, err = ident.VerifyEmail(context.Background(), call.Token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return u
}

func TestListPlansExcludesTrial(t *testing.T) {
	svc, _, _, _, _, _ := setup(t)
	items, err := svc.ListPlans(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	for _, p := range items {
		if p.Code == PlanTrial || !p.Sellable {
			t.Fatalf("unexpected plan %+v", p)
		}
	}
}

func TestGetSubscriptionCreatesTrial(t *testing.T) {
	svc, ident, mail, orgs, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	sub, err := svc.GetSubscription(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sub.PlanCode != PlanTrial || sub.Status != StatusTrialing {
		t.Fatalf("%+v", sub)
	}
	if !strings.HasPrefix(sub.ID, PrefixSub) {
		t.Fatalf("id=%s", sub.ID)
	}
	if sub.SeatsIncluded != 3 {
		t.Fatalf("seats=%d", sub.SeatsIncluded)
	}
	again, err := svc.GetSubscription(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != sub.ID {
		t.Fatalf("expected same trial row %s vs %s", sub.ID, again.ID)
	}
}

func TestCheckoutThenWebhookActivates(t *testing.T) {
	svc, ident, mail, orgs, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := svc.Checkout(context.Background(), u.ID, o.ID, PlanIDTeam, "cabinetstudio://billing/success", "cabinetstudio://billing/cancel")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(sess.ID, PrefixCheckout) || !strings.Contains(sess.URL, sess.ID) {
		t.Fatalf("%+v", sess)
	}
	pre, err := svc.GetSubscription(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if pre.Status != StatusTrialing {
		t.Fatalf("checkout must not activate: %+v", pre)
	}
	if err := svc.CompleteCheckout(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	sub, err := svc.GetSubscription(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Status != StatusActive || sub.PlanCode != PlanTeam {
		t.Fatalf("%+v", sub)
	}
	items, _, err := svc.ListInvoices(context.Background(), u.ID, o.ID, "", "", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Status != InvoicePaid {
		t.Fatalf("invoices=%+v", items)
	}
	o2, err := orgs.Get(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if o2.Provider != ProviderFake || o2.ProviderCustomerID == "" {
		t.Fatalf("org provider=%q customer=%q", o2.Provider, o2.ProviderCustomerID)
	}
}

func TestWebhookReplayNoop(t *testing.T) {
	svc, ident, mail, orgs, _, fp := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := svc.Checkout(context.Background(), u.ID, o.ID, PlanIDStudio, "https://app.example/ok", "https://app.example/cancel")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := fp.CompletedPayload(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.HandleWebhook(context.Background(), raw); err != nil {
		t.Fatal(err)
	}
	if err := svc.HandleWebhook(context.Background(), raw); err != nil {
		t.Fatalf("replay: %v", err)
	}
	items, _, err := svc.ListInvoices(context.Background(), u.ID, o.ID, "", "", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("replay inserted extra invoices: %d", len(items))
	}
}

func TestCancelAndResume(t *testing.T) {
	svc, ident, mail, orgs, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := svc.Checkout(context.Background(), u.ID, o.ID, PlanIDTeam, "cabinetstudio://ok", "cabinetstudio://cancel")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteCheckout(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	sub, err := svc.Cancel(context.Background(), u.ID, o.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if !sub.CancelAtPeriodEnd || sub.Status != StatusActive {
		t.Fatalf("%+v", sub)
	}
	sub, err = svc.Resume(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sub.CancelAtPeriodEnd || sub.CanceledAt != nil {
		t.Fatalf("%+v", sub)
	}
}

func TestViewerForbiddenOnCheckout(t *testing.T) {
	svc, ident, mail, orgs, omem, _ := setup(t)
	owner := mustUser(t, ident, mail, "owner@example.com")
	viewer := mustUser(t, ident, mail, "viewer@example.com")
	o, _, err := orgs.Create(context.Background(), owner.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := omem.Members().Insert(context.Background(), &org.Membership{
		ID: "mbr_viewer1", OrganizationID: o.ID, UserID: viewer.ID,
		Role: org.RoleViewer, Status: org.MemberActive, ActivatedAt: &now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Checkout(context.Background(), viewer.ID, o.ID, PlanIDTeam, "cabinetstudio://ok", "cabinetstudio://cancel")
	if codeOf(err) != utils.CodeForbidden {
		t.Fatalf("code=%s err=%v", codeOf(err), err)
	}
}

func TestPreviewDowngradeSeatLimit(t *testing.T) {
	svc, ident, mail, orgs, omem, _ := setup(t)
	owner := mustUser(t, ident, mail, "owner@example.com")
	extra := mustUser(t, ident, mail, "extra@example.com")
	o, _, err := orgs.Create(context.Background(), owner.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := omem.Members().Insert(context.Background(), &org.Membership{
		ID: "mbr_extra1", OrganizationID: o.ID, UserID: extra.ID,
		Role: org.RoleViewer, Status: org.MemberActive, ActivatedAt: &now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	sess, err := svc.Checkout(context.Background(), owner.ID, o.ID, PlanIDTeam, "cabinetstudio://ok", "cabinetstudio://cancel")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteCheckout(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	_, err = svc.PreviewDowngrade(context.Background(), owner.ID, o.ID, PlanIDStudio, nil)
	if codeOf(err) != utils.CodeSeatLimit {
		t.Fatalf("code=%s err=%v", codeOf(err), err)
	}
	one := 1
	_, err = svc.PreviewDowngrade(context.Background(), owner.ID, o.ID, "", &one)
	if codeOf(err) != utils.CodeSeatLimit {
		t.Fatalf("seats code=%s err=%v", codeOf(err), err)
	}
}

func TestCheckoutUnverifiedEmail(t *testing.T) {
	svc, ident, _, orgs, omem, _ := setup(t)
	u, err := ident.Register(context.Background(), identity.RegisterInput{
		Email: "raw@example.com", Password: "correct-horse-battery-staple-19", DisplayName: "Raw",
		TermsVersion: "2026-06-01", PrivacyVersion: "2026-06-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	trial := now.Add(14 * 24 * time.Hour)
	o := &org.Organization{
		ID: "org_unverified", Name: "X", OwnerUserID: u.ID, Status: org.OrgActive,
		TrialEndsAt: &trial, CreatedAt: now, UpdatedAt: now,
	}
	if err := omem.Orgs().Insert(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	if err := omem.Members().Insert(context.Background(), &org.Membership{
		ID: "mbr_unverified", OrganizationID: o.ID, UserID: u.ID,
		Role: org.RoleOwner, Status: org.MemberActive, ActivatedAt: &now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Checkout(context.Background(), u.ID, o.ID, PlanIDTeam, "cabinetstudio://ok", "cabinetstudio://cancel")
	if codeOf(err) != utils.CodeEmailUnverified {
		t.Fatalf("code=%s err=%v", codeOf(err), err)
	}
	_ = orgs
}
