package tables

import "time"

type Entitlement struct {
	PublicID           string             `gorm:"column:public_id;primaryKey"`
	UserID             string             `gorm:"column:user_id"`
	OrganizationID     *string            `gorm:"column:organization_id"`
	DeviceID           *string            `gorm:"column:device_id"`
	PlanID             *string            `gorm:"column:plan_id"`
	Plan               string             `gorm:"column:plan"`
	SubscriptionStatus string             `gorm:"column:subscription_status"`
	Capabilities       DeviceCapabilities `gorm:"column:capabilities;serializer:json"`
	IssuedAt           time.Time          `gorm:"column:issued_at;autoCreateTime:false"`
	RefreshAfter       time.Time          `gorm:"column:refresh_after"`
	ExpiresAt          time.Time          `gorm:"column:expires_at"`
	GraceUntil         *time.Time         `gorm:"column:grace_until"`
	SchemaVersion      int                `gorm:"column:schema_version"`
	KeyID              string             `gorm:"column:key_id"`
	Signature          string             `gorm:"column:signature"`
	CreatedAt          time.Time          `gorm:"column:created_at;autoCreateTime:false"`
}

func (Entitlement) TableName() string { return "entitlements" }
