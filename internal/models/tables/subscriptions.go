package tables

import "time"

type Subscription struct {
	PublicID               string     `gorm:"column:public_id;primaryKey"`
	OrganizationID         string     `gorm:"column:organization_id"`
	PlanID                 string     `gorm:"column:plan_id"`
	PlanCode               string     `gorm:"column:plan_code"`
	Status                 string     `gorm:"column:status"`
	Provider               string     `gorm:"column:provider"`
	ProviderSubscriptionID *string    `gorm:"column:provider_subscription_id"`
	SeatsIncluded          int        `gorm:"column:seats_included"`
	TrialEndsAt            *time.Time `gorm:"column:trial_ends_at"`
	CurrentPeriodStart     time.Time  `gorm:"column:current_period_start"`
	CurrentPeriodEnd       time.Time  `gorm:"column:current_period_end"`
	CancelAtPeriodEnd      bool       `gorm:"column:cancel_at_period_end"`
	CanceledAt             *time.Time `gorm:"column:canceled_at"`
	GraceUntil             *time.Time `gorm:"column:grace_until"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Subscription) TableName() string { return "subscriptions" }
