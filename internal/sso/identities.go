package sso

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) ListIdentities(ctx context.Context, userID string) ([]Identity, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	items, err := s.identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return items, nil
}

func (s *Service) Link(ctx context.Context, userID, provider, code, state string) (*Identity, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(state) == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	p, err := s.provider(provider)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	st, err := s.states.Consume(ctx, state, provider, now)
	if err != nil {
		return nil, utils.E(utils.CodeInvalidState)
	}
	if st.Intent != IntentLink || st.UserID != userID {
		return nil, utils.E(utils.CodeInvalidState)
	}
	verifier := DeriveVerifier(s.cfg.PKCESecret, st.ID)
	if st.CodeVerifierHash != "" && HashVerifier(verifier) != st.CodeVerifierHash {
		return nil, utils.E(utils.CodeInvalidState)
	}
	profile, err := p.Exchange(ctx, code, st.RedirectURI, verifier)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if strings.TrimSpace(profile.ProviderUserID) == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	if other, err := s.identities.GetByProviderSubject(ctx, provider, profile.ProviderUserID); err == nil {
		if other.UserID != userID {
			return nil, utils.E(utils.CodeIdentityAlreadyLinked)
		}
		return nil, utils.E(utils.CodeProviderAlreadyLinked)
	} else if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if _, err := s.identities.GetByUserProvider(ctx, userID, provider); err == nil {
		return nil, utils.E(utils.CodeProviderAlreadyLinked)
	} else if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
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
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.oauth_link", "user_id", userID, "provider", provider)
	return id, nil
}

func (s *Service) Unlink(ctx context.Context, userID, provider string) error {
	if userID == "" {
		return utils.E(utils.CodeUnauthenticated)
	}
	if _, err := s.provider(provider); err != nil {
		return err
	}
	u, err := s.accounts.UserByID(ctx, userID)
	if err != nil {
		return mapRepoErr(err)
	}
	items, err := s.identities.ListByUser(ctx, userID)
	if err != nil {
		return mapRepoErr(err)
	}
	var found *Identity
	for i := range items {
		if items[i].Provider == provider {
			found = &items[i]
			break
		}
	}
	if found == nil {
		return utils.E(utils.CodeNotFound)
	}
	hasPassword := u.PasswordHash != ""
	if !hasPassword && len(items) <= 1 {
		return utils.E(utils.CodeLastLoginMethod)
	}
	if err := s.identities.DeleteByUserProvider(ctx, userID, provider); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.oauth_unlink", "user_id", userID, "provider", provider)
	return nil
}
