package postgres

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
)

func TestTableNames(t *testing.T) {
	if (DeviceModel{}).TableName() != "devices" {
		t.Fatal("devices")
	}
	if (EntitlementModel{}).TableName() != "entitlements" {
		t.Fatal("entitlements")
	}
}

func TestDeviceModelRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("IST", 19800))
	d := &device.Device{
		ID:         "dev_1",
		UserID:     "usr_1",
		Name:       "Studio iMac",
		Platform:   device.PlatformMacOS,
		Status:     device.StatusActive,
		LastSeenAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	got := deviceFromModel(ptr(deviceToModel(d)))
	if got.ID != d.ID || got.Name != d.Name || got.Platform != d.Platform {
		t.Fatalf("%+v", got)
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf("tz=%v", got.CreatedAt.Location())
	}
}

func TestEntitlementModelRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	orgID := "org_1"
	devID := "dev_1"
	e := &device.Entitlement{
		ID:                 "ent_1",
		UserID:             "usr_1",
		OrganizationID:     &orgID,
		DeviceID:           &devID,
		Plan:               device.PlanTrial,
		SubscriptionStatus: device.SubTrialing,
		Capabilities: device.Capabilities{
			CloudProjects: true,
			MaxSeats:      3,
			MaxDevices:    3,
		},
		IssuedAt:      now,
		RefreshAfter:  now.Add(12 * time.Hour),
		ExpiresAt:     now.Add(24 * time.Hour),
		SchemaVersion: 1,
		KeyID:         device.KeyIDV1,
		Signature:     "abc",
		CreatedAt:     now,
	}
	got := entitlementFromModel(ptr(entitlementToModel(e)))
	if got.ID != e.ID || got.Plan != e.Plan || got.KeyID != e.KeyID {
		t.Fatalf("%+v", got)
	}
	if got.OrganizationID == nil || *got.OrganizationID != orgID {
		t.Fatalf("org=%v", got.OrganizationID)
	}
	if got.Capabilities.MaxSeats != 3 || !got.Capabilities.CloudProjects {
		t.Fatalf("caps=%+v", got.Capabilities)
	}
}

func TestMapGormNotFound(t *testing.T) {
	if !errors.Is(mapGorm(gorm.ErrRecordNotFound), device.ErrNotFound) {
		t.Fatal("expected device.ErrNotFound")
	}
}

func ptr[T any](v T) *T { return &v }
