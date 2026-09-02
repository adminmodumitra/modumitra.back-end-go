package tables

import "time"

type OAuthState struct {
	PublicID         string     `gorm:"column:public_id;primaryKey"`
	State            string     `gorm:"column:state"`
	Provider         string     `gorm:"column:provider"`
	Intent           string     `gorm:"column:intent"`
	RedirectURI      string     `gorm:"column:redirect_uri"`
	CodeChallenge    *string    `gorm:"column:code_challenge"`
	CodeVerifierHash *string    `gorm:"column:code_verifier_hash"`
	UserID           *string    `gorm:"column:user_id"`
	ConsumedAt       *time.Time `gorm:"column:consumed_at"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime:false"`
}

func (OAuthState) TableName() string { return "oauth_states" }
