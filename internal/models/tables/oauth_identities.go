package tables

import "time"

type OAuthIdentity struct {
	PublicID                string    `gorm:"column:public_id;primaryKey"`
	UserID                  string    `gorm:"column:user_id"`
	Provider                string    `gorm:"column:provider"`
	ProviderUserID          string    `gorm:"column:provider_user_id"`
	Email                   *string   `gorm:"column:email"`
	EmailVerifiedByProvider bool      `gorm:"column:email_verified_by_provider"`
	LinkedAt                time.Time `gorm:"column:linked_at"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (OAuthIdentity) TableName() string { return "oauth_identities" }
