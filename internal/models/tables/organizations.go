package tables

import "time"

// Organization maps schema-v1 organizations.
type Organization struct {
	PublicID           string     `gorm:"column:public_id;primaryKey"`
	Name               string     `gorm:"column:name"`
	OwnerUserID        string     `gorm:"column:owner_user_id"`
	Status             string     `gorm:"column:status"`
	TrialEndsAt        *time.Time `gorm:"column:trial_ends_at"`
	Provider           *string    `gorm:"column:provider"`
	ProviderCustomerID *string    `gorm:"column:provider_customer_id"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Organization) TableName() string { return "organizations" }
