package sso

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) loginOrProvision(ctx context.Context, provider string, profile Profile) (*identity.User, error) {
	if strings.TrimSpace(profile.ProviderUserID) == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	existing, err := s.identities.GetByProviderSubject(ctx, provider, profile.ProviderUserID)
	if err == nil {
		u, uerr := s.accounts.UserByID(ctx, existing.UserID)
		if uerr != nil {
			return nil, mapRepoErr(uerr)
		}
		return u, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	email := strings.TrimSpace(profile.Email)
	if email == "" {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "email"})
	}
	if taken, terr := s.accounts.UserByEmail(ctx, email); terr == nil && taken != nil {
		return nil, utils.E(utils.CodeEmailTaken)
	} else if terr != nil && !errors.Is(terr, identity.ErrNotFound) {
		return nil, mapRepoErr(terr)
	}
	now := s.clock.Now()
	display := strings.TrimSpace(profile.DisplayName)
	if display == "" {
		display = email
		if i := strings.IndexByte(display, '@'); i > 0 {
			display = display[:i]
		}
	}
	if utf8.RuneCountInString(display) > 80 {
		display = string([]rune(display)[:80])
	}
	u := &identity.User{
		ID:              s.accounts.NewUserID(),
		Email:           email,
		EmailNormalized: strings.ToLower(email),
		PasswordHash:    "",
		DisplayName:     display,
		TermsVersion:    s.cfg.TermsVersion,
		PrivacyVersion:  s.cfg.PrivacyVersion,
		Status:          identity.StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if u.TermsVersion == "" {
		u.TermsVersion = s.accounts.TermsVersion()
	}
	if u.PrivacyVersion == "" {
		u.PrivacyVersion = s.accounts.PrivacyVersion()
	}
	if profile.EmailVerified {
		t := now
		u.EmailVerifiedAt = &t
	}
	if err := s.accounts.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	if err := s.linkIdentity(ctx, u.ID, provider, profile, now); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) linkIdentity(ctx context.Context, userID, provider string, profile Profile, now time.Time) error {
	id := &Identity{
		ID:             s.ids.Identity(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: profile.ProviderUserID,
		Email:          strings.TrimSpace(profile.Email),
		EmailVerified:  profile.EmailVerified,
		LinkedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.identities.Insert(ctx, id); err != nil {
		return mapRepoErr(err)
	}
	return nil
}
