package postgres

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
)

func userToModel(u *identity.User) tables.User {
	created := u.CreatedAt.UTC()
	pwdAt := utcPtr(u.PasswordUpdatedAt)
	if pwdAt == nil {
		pwdAt = &created
	}
	termsAt := u.TermsAcceptedAt
	if termsAt.IsZero() {
		termsAt = created
	} else {
		termsAt = termsAt.UTC()
	}
	return tables.User{
		PublicID:          u.ID,
		Email:             u.Email,
		EmailNormalized:   u.EmailNormalized,
		EmailVerifiedAt:   utcPtr(u.EmailVerifiedAt),
		PasswordHash:      u.PasswordHash,
		PasswordUpdatedAt: pwdAt,
		DisplayName:       u.DisplayName,
		TermsVersion:      u.TermsVersion,
		PrivacyVersion:    u.PrivacyVersion,
		TermsAcceptedAt:   termsAt,
		Status:            u.Status,
		LastLoginAt:       utcPtr(u.LastLoginAt),
		CreatedAt:         created,
		UpdatedAt:         u.UpdatedAt.UTC(),
	}
}

func userFromModel(m *tables.User) *identity.User {
	return &identity.User{
		ID:                m.PublicID,
		Email:             m.Email,
		EmailNormalized:   m.EmailNormalized,
		EmailVerifiedAt:   utcPtr(m.EmailVerifiedAt),
		PasswordHash:      m.PasswordHash,
		PasswordUpdatedAt: utcPtr(m.PasswordUpdatedAt),
		DisplayName:       m.DisplayName,
		TermsVersion:      m.TermsVersion,
		PrivacyVersion:    m.PrivacyVersion,
		TermsAcceptedAt:   m.TermsAcceptedAt.UTC(),
		Status:            m.Status,
		LastLoginAt:       utcPtr(m.LastLoginAt),
		CreatedAt:         m.CreatedAt.UTC(),
		UpdatedAt:         m.UpdatedAt.UTC(),
	}
}

func sessionToModel(s *identity.Session) tables.Session {
	return tables.Session{
		PublicID:        s.ID,
		UserID:          s.UserID,
		DeviceID:        nil, // insert path historically never writes device_id
		RefreshHash:     s.RefreshHash,
		FamilyID:        s.FamilyID,
		ExpiresAt:       s.ExpiresAt.UTC(),
		RevokedAt:       utcPtr(s.RevokedAt),
		ReuseDetectedAt: utcPtr(s.ReuseDetectedAt),
		ReplacedBy:      s.ReplacedBy,
		DeviceName:      emptyToNil(s.DeviceName),
		CreatedAt:       s.CreatedAt.UTC(),
		UpdatedAt:       s.UpdatedAt.UTC(),
		LastSeenAt:      s.LastSeenAt.UTC(),
	}
}

func sessionFromModel(m *tables.Session) *identity.Session {
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

func tokenToModel(t *identity.EmailToken) tables.EmailToken {
	return tables.EmailToken{
		PublicID:   t.ID,
		UserID:     t.UserID,
		Purpose:    t.Purpose,
		TokenHash:  t.TokenHash,
		ExpiresAt:  t.ExpiresAt.UTC(),
		ConsumedAt: utcPtr(t.ConsumedAt),
		CreatedAt:  t.CreatedAt.UTC(),
	}
}

func tokenFromModel(m *tables.EmailToken) *identity.EmailToken {
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
