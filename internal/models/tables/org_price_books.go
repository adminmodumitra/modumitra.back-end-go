package tables

import "time"

// OrgPriceBook maps org_price_books: one price book (rates JSON) per
// organization, served to the engine proxy instead of bundled rates.
type OrgPriceBook struct {
	OrgID     string    `gorm:"column:org_id;primaryKey"`
	Book      string    `gorm:"column:book"` // JSONB object
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	UpdatedBy string    `gorm:"column:updated_by"`
}

func (OrgPriceBook) TableName() string { return "org_price_books" }
