package device_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/orglookup"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func setup(t *testing.T) (*device.Service, *org.Service, *identity.Service, *identity.RecordingMailer, device.HMACSigner, *device.Memory) {
	t.Helper()
	icfg := identity.Settings{
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     30 * 24 * time.Hour,
		EmailTokenTTL:  24 * time.Hour,
		JWTSecret:      "test-secret-for-cabinet-studio-hs256",
		Issuer:         identity.Issuer,
	}
	ident, imail, mem := identity.NewMemoryService(icfg)
	orgs, _, _ := org.NewMemoryService(mem.Users(), 14)
	signer := device.NewHMACSigner("test-entitlement-signing-key", device.KeyIDV1)
	clk := &device.FrozenClock{T: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}
	store := device.NewMemory()
	svc := device.NewService(store.Devices(), store.Entitlements(), store, orglookup.Adapter{Service: orgs}, signer, clk, device.NewIDGen(), outbox.NewMemoryWriter())
	return svc, orgs, ident, imail, signer, store
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
	_, err := ident.Register(context.Background(), identity.RegisterInput{
		Email:          email,
		Password:       "correct-horse-battery-staple-19",
		DisplayName:    "Ada Lovelace",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
	})
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	call, ok := mail.LastVerification()
	if !ok {
		t.Fatal("missing verification mail")
	}
	u, err := ident.VerifyEmail(context.Background(), call.Token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return u
}

func TestActivateFirstDeviceSigns(t *testing.T) {
	svc, _, ident, mail, signer, store := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	d, snap, err := svc.Activate(context.Background(), u.ID, "Studio iMac", device.PlatformMacOS)
	if err != nil {
		t.Fatal(err)
	}
	if d == nil || !strings.HasPrefix(d.ID, device.PrefixDevice) {
		t.Fatalf("device=%+v", d)
	}
	if snap == nil || snap.Signature == "" {
		t.Fatal("missing signature")
	}
	if snap.Plan != device.PlanNone || snap.OrganizationID != nil {
		t.Fatalf("snap=%+v", snap)
	}
	if !signer.VerifySnapshot(*snap) {
		t.Fatal("signature did not verify")
	}
	if store.EntitlementCount() != 1 {
		t.Fatalf("ents=%d", store.EntitlementCount())
	}
}

func TestFourthActiveDeviceLimit(t *testing.T) {
	svc, _, ident, mail, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	for i := 0; i < device.DeviceLimit; i++ {
		if _, _, err := svc.Activate(context.Background(), u.ID, "Dev", device.PlatformMacOS); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
	}
	_, _, err := svc.Activate(context.Background(), u.ID, "Overflow", device.PlatformLinux)
	if codeOf(err) != utils.CodeDeviceLimit {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestDeactivateFreesSlot(t *testing.T) {
	svc, _, ident, mail, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	var first *device.Device
	for i := 0; i < device.DeviceLimit; i++ {
		d, _, err := svc.Activate(context.Background(), u.ID, "Dev", device.PlatformMacOS)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = d
		}
	}
	if _, err := svc.Deactivate(context.Background(), u.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Activate(context.Background(), u.ID, "Replacement", device.PlatformWindows); err != nil {
		t.Fatalf("expected slot freed: %v", err)
	}
}

func TestPatchDeleteOtherUserNotFound(t *testing.T) {
	svc, _, ident, mail, _, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	d, _, err := svc.Activate(context.Background(), a.ID, "Ada Mac", device.PlatformMacOS)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Rename(context.Background(), b.ID, d.ID, "Stolen"); codeOf(err) != utils.CodeNotFound {
		t.Fatalf("rename err=%v code=%s", err, codeOf(err))
	}
	if _, err := svc.Deactivate(context.Background(), b.ID, d.ID); codeOf(err) != utils.CodeNotFound {
		t.Fatalf("delete err=%v code=%s", err, codeOf(err))
	}
}

func TestSnapshotNoOrgPlanNone(t *testing.T) {
	svc, _, ident, mail, signer, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	_, snap, err := svc.Activate(context.Background(), u.ID, "Laptop", device.PlatformLinux)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Plan != device.PlanNone || snap.SubscriptionStatus != device.SubNone {
		t.Fatalf("plan=%s status=%s", snap.Plan, snap.SubscriptionStatus)
	}
	if snap.OrganizationID != nil {
		t.Fatalf("org=%v", snap.OrganizationID)
	}
	if snap.Capabilities.CloudProjects || snap.Capabilities.MaxSeats != 0 {
		t.Fatalf("caps=%+v", snap.Capabilities)
	}
	if snap.GraceUntil != nil {
		t.Fatalf("grace=%v", snap.GraceUntil)
	}
	if snap.Signature == "" || !signer.VerifySnapshot(*snap) {
		t.Fatal("signature")
	}
}

func TestSnapshotTrialOrg(t *testing.T) {
	svc, orgs, ident, mail, signer, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	_, snap, err := svc.Activate(context.Background(), u.ID, "Studio iMac", device.PlatformMacOS)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Plan != device.PlanTrial || snap.SubscriptionStatus != device.SubTrialing {
		t.Fatalf("plan=%s status=%s", snap.Plan, snap.SubscriptionStatus)
	}
	if snap.OrganizationID == nil || *snap.OrganizationID != o.ID {
		t.Fatalf("org=%v want %s", snap.OrganizationID, o.ID)
	}
	if !snap.Capabilities.CloudProjects || snap.Capabilities.MaxDevices != 3 {
		t.Fatalf("caps=%+v", snap.Capabilities)
	}
	if snap.GraceUntil == nil {
		t.Fatal("grace nil")
	}
	if !signer.VerifySnapshot(*snap) {
		t.Fatal("signature")
	}
}

func TestIssueDeviceNotOwned(t *testing.T) {
	svc, orgs, ident, mail, _, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := svc.Activate(context.Background(), b.ID, "Charles PC", device.PlatformWindows)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Issue(context.Background(), a.ID, o.ID, other.ID)
	if codeOf(err) != utils.CodeDeviceNotOwned {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestDeactivateIdempotent(t *testing.T) {
	svc, _, ident, mail, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	d, _, err := svc.Activate(context.Background(), u.ID, "Mac", device.PlatformMacOS)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Deactivate(context.Background(), u.ID, d.ID); err != nil {
		t.Fatal(err)
	}
	again, err := svc.Deactivate(context.Background(), u.ID, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != device.StatusDeactivated {
		t.Fatalf("status=%s", again.Status)
	}
}
