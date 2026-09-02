package sso

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	identities IdentityRepo
	states     StateRepo
	devices    DeviceCodeRepo
	tx         TxRunner
	accounts   Accounts
	providers  map[string]Provider
	clock      Clock
	ids        IDGen
	cfg        Settings
	outbox     outbox.Writer
}

func NewService(
	identities IdentityRepo,
	states StateRepo,
	devices DeviceCodeRepo,
	tx TxRunner,
	accounts Accounts,
	providers map[string]Provider,
	clock Clock,
	ids IDGen,
	cfg Settings,
	ob outbox.Writer,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if cfg.StateTTL <= 0 {
		cfg.StateTTL = DefaultStateTTL
	}
	if cfg.DeviceTTL <= 0 {
		cfg.DeviceTTL = DefaultDeviceTTL
	}
	if cfg.DeviceInterval <= 0 {
		cfg.DeviceInterval = DefaultDeviceInterval
	}
	if cfg.DeviceClientID == "" {
		cfg.DeviceClientID = DesktopClientID
	}
	if cfg.VerificationURI == "" {
		cfg.VerificationURI = "https://app.cabinetstudio.example/device"
	}
	if providers == nil {
		providers = map[string]Provider{}
	}
	return &Service{
		identities: identities,
		states:     states,
		devices:    devices,
		tx:         tx,
		accounts:   accounts,
		providers:  providers,
		clock:      clock,
		ids:        ids,
		cfg:        cfg,
		outbox:     ob,
	}
}

func NewMemoryService(accounts Accounts, cfg Settings) (*Service, *Memory, map[string]*FakeProvider) {
	mem := NewMemory()
	fakes := map[string]*FakeProvider{
		ProviderGoogle:   NewFakeProvider(ProviderGoogle, "Google", true),
		ProviderFacebook: NewFakeProvider(ProviderFacebook, "Facebook", false),
		ProviderX:        NewFakeProvider(ProviderX, "X", false),
	}
	providers := map[string]Provider{}
	for k, v := range fakes {
		providers[k] = v
	}
	svc := NewService(mem.Identities(), mem.States(), mem.DeviceCodes(), mem, accounts, providers, SystemClock{}, NewIDGen(), cfg, outbox.NewMemoryWriter())
	return svc, mem, fakes
}

func (s *Service) provider(id string) (Provider, error) {
	if !KnownProvider(id) {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "provider"})
	}
	p, ok := s.providers[id]
	if !ok || p == nil {
		return nil, utils.E(utils.CodeUnsupportedProvider)
	}
	return p, nil
}

func (s *Service) ListProviders() []ProviderInfo {
	order := []string{ProviderGoogle, ProviderFacebook, ProviderX}
	var out []ProviderInfo
	for _, id := range order {
		p, ok := s.providers[id]
		if !ok || p == nil {
			continue
		}
		out = append(out, ProviderInfo{ID: p.ID(), Label: p.Label(), DeviceCodeSupported: p.DeviceCodeSupported()})
	}
	if out == nil {
		out = []ProviderInfo{}
	}
	return out
}

func allowedRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "cabinetstudio", "https":
		return true
	default:
		return false
	}
}

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

func (s *Service) CreateDeviceCode(ctx context.Context, clientID, provider string) (*DeviceCodeResult, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "client_id"})
	}
	if clientID != s.cfg.DeviceClientID {
		return nil, utils.E(utils.CodeInvalidClient)
	}
	p, err := s.provider(provider)
	if err != nil {
		return nil, err
	}
	if !p.DeviceCodeSupported() {
		return nil, utils.E(utils.CodeUnsupportedProvider)
	}
	now := s.clock.Now()
	id := s.ids.Device()
	userCode := s.ids.UserCode()
	d := &DeviceCode{
		ID:              id,
		DeviceCode:      id,
		UserCode:        userCode,
		ClientID:        clientID,
		Provider:        provider,
		Status:          StatusPending,
		IntervalSeconds: s.cfg.DeviceInterval,
		ExpiresAt:       now.Add(s.cfg.DeviceTTL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.devices.Insert(ctx, d); err != nil {
		return nil, mapRepoErr(err)
	}
	complete := s.cfg.VerificationURI
	if strings.Contains(complete, "?") {
		complete += "&user_code=" + url.QueryEscape(userCode)
	} else {
		complete += "?user_code=" + url.QueryEscape(userCode)
	}
	return &DeviceCodeResult{
		DeviceCode:              d.DeviceCode,
		UserCode:                d.UserCode,
		VerificationURI:         s.cfg.VerificationURI,
		VerificationURIComplete: complete,
		ExpiresIn:               int(s.cfg.DeviceTTL.Seconds()),
		Interval:                d.IntervalSeconds,
	}, nil
}

func (s *Service) PollDeviceToken(ctx context.Context, deviceCode string) (*identity.User, *identity.TokenPair, error) {
	if strings.TrimSpace(deviceCode) == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	d, err := s.devices.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidGrant)
		}
		return nil, nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if d.Status == StatusConsumed {
		return nil, nil, utils.E(utils.CodeInvalidGrant)
	}
	if d.Status == StatusDenied {
		return nil, nil, utils.E(utils.CodeAccessDenied)
	}
	if d.Status == StatusExpired || !d.ExpiresAt.After(now) {
		if d.Status != StatusExpired {
			d.Status = StatusExpired
			d.UpdatedAt = now
			_ = s.devices.Update(ctx, d)
		}
		return nil, nil, utils.E(utils.CodeExpiredToken)
	}
	if d.Status == StatusPending {
		if d.LastPollAt != nil && now.Sub(*d.LastPollAt) < time.Duration(d.IntervalSeconds)*time.Second {
			return nil, nil, utils.E(utils.CodeSlowDown).WithDetails(map[string]any{"interval": d.IntervalSeconds})
		}
		t := now
		d.LastPollAt = &t
		d.UpdatedAt = now
		_ = s.devices.Update(ctx, d)
		return nil, nil, utils.E(utils.CodeAuthorizationPending).WithDetails(map[string]any{"interval": d.IntervalSeconds})
	}
	if d.Status != StatusAuthorized || d.UserID == "" {
		return nil, nil, utils.E(utils.CodeAuthorizationPending).WithDetails(map[string]any{"interval": d.IntervalSeconds})
	}
	user, pair, err := s.accounts.LoginAs(ctx, d.UserID, "")
	if err != nil {
		return nil, nil, err
	}
	cons := now
	d.Status = StatusConsumed
	d.ConsumedAt = &cons
	d.UpdatedAt = now
	if err := s.devices.Update(ctx, d); err != nil {
		return nil, nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.device_login", "user_id", user.ID)
	return user, pair, nil
}

func (s *Service) AuthorizeDevice(ctx context.Context, userCode, userID string) error {
	if strings.TrimSpace(userCode) == "" || strings.TrimSpace(userID) == "" {
		return utils.E(utils.CodeValidation)
	}
	if _, err := s.accounts.UserByID(ctx, userID); err != nil {
		return mapRepoErr(err)
	}
	d, err := s.devices.GetByUserCode(ctx, userCode)
	if err != nil {
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	if d.Status != StatusPending || !d.ExpiresAt.After(now) {
		return utils.E(utils.CodeInvalidGrant)
	}
	t := now
	d.Status = StatusAuthorized
	d.UserID = userID
	d.AuthorizedAt = &t
	d.UpdatedAt = now
	return mapRepoErr(s.devices.Update(ctx, d))
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	if errors.Is(err, ErrInvalidState) {
		return utils.E(utils.CodeInvalidState)
	}
	if errors.Is(err, identity.ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}
