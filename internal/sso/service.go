package sso

import (
	"errors"
	"net/url"
	"strings"

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
