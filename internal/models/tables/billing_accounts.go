package tables

import "time"

// BillingAccount maps schema-v1 billing_accounts.
// owner_id is polymorphic (usr_ or org_); there is no FK to users/organizations.
type BillingAccount struct {
	PublicID           string    `gorm:"column:public_id;primaryKey"`
	OwnerType          string    `gorm:"column:owner_type"`
	OwnerID            string    `gorm:"column:owner_id"`
	Provider           string    `gorm:"column:provider"`
	ProviderCustomerID *string   `gorm:"column:provider_customer_id"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (BillingAccount) TableName() string { return "billing_accounts" }
