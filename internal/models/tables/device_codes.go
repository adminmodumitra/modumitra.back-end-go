package tables

import "time"

type DeviceCode struct {
	PublicID        string     `gorm:"column:public_id;primaryKey"`
	DeviceCode      string     `gorm:"column:device_code"`
	UserCode        string     `gorm:"column:user_code"`
	ClientID        string     `gorm:"column:client_id"`
	Provider        string     `gorm:"column:provider"`
	Status          string     `gorm:"column:status"`
	UserID          *string    `gorm:"column:user_id"`
	IntervalSeconds int        `gorm:"column:interval_seconds"`
	LastPollAt      *time.Time `gorm:"column:last_poll_at"`
	AuthorizedAt    *time.Time `gorm:"column:authorized_at"`
	ConsumedAt      *time.Time `gorm:"column:consumed_at"`
	ExpiresAt       time.Time  `gorm:"column:expires_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (DeviceCode) TableName() string { return "device_codes" }
