package tables

import "time"

// Session maps schema-v1 sessions.
type Session struct {
	PublicID        string     `gorm:"column:public_id;primaryKey"`
	UserID          string     `gorm:"column:user_id"`
	DeviceID        *string    `gorm:"column:device_id"`
	RefreshHash     string     `gorm:"column:refresh_token_hash"`
	FamilyID        string     `gorm:"column:family_id"`
	ExpiresAt       time.Time  `gorm:"column:expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	ReuseDetectedAt *time.Time `gorm:"column:reuse_detected_at"`
	ReplacedBy      *string    `gorm:"column:replaced_by_session_id"`
	DeviceName      *string    `gorm:"column:device_name"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
	LastSeenAt      time.Time  `gorm:"column:last_seen_at"`
}

func (Session) TableName() string { return "sessions" }
