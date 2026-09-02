package tables

import "time"

type BillingEvent struct {
	PublicID        string         `gorm:"column:public_id;primaryKey"`
	Provider        string         `gorm:"column:provider"`
	ProviderEventID string         `gorm:"column:provider_event_id"`
	Type            string         `gorm:"column:type"`
	OrganizationID  *string        `gorm:"column:organization_id"`
	SubscriptionID  *string        `gorm:"column:subscription_id"`
	Status          string         `gorm:"column:status"`
	PayloadDigest   string         `gorm:"column:payload_digest"`
	Summary         map[string]any `gorm:"column:summary;serializer:json"`
	Error           *string        `gorm:"column:error"`
	ProcessedAt     *time.Time     `gorm:"column:processed_at"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime:false"`
}

func (BillingEvent) TableName() string { return "billing_events" }
