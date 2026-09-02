package sso

import (
	"context"
	"net/url"
	"sync"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// FakeProvider maps authorization_code → Profile. No network.
type FakeProvider struct {
	id       string
	label    string
	device   bool
	authBase string
	mu       sync.Mutex
	codes    map[string]Profile
}

func NewFakeProvider(id, label string, device bool) *FakeProvider {
	base := "https://example.invalid/oauth/authorize"
	switch id {
	case ProviderGoogle:
		base = "https://accounts.google.com/o/oauth2/v2/auth"
	case ProviderFacebook:
		base = "https://www.facebook.com/v19.0/dialog/oauth"
	case ProviderX:
		base = "https://twitter.com/i/oauth2/authorize"
	}
	return &FakeProvider{
		id:       id,
		label:    label,
		device:   device,
		authBase: base,
		codes:    map[string]Profile{},
	}
}

func (f *FakeProvider) ID() string                { return f.id }
func (f *FakeProvider) Label() string             { return f.label }
func (f *FakeProvider) DeviceCodeSupported() bool { return f.device }

func (f *FakeProvider) Set(code string, p Profile) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[code] = p
}

func (f *FakeProvider) AuthCodeURL(state, redirectURI, codeChallenge string) string {
	u, err := url.Parse(f.authBase)
	if err != nil {
		return f.authBase
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge_method", "S256")
	if codeChallenge != "" {
		q.Set("code_challenge", codeChallenge)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (f *FakeProvider) Exchange(_ context.Context, code, _, _ string) (Profile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.codes[code]
	if !ok {
		return Profile{}, utils.E(utils.CodeInvalidToken)
	}
	return p, nil
}

var _ Provider = (*FakeProvider)(nil)
