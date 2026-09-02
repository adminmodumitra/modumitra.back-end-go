package tables

import "time"

// User maps schema-v1 users. Not a domain type; no AutoMigrate.
type User struct {
	PublicID          string     `gorm:"column:public_id;primaryKey"`
	Email             string     `gorm:"column:email"`
	EmailNormalized   string     `gorm:"column:email_normalized"`
	EmailVerifiedAt   *time.Time `gorm:"column:email_verified_at"`
	PasswordHash      string     `gorm:"column:password_hash"`
	PasswordUpdatedAt *time.Time `gorm:"column:password_updated_at"`
	DisplayName       string     `gorm:"column:display_name"`
	TermsVersion      string     `gorm:"column:terms_version"`
	PrivacyVersion    string     `gorm:"column:privacy_version"`
	TermsAcceptedAt   time.Time  `gorm:"column:terms_accepted_at"`
	Status            string     `gorm:"column:status"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (User) TableName() string { return "users" }
