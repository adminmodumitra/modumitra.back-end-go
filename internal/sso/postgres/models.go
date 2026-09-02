package postgres

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

func identityToModel(i *sso.Identity) tables.OAuthIdentity {
	return tables.OAuthIdentity{
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

func identityFromModel(m *tables.OAuthIdentity) *sso.Identity {
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

func stateToModel(s *sso.State) tables.OAuthState {
	return tables.OAuthState{
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

func stateFromModel(m *tables.OAuthState) *sso.State {
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

func deviceToModel(d *sso.DeviceCode) tables.DeviceCode {
	return tables.DeviceCode{
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

func deviceFromModel(m *tables.DeviceCode) *sso.DeviceCode {
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
