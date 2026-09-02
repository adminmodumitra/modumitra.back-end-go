package tables

import "time"

// Invitation maps schema-v1 invitations.
type Invitation struct {
	PublicID         string     `gorm:"column:public_id;primaryKey"`
	OrganizationID   string     `gorm:"column:organization_id"`
	Email            string     `gorm:"column:email"`
	Role             string     `gorm:"column:role"`
	Status           string     `gorm:"column:status"`
	TokenHash        string     `gorm:"column:token_hash"`
	InvitedByUserID  string     `gorm:"column:invited_by_user_id"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	AcceptedAt       *time.Time `gorm:"column:accepted_at"`
	AcceptedByUserID *string    `gorm:"column:accepted_by_user_id"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	RevokedByUserID  *string    `gorm:"column:revoked_by_user_id"`
	ResentAt         *time.Time `gorm:"column:resent_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Invitation) TableName() string { return "invitations" }
