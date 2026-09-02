package org

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func setup(t *testing.T) (*Service, *identity.Service, *identity.RecordingMailer, *RecordingMailer) {
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
	orgs, omail, _ := NewMemoryService(mem.Users(), 14)
	return orgs, ident, imail, omail
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
	u, err = ident.VerifyEmail(context.Background(), call.Token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return u
}

func TestCreateOrgOwnerAndTrial(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, m, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(o.ID, "org_") {
		t.Fatalf("id=%s", o.ID)
	}
	if m.Role != RoleOwner || m.Status != MemberActive || m.UserID != u.ID {
		t.Fatalf("membership=%+v", m)
	}
	if o.OwnerUserID != u.ID {
		t.Fatalf("owner=%s", o.OwnerUserID)
	}
	if o.TrialEndsAt == nil {
		t.Fatal("trial_ends_at nil")
	}
	want := o.CreatedAt.Add(14 * 24 * time.Hour)
	if o.TrialEndsAt.Sub(want) > time.Second || want.Sub(*o.TrialEndsAt) > time.Second {
		t.Fatalf("trial=%v want ~%v", o.TrialEndsAt, want)
	}
}

func TestCreateOrgUnverified(t *testing.T) {
	orgs, ident, _, _ := setup(t)
	u, err := ident.Register(context.Background(), identity.RegisterInput{
		Email:          "ada@example.com",
		Password:       "correct-horse-battery-staple-19",
		DisplayName:    "Ada",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if codeOf(err) != utils.CodeEmailUnverified {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestListOnlyCallerOrgs(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	oa, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	ob, _, err := orgs.Create(context.Background(), b.ID, "Charles Org")
	if err != nil {
		t.Fatal(err)
	}
	items, _, err := orgs.List(context.Background(), a.ID, "", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Organization.ID != oa.ID {
		t.Fatalf("items=%+v", items)
	}
	items, _, err = orgs.List(context.Background(), b.ID, "", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Organization.ID != ob.ID {
		t.Fatalf("items=%+v", items)
	}
}

func TestGetNonMemberNotFound(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	_, err = orgs.Get(context.Background(), b.ID, o.ID)
	if codeOf(err) != utils.CodeNotFound {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestViewerCannotPatch(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "viewer@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	_, err = orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	tok, ok := omail.LastInvite()
	if !ok {
		t.Fatal("missing invite")
	}
	if _, _, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token); err != nil {
		t.Fatal(err)
	}
	_, err = orgs.UpdateName(context.Background(), b.ID, o.ID, "Nope")
	if codeOf(err) != utils.CodeForbidden {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestAdminPatchCannotTransfer(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "admin@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleAdmin); err != nil {
		t.Fatal(err)
	}
	tok, _ := omail.LastInvite()
	if _, _, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token); err != nil {
		t.Fatal(err)
	}
	updated, err := orgs.UpdateName(context.Background(), b.ID, o.ID, "New Name")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New Name" {
		t.Fatalf("name=%s", updated.Name)
	}
	_, err = orgs.TransferOwnership(context.Background(), b.ID, o.ID, a.ID)
	if codeOf(err) != utils.CodeForbidden {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestTransferOwnership(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleAdmin); err != nil {
		t.Fatal(err)
	}
	tok, _ := omail.LastInvite()
	if _, _, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token); err != nil {
		t.Fatal(err)
	}
	updated, err := orgs.TransferOwnership(context.Background(), a.ID, o.ID, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.OwnerUserID != b.ID {
		t.Fatalf("owner=%s", updated.OwnerUserID)
	}
	old, err := orgs.members.GetByOrgUser(context.Background(), o.ID, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if old.Role != RoleAdmin {
		t.Fatalf("old role=%s", old.Role)
	}
	neu, err := orgs.members.GetByOrgUser(context.Background(), o.ID, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if neu.Role != RoleOwner {
		t.Fatalf("new role=%s", neu.Role)
	}
}

func TestLastOwnerCannotLeave(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	err = orgs.Leave(context.Background(), a.ID, o.ID)
	if codeOf(err) != utils.CodeLastOwner {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestInviteAccept(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "charles@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleEngineer); err != nil {
		t.Fatal(err)
	}
	tok, _ := omail.LastInvite()
	joined, m, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token)
	if err != nil {
		t.Fatal(err)
	}
	if joined.ID != o.ID || m.Role != RoleEngineer || m.Status != MemberActive {
		t.Fatalf("joined=%+v m=%+v", joined, m)
	}
}

func TestInviteOwnerRoleRejected(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	_, err = orgs.CreateInvitation(context.Background(), a.ID, o.ID, "x@example.com", RoleOwner)
	if codeOf(err) != utils.CodeRoleNotAssignable {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestSuspendOwnerRejected(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "admin@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleAdmin); err != nil {
		t.Fatal(err)
	}
	tok, _ := omail.LastInvite()
	if _, _, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token); err != nil {
		t.Fatal(err)
	}
	_, err = orgs.SuspendMember(context.Background(), b.ID, o.ID, a.ID)
	if codeOf(err) != utils.CodeLastOwner && codeOf(err) != utils.CodeForbidden {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestAdminCannotDemoteOwner(t *testing.T) {
	orgs, ident, mail, omail := setup(t)
	a := mustUser(t, ident, mail, "ada@example.com")
	b := mustUser(t, ident, mail, "admin@example.com")
	o, _, err := orgs.Create(context.Background(), a.ID, "Ada Org")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := orgs.CreateInvitation(context.Background(), a.ID, o.ID, b.Email, RoleAdmin); err != nil {
		t.Fatal(err)
	}
	tok, _ := omail.LastInvite()
	if _, _, err := orgs.AcceptInvitation(context.Background(), b.ID, tok.Token); err != nil {
		t.Fatal(err)
	}
	_, err = orgs.PatchMember(context.Background(), b.ID, o.ID, a.ID, RoleViewer, "")
	if codeOf(err) != utils.CodeForbidden {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestCreateOrgEnqueuesProto(t *testing.T) {
	orgs, ident, mail, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	mw, ok := orgs.Outbox().(*outbox.MemoryWriter)
	if !ok {
		t.Fatal("expected memory outbox writer")
	}
	var found *pb.Envelope
	for _, row := range mw.Rows() {
		var env pb.Envelope
		if err := proto.Unmarshal(row.Payload, &env); err != nil {
			t.Fatalf("payload not protobuf: %v", err)
		}
		if env.EventType == outbox.EventOrgCreated {
			found = &env
			if row.Topic != outbox.TopicOrg {
				t.Fatalf("topic=%s", row.Topic)
			}
		}
	}
	if found == nil {
		t.Fatal("missing org.created outbox row")
	}
	body := found.GetOrgCreated()
	if body == nil || body.OrganizationId != o.ID || body.OwnerUserId != u.ID {
		t.Fatalf("body=%+v org=%s user=%s", body, o.ID, u.ID)
	}
}
