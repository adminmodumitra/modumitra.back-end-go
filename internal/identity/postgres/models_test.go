package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func TestTableNames(t *testing.T) {
	if (tables.User{}).TableName() != "users" {
		t.Fatalf("users")
	}
	if (tables.Session{}).TableName() != "sessions" {
		t.Fatalf("sessions")
	}
	if (tables.EmailToken{}).TableName() != "email_tokens" {
		t.Fatalf("email_tokens")
	}
}

func TestUserModelRoundTrip(t *testing.T) {
	verified := time.Date(2026, 1, 2, 3, 4, 5, 6, time.FixedZone("IST", 19800))
	u := &identity.User{
		ID:              "usr_abc",
		Email:           "Ada@Example.com",
		EmailNormalized: "ada@example.com",
		EmailVerifiedAt: &verified,
		PasswordHash:    "argon2id$...",
		DisplayName:     "Ada",
		TermsVersion:    "1",
		PrivacyVersion:  "1",
		Status:          identity.StatusActive,
		CreatedAt:       verified,
		UpdatedAt:       verified,
	}
	got := userFromModel(ptr(userToModel(u)))
	if got.ID != u.ID || got.Email != u.Email || got.PasswordHash != u.PasswordHash {
		t.Fatalf("got=%+v", got)
	}
	if got.EmailVerifiedAt == nil || got.EmailVerifiedAt.Location() != time.UTC {
		t.Fatalf("verified tz=%v", got.EmailVerifiedAt)
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf("created tz=%v", got.CreatedAt.Location())
	}
}

func TestSessionModelEmptyDeviceName(t *testing.T) {
	now := time.Now().UTC()
	s := &identity.Session{
		ID:          "ses_1",
		UserID:      "usr_1",
		RefreshHash: "hash",
		FamilyID:    "fam_1",
		ExpiresAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastSeenAt:  now,
	}
	m := sessionToModel(s)
	if m.DeviceName != nil || m.DeviceID != nil {
		t.Fatalf("expected nil device fields: %+v", m)
	}
	got := sessionFromModel(&m)
	if got.DeviceName != "" {
		t.Fatalf("device name=%q", got.DeviceName)
	}
}

func TestMapGormNotFound(t *testing.T) {
	if !errors.Is(mapGorm(gorm.ErrRecordNotFound), identity.ErrNotFound) {
		t.Fatal("expected identity.ErrNotFound")
	}
}

func TestUniqueEmailMapsToEmailTaken(t *testing.T) {
	if !isUnique(gorm.ErrDuplicatedKey) {
		t.Fatal("ErrDuplicatedKey should be unique")
	}
	err := mapGorm(utils.E(utils.CodeEmailTaken))
	var ae *utils.Error
	if !errors.As(err, &ae) || ae.Code != utils.CodeEmailTaken {
		t.Fatalf("err=%v", err)
	}
}

func ptr[T any](v T) *T { return &v }
