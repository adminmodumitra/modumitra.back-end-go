package sso

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Start(ctx context.Context, in StartInput) (*StartResult, error) {
	p, err := s.provider(in.Provider)
	if err != nil {
		return nil, err
	}
	intent := strings.TrimSpace(in.Intent)
	if intent != IntentLogin && intent != IntentLink {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "intent"})
	}
	redirect := strings.TrimSpace(in.RedirectURI)
	if redirect == "" || !allowedRedirect(redirect) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	if intent == IntentLink {
		if strings.TrimSpace(in.ActorID) == "" {
			return nil, utils.E(utils.CodeUnauthenticated)
		}
	}
	now := s.clock.Now()
	id := s.ids.State()
	verifier := DeriveVerifier(s.cfg.PKCESecret, id)
	challenge := S256Challenge(verifier)
	st := &State{
		ID:               id,
		State:            id,
		Provider:         in.Provider,
		Intent:           intent,
		RedirectURI:      redirect,
		CodeChallenge:    challenge,
		CodeVerifierHash: HashVerifier(verifier),
		UserID:           strings.TrimSpace(in.ActorID),
		ExpiresAt:        now.Add(s.cfg.StateTTL),
		CreatedAt:        now,
	}
	if intent == IntentLogin {
		st.UserID = ""
	}
	if err := s.states.Insert(ctx, st); err != nil {
		return nil, mapRepoErr(err)
	}
	return &StartResult{
		AuthorizationURL: p.AuthCodeURL(st.State, redirect, challenge),
		State:            st.State,
	}, nil
}

func (s *Service) Callback(ctx context.Context, provider, code, state, providerErr string) (*identity.User, *identity.TokenPair, error) {
	if strings.TrimSpace(providerErr) != "" {
		return nil, nil, utils.E(utils.CodeOAuthDenied)
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(state) == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	p, err := s.provider(provider)
	if err != nil {
		return nil, nil, err
	}
	now := s.clock.Now()
	st, err := s.states.Consume(ctx, state, provider, now)
	if err != nil {
		return nil, nil, utils.E(utils.CodeInvalidState)
	}
	if st.Intent != IntentLogin {
		return nil, nil, utils.E(utils.CodeInvalidState)
	}
	verifier := DeriveVerifier(s.cfg.PKCESecret, st.ID)
	if st.CodeVerifierHash != "" && HashVerifier(verifier) != st.CodeVerifierHash {
		return nil, nil, utils.E(utils.CodeInvalidState)
	}
	profile, err := p.Exchange(ctx, code, st.RedirectURI, verifier)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	u, err := s.loginOrProvision(ctx, provider, profile)
	if err != nil {
		return nil, nil, err
	}
	user, pair, err := s.accounts.LoginAs(ctx, u.ID, "")
	if err != nil {
		return nil, nil, err
	}
	slog.InfoContext(ctx, "auth.oauth_login", "user_id", user.ID, "provider", provider)
	return user, pair, nil
}
