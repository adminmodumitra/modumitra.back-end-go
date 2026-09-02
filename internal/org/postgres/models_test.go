package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func TestTableNames(t *testing.T) {
	if (tables.Organization{}).TableName() != "organizations" {
		t.Fatal("organizations")
	}
	if (tables.Membership{}).TableName() != "memberships" {
		t.Fatal("memberships")
	}
	if (tables.Invitation{}).TableName() != "invitations" {
		t.Fatal("invitations")
	}
}

func TestOrgModelRoundTrip(t *testing.T) {
	trial := time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("IST", 19800))
	o := &org.Organization{
		ID:          "org_1",
		Name:        "Acme",
		OwnerUserID: "usr_1",
		Status:      org.OrgActive,
		TrialEndsAt: &trial,
		CreatedAt:   trial,
		UpdatedAt:   trial,
	}
	got := orgFromModel(ptr(orgToModel(o)))
	if got.ID != o.ID || got.Name != o.Name || got.OwnerUserID != o.OwnerUserID {
		t.Fatalf("%+v", got)
	}
	if got.TrialEndsAt == nil || got.TrialEndsAt.Location() != time.UTC {
		t.Fatalf("trial tz=%v", got.TrialEndsAt)
	}
}

func TestMembershipStateMapsToStatus(t *testing.T) {
	now := time.Now().UTC()
	m := &org.Membership{
		ID:             "mbr_1",
		OrganizationID: "org_1",
		UserID:         "usr_1",
		Role:           org.RoleOwner,
		Status:         org.MemberActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	row := memberToModel(m)
	if row.State != org.MemberActive {
		t.Fatalf("state=%q", row.State)
	}
	got := memberFromModel(&row)
	if got.Status != org.MemberActive {
		t.Fatalf("status=%q", got.Status)
	}
}

func TestMapGormNotFound(t *testing.T) {
	if !errors.Is(mapGorm(gorm.ErrRecordNotFound), org.ErrNotFound) {
		t.Fatal("expected org.ErrNotFound")
	}
}

func TestUniquePendingInvite(t *testing.T) {
	if !isUnique(gorm.ErrDuplicatedKey) {
		t.Fatal("ErrDuplicatedKey")
	}
	err := utils.E(utils.CodeInvitationPending)
	if !isUnique(gorm.ErrDuplicatedKey) || err.Code != utils.CodeInvitationPending {
		t.Fatal(err)
	}
}

func ptr[T any](v T) *T { return &v }
