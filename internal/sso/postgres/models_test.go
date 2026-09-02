package postgres

import (
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

func TestTableNames(t *testing.T) {
	if (tables.OAuthIdentity{}).TableName() != "oauth_identities" {
		t.Fatal("oauth_identities")
	}
	if (tables.OAuthState{}).TableName() != "oauth_states" {
		t.Fatal("oauth_states")
	}
	if (tables.DeviceCode{}).TableName() != "device_codes" {
		t.Fatal("device_codes")
	}
}

func TestIdentityRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("IST", 19800))
	id := &sso.Identity{
		ID: "oid_abc", UserID: "usr_1", Provider: "google", ProviderUserID: "gid",
		Email: "ada@example.com", EmailVerified: true, LinkedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	got := identityFromModel(ptr(identityToModel(id)))
	if got.ID != id.ID || got.Email != id.Email || got.EmailVerified != true {
		t.Fatalf("%+v", got)
	}
	if got.LinkedAt.Location() != time.UTC {
		t.Fatalf("tz=%v", got.LinkedAt.Location())
	}
}

func TestStateEmptyUserID(t *testing.T) {
	now := time.Now().UTC()
	st := &sso.State{ID: "st_1", State: "st_1", Provider: "google", Intent: "login", RedirectURI: "cabinetstudio://oauth/callback", ExpiresAt: now, CreatedAt: now}
	m := stateToModel(st)
	if m.UserID != nil || m.CodeChallenge != nil {
		t.Fatalf("%+v", m)
	}
}

func ptr[T any](v T) *T { return &v }
