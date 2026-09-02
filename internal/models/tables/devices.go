package tables

import "time"

type Device struct {
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

func (Device) TableName() string { return "devices" }
