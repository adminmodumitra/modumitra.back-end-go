package postgres

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

// UserModel maps schema-v1 users. Not a domain type; no AutoMigrate.
type UserModel struct {
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

func (UserModel) TableName() string { return "users" }

func userToModel(u *identity.User) UserModel {
	return UserModel{
		PublicID:        u.ID,
		Email:           u.Email,
		EmailNormalized: u.EmailNormalized,
		EmailVerifiedAt: utcPtr(u.EmailVerifiedAt),
		PasswordHash:    u.PasswordHash,
		DisplayName:     u.DisplayName,
		TermsVersion:    u.TermsVersion,
		PrivacyVersion:  u.PrivacyVersion,
		Status:          u.Status,
		LastLoginAt:     utcPtr(u.LastLoginAt),
		CreatedAt:       u.CreatedAt.UTC(),
		UpdatedAt:       u.UpdatedAt.UTC(),
	}
}

func userFromModel(m *UserModel) *identity.User {
	return &identity.User{
		ID:              m.PublicID,
		Email:           m.Email,
		EmailNormalized: m.EmailNormalized,
		EmailVerifiedAt: utcPtr(m.EmailVerifiedAt),
		PasswordHash:    m.PasswordHash,
		DisplayName:     m.DisplayName,
		TermsVersion:    m.TermsVersion,
		PrivacyVersion:  m.PrivacyVersion,
		Status:          m.Status,
		LastLoginAt:     utcPtr(m.LastLoginAt),
		CreatedAt:       m.CreatedAt.UTC(),
		UpdatedAt:       m.UpdatedAt.UTC(),
	}
}

// SessionModel maps schema-v1 sessions.
type SessionModel struct {
	PublicID        string     `gorm:"column:public_id;primaryKey"`
	UserID          string     `gorm:"column:user_id"`
	DeviceID        *string    `gorm:"column:device_id"`
	RefreshHash     string     `gorm:"column:refresh_token_hash"`
	FamilyID        string     `gorm:"column:family_id"`
	ExpiresAt       time.Time  `gorm:"column:expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	ReuseDetectedAt *time.Time `gorm:"column:reuse_detected_at"`
	ReplacedBy      *string    `gorm:"column:replaced_by_session_id"`
	DeviceName      *string    `gorm:"column:device_name"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
	LastSeenAt      time.Time  `gorm:"column:last_seen_at"`
}

func (SessionModel) TableName() string { return "sessions" }

func sessionToModel(s *identity.Session) SessionModel {
	return SessionModel{
		PublicID:    s.ID,
		UserID:      s.UserID,
		DeviceID:    nil, // insert path historically never writes device_id
		RefreshHash: s.RefreshHash,
		FamilyID:    s.FamilyID,
		ExpiresAt:   s.ExpiresAt.UTC(),
		DeviceName:  emptyToNil(s.DeviceName),
		CreatedAt:   s.CreatedAt.UTC(),
		UpdatedAt:   s.UpdatedAt.UTC(),
		LastSeenAt:  s.LastSeenAt.UTC(),
	}
}

func sessionFromModel(m *SessionModel) *identity.Session {
	return &identity.Session{
		ID:              m.PublicID,
		UserID:          m.UserID,
		DeviceID:        m.DeviceID,
		RefreshHash:     m.RefreshHash,
		FamilyID:        m.FamilyID,
		ExpiresAt:       m.ExpiresAt.UTC(),
		RevokedAt:       utcPtr(m.RevokedAt),
		ReuseDetectedAt: utcPtr(m.ReuseDetectedAt),
		ReplacedBy:      m.ReplacedBy,
		DeviceName:      deref(m.DeviceName),
		CreatedAt:       m.CreatedAt.UTC(),
		UpdatedAt:       m.UpdatedAt.UTC(),
		LastSeenAt:      m.LastSeenAt.UTC(),
	}
}

// EmailTokenModel maps schema-v1 email_tokens.
type EmailTokenModel struct {
	PublicID   string     `gorm:"column:public_id;primaryKey"`
	UserID     string     `gorm:"column:user_id"`
	Purpose    string     `gorm:"column:purpose"`
	TokenHash  string     `gorm:"column:token_hash"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime:false"`
}

func (EmailTokenModel) TableName() string { return "email_tokens" }

func tokenToModel(t *identity.EmailToken) EmailTokenModel {
	return EmailTokenModel{
		PublicID:  t.ID,
		UserID:    t.UserID,
		Purpose:   t.Purpose,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt.UTC(),
		CreatedAt: t.CreatedAt.UTC(),
	}
}

func tokenFromModel(m *EmailTokenModel) *identity.EmailToken {
	return &identity.EmailToken{
		ID:         m.PublicID,
		UserID:     m.UserID,
		Purpose:    m.Purpose,
		TokenHash:  m.TokenHash,
		ExpiresAt:  m.ExpiresAt.UTC(),
		ConsumedAt: utcPtr(m.ConsumedAt),
		CreatedAt:  m.CreatedAt.UTC(),
	}
}
