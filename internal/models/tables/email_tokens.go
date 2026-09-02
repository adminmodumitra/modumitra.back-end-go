package tables

import "time"

// EmailToken maps schema-v1 email_tokens.
type EmailToken struct {
	PublicID   string     `gorm:"column:public_id;primaryKey"`
	UserID     string     `gorm:"column:user_id"`
	Purpose    string     `gorm:"column:purpose"`
	TokenHash  string     `gorm:"column:token_hash"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime:false"`
}

func (EmailToken) TableName() string { return "email_tokens" }
