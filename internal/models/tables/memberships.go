package tables

import "time"

// Membership maps schema-v1 memberships. Column "state" is API "status".
type Membership struct {
	PublicID       string     `gorm:"column:public_id;primaryKey"`
	OrganizationID string     `gorm:"column:organization_id"`
	UserID         string     `gorm:"column:user_id"`
	Role           string     `gorm:"column:role"`
	State          string     `gorm:"column:state"`
	InvitedAt      *time.Time `gorm:"column:invited_at"`
	ActivatedAt    *time.Time `gorm:"column:activated_at"`
	SuspendedAt    *time.Time `gorm:"column:suspended_at"`
	RemovedAt      *time.Time `gorm:"column:removed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Membership) TableName() string { return "memberships" }
