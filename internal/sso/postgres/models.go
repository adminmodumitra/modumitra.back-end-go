package postgres

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

type IdentityModel struct {
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

func (IdentityModel) TableName() string { return "oauth_identities" }

func identityToModel(i *sso.Identity) IdentityModel {
	return IdentityModel{
		PublicID:                i.ID,
		UserID:                  i.UserID,
		Provider:                i.Provider,
		ProviderUserID:          i.ProviderUserID,
		Email:                   emptyToNil(i.Email),
		EmailVerifiedByProvider: i.EmailVerified,
		LinkedAt:                i.LinkedAt.UTC(),
		CreatedAt:               i.CreatedAt.UTC(),
		UpdatedAt:               i.UpdatedAt.UTC(),
	}
}

func identityFromModel(m *IdentityModel) *sso.Identity {
	return &sso.Identity{
		ID:             m.PublicID,
		UserID:         m.UserID,
		Provider:       m.Provider,
		ProviderUserID: m.ProviderUserID,
		Email:          deref(m.Email),
		EmailVerified:  m.EmailVerifiedByProvider,
		LinkedAt:       m.LinkedAt.UTC(),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
	}
}

type StateModel struct {
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

func (StateModel) TableName() string { return "oauth_states" }

func stateToModel(s *sso.State) StateModel {
	return StateModel{
		PublicID:         s.ID,
		State:            s.State,
		Provider:         s.Provider,
		Intent:           s.Intent,
		RedirectURI:      s.RedirectURI,
		CodeChallenge:    emptyToNil(s.CodeChallenge),
		CodeVerifierHash: emptyToNil(s.CodeVerifierHash),
		UserID:           emptyToNil(s.UserID),
		ConsumedAt:       utcPtr(s.ConsumedAt),
		ExpiresAt:        s.ExpiresAt.UTC(),
		CreatedAt:        s.CreatedAt.UTC(),
	}
}

func stateFromModel(m *StateModel) *sso.State {
	return &sso.State{
		ID:               m.PublicID,
		State:            m.State,
		Provider:         m.Provider,
		Intent:           m.Intent,
		RedirectURI:      m.RedirectURI,
		CodeChallenge:    deref(m.CodeChallenge),
		CodeVerifierHash: deref(m.CodeVerifierHash),
		UserID:           deref(m.UserID),
		ConsumedAt:       utcPtr(m.ConsumedAt),
		ExpiresAt:        m.ExpiresAt.UTC(),
		CreatedAt:        m.CreatedAt.UTC(),
	}
}

type DeviceCodeModel struct {
	PublicID        string     `gorm:"column:public_id;primaryKey"`
	DeviceCode      string     `gorm:"column:device_code"`
	UserCode        string     `gorm:"column:user_code"`
	ClientID        string     `gorm:"column:client_id"`
	Provider        string     `gorm:"column:provider"`
	Status          string     `gorm:"column:status"`
	UserID          *string    `gorm:"column:user_id"`
	IntervalSeconds int        `gorm:"column:interval_seconds"`
	LastPollAt      *time.Time `gorm:"column:last_poll_at"`
	AuthorizedAt    *time.Time `gorm:"column:authorized_at"`
	ConsumedAt      *time.Time `gorm:"column:consumed_at"`
	ExpiresAt       time.Time  `gorm:"column:expires_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (DeviceCodeModel) TableName() string { return "device_codes" }

func deviceToModel(d *sso.DeviceCode) DeviceCodeModel {
	return DeviceCodeModel{
		PublicID:        d.ID,
		DeviceCode:      d.DeviceCode,
		UserCode:        d.UserCode,
		ClientID:        d.ClientID,
		Provider:        d.Provider,
		Status:          d.Status,
		UserID:          emptyToNil(d.UserID),
		IntervalSeconds: d.IntervalSeconds,
		LastPollAt:      utcPtr(d.LastPollAt),
		AuthorizedAt:    utcPtr(d.AuthorizedAt),
		ConsumedAt:      utcPtr(d.ConsumedAt),
		ExpiresAt:       d.ExpiresAt.UTC(),
		CreatedAt:       d.CreatedAt.UTC(),
		UpdatedAt:       d.UpdatedAt.UTC(),
	}
}

func deviceFromModel(m *DeviceCodeModel) *sso.DeviceCode {
	return &sso.DeviceCode{
		ID:              m.PublicID,
		DeviceCode:      m.DeviceCode,
		UserCode:        m.UserCode,
		ClientID:        m.ClientID,
		Provider:        m.Provider,
		Status:          m.Status,
		UserID:          deref(m.UserID),
		IntervalSeconds: m.IntervalSeconds,
		LastPollAt:      utcPtr(m.LastPollAt),
		AuthorizedAt:    utcPtr(m.AuthorizedAt),
		ConsumedAt:      utcPtr(m.ConsumedAt),
		ExpiresAt:       m.ExpiresAt.UTC(),
		CreatedAt:       m.CreatedAt.UTC(),
		UpdatedAt:       m.UpdatedAt.UTC(),
	}
}
