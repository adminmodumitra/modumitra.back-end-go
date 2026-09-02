package tables

import "time"

type Plan struct {
	PublicID          string           `gorm:"column:public_id;primaryKey"`
	Code              string           `gorm:"column:code"`
	Name              string           `gorm:"column:name"`
	Interval          string           `gorm:"column:interval"`
	Currency          string           `gorm:"column:currency"`
	UnitAmountMinor   int              `gorm:"column:unit_amount_minor"`
	SeatsIncluded     int              `gorm:"column:seats_included"`
	MaxDevicesPerUser int              `gorm:"column:max_devices_per_user"`
	TrialDays         int              `gorm:"column:trial_days"`
	Sellable          bool             `gorm:"column:sellable"`
	Active            bool             `gorm:"column:active"`
	Capabilities      PlanCapabilities `gorm:"column:capabilities;serializer:json"`
	Provider          *string          `gorm:"column:provider"`
	ProviderPriceID   *string          `gorm:"column:provider_price_id"`
	SortOrder         int              `gorm:"column:sort_order"`
	CreatedAt         time.Time        `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time        `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Plan) TableName() string { return "plans" }
