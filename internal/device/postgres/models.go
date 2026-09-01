package postgres

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
)

type DeviceModel struct {
	PublicID          string     `gorm:"column:public_id;primaryKey"`
	UserID            string     `gorm:"column:user_id"`
	Name              string     `gorm:"column:name"`
	Platform          string     `gorm:"column:platform"`
	Status            string     `gorm:"column:status"`
	AppVersion        *string    `gorm:"column:app_version"`
	LastSeenAt        time.Time  `gorm:"column:last_seen_at;autoCreateTime:false;autoUpdateTime:false"`
	DeactivatedAt     *time.Time `gorm:"column:deactivated_at"`
	DeactivatedReason *string    `gorm:"column:deactivated_reason"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (DeviceModel) TableName() string { return "devices" }

func deviceToModel(d *device.Device) DeviceModel {
	return DeviceModel{
		PublicID:          d.ID,
		UserID:            d.UserID,
		Name:              d.Name,
		Platform:          d.Platform,
		Status:            d.Status,
		AppVersion:        emptyToNil(d.AppVersion),
		LastSeenAt:        d.LastSeenAt.UTC(),
		DeactivatedAt:     utcPtr(d.DeactivatedAt),
		DeactivatedReason: emptyToNil(d.DeactivatedReason),
		CreatedAt:         d.CreatedAt.UTC(),
		UpdatedAt:         d.UpdatedAt.UTC(),
	}
}

func deviceFromModel(m *DeviceModel) *device.Device {
	return &device.Device{
		ID:                m.PublicID,
		UserID:            m.UserID,
		Name:              m.Name,
		Platform:          m.Platform,
		Status:            m.Status,
		AppVersion:        deref(m.AppVersion),
		LastSeenAt:        m.LastSeenAt.UTC(),
		DeactivatedAt:     utcPtr(m.DeactivatedAt),
		DeactivatedReason: deref(m.DeactivatedReason),
		CreatedAt:         m.CreatedAt.UTC(),
		UpdatedAt:         m.UpdatedAt.UTC(),
	}
}

type EntitlementModel struct {
	PublicID           string              `gorm:"column:public_id;primaryKey"`
	UserID             string              `gorm:"column:user_id"`
	OrganizationID     *string             `gorm:"column:organization_id"`
	DeviceID           *string             `gorm:"column:device_id"`
	PlanID             *string             `gorm:"column:plan_id"`
	Plan               string              `gorm:"column:plan"`
	SubscriptionStatus string              `gorm:"column:subscription_status"`
	Capabilities       device.Capabilities `gorm:"column:capabilities;serializer:json"`
	IssuedAt           time.Time           `gorm:"column:issued_at;autoCreateTime:false"`
	RefreshAfter       time.Time           `gorm:"column:refresh_after"`
	ExpiresAt          time.Time           `gorm:"column:expires_at"`
	GraceUntil         *time.Time          `gorm:"column:grace_until"`
	SchemaVersion      int                 `gorm:"column:schema_version"`
	KeyID              string              `gorm:"column:key_id"`
	Signature          string              `gorm:"column:signature"`
	CreatedAt          time.Time           `gorm:"column:created_at;autoCreateTime:false"`
}

func (EntitlementModel) TableName() string { return "entitlements" }

func entitlementToModel(e *device.Entitlement) EntitlementModel {
	return EntitlementModel{
		PublicID:           e.ID,
		UserID:             e.UserID,
		OrganizationID:     e.OrganizationID,
		DeviceID:           e.DeviceID,
		PlanID:             e.PlanID,
		Plan:               e.Plan,
		SubscriptionStatus: e.SubscriptionStatus,
		Capabilities:       e.Capabilities,
		IssuedAt:           e.IssuedAt.UTC(),
		RefreshAfter:       e.RefreshAfter.UTC(),
		ExpiresAt:          e.ExpiresAt.UTC(),
		GraceUntil:         utcPtr(e.GraceUntil),
		SchemaVersion:      e.SchemaVersion,
		KeyID:              e.KeyID,
		Signature:          e.Signature,
		CreatedAt:          e.CreatedAt.UTC(),
	}
}

func entitlementFromModel(m *EntitlementModel) *device.Entitlement {
	return &device.Entitlement{
		ID:                 m.PublicID,
		UserID:             m.UserID,
		OrganizationID:     m.OrganizationID,
		DeviceID:           m.DeviceID,
		PlanID:             m.PlanID,
		Plan:               m.Plan,
		SubscriptionStatus: m.SubscriptionStatus,
		Capabilities:       m.Capabilities,
		IssuedAt:           m.IssuedAt.UTC(),
		RefreshAfter:       m.RefreshAfter.UTC(),
		ExpiresAt:          m.ExpiresAt.UTC(),
		GraceUntil:         utcPtr(m.GraceUntil),
		SchemaVersion:      m.SchemaVersion,
		KeyID:              m.KeyID,
		Signature:          m.Signature,
		CreatedAt:          m.CreatedAt.UTC(),
	}
}
