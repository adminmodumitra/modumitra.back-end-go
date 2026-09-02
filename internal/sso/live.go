package sso

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Live adapters talk to Google / Facebook / X. Tests must use FakeProvider.

var (
	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
	}
	facebookEndpoint = oauth2.Endpoint{
		AuthURL:  "https://www.facebook.com/v19.0/dialog/oauth",
		TokenURL: "https://graph.facebook.com/v19.0/oauth/access_token",
	}
	xEndpoint = oauth2.Endpoint{
		AuthURL:  "https://twitter.com/i/oauth2/authorize",
		TokenURL: "https://api.twitter.com/2/oauth2/token",
	}
)

type liveProvider struct {
	id       string
	label    string
	device   bool
	conf     oauth2.Config
	userinfo string
}

func (p *liveProvider) ID() string                { return p.id }
func (p *liveProvider) Label() string             { return p.label }
func (p *liveProvider) DeviceCodeSupported() bool { return p.device }

func (p *liveProvider) AuthCodeURL(state, redirectURI, codeChallenge string) string {
	conf := p.conf
	conf.RedirectURL = redirectURI
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	if codeChallenge != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
	}
	return conf.AuthCodeURL(state, opts...)
}

func (p *liveProvider) Exchange(ctx context.Context, code, redirectURI, codeVerifier string) (Profile, error) {
	conf := p.conf
	conf.RedirectURL = redirectURI
	opts := []oauth2.AuthCodeOption{}
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
	tok, err := conf.Exchange(ctx, code, opts...)
	if err != nil {
		return Profile{}, utils.E(utils.CodeInvalidToken).With(err)
	}
	client := conf.Client(ctx, tok)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userinfo, nil)
	if err != nil {
		return Profile{}, utils.Generic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Profile{}, utils.Generic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Profile{}, utils.Generic(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Profile{}, utils.E(utils.CodeInvalidToken)
	}
	return parseLiveProfile(p.id, body)
}

func parseLiveProfile(id string, body []byte) (Profile, error) {
	switch id {
	case ProviderGoogle:
		var g struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
		}
		if err := json.Unmarshal(body, &g); err != nil || strings.TrimSpace(g.Sub) == "" {
			return Profile{}, utils.E(utils.CodeInvalidToken)
		}
		return Profile{ProviderUserID: g.Sub, Email: g.Email, EmailVerified: g.EmailVerified, DisplayName: g.Name}, nil
	case ProviderFacebook:
		var fb struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.Unmarshal(body, &fb); err != nil || strings.TrimSpace(fb.ID) == "" {
			return Profile{}, utils.E(utils.CodeInvalidToken)
		}
		return Profile{ProviderUserID: fb.ID, Email: fb.Email, EmailVerified: fb.Email != "", DisplayName: fb.Name}, nil
	case ProviderX:
		var x struct {
			Data struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				ConfirmedEmail string `json:"confirmed_email"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &x); err != nil || strings.TrimSpace(x.Data.ID) == "" {
			return Profile{}, utils.E(utils.CodeInvalidToken)
		}
		email := x.Data.ConfirmedEmail
		return Profile{ProviderUserID: x.Data.ID, Email: email, EmailVerified: email != "", DisplayName: x.Data.Name}, nil
	default:
		return Profile{}, utils.E(utils.CodeUnsupportedProvider)
	}
}

func newLive(id, label, clientID, clientSecret string, ep oauth2.Endpoint, scopes []string, userinfo string, device bool) *liveProvider {
	return &liveProvider{
		id:     id,
		label:  label,
		device: device,
		conf: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     ep,
			Scopes:       scopes,
		},
		userinfo: userinfo,
	}
}

// LiveProviders builds google/facebook/x adapters for credentials that are set.
// Empty client id or secret disables that provider. HTTP client timeout is short.
func LiveProviders(cfg config.Config) map[string]Provider {
	out := map[string]Provider{}
	if cfg.OAuthGoogle.Enabled() {
		out[ProviderGoogle] = newLive(ProviderGoogle, "Google", cfg.OAuthGoogle.ClientID, cfg.OAuthGoogle.ClientSecret,
			googleEndpoint, []string{"openid", "email", "profile"},
			"https://openidconnect.googleapis.com/v1/userinfo", true)
	}
	if cfg.OAuthFacebook.Enabled() {
		out[ProviderFacebook] = newLive(ProviderFacebook, "Facebook", cfg.OAuthFacebook.ClientID, cfg.OAuthFacebook.ClientSecret,
			facebookEndpoint, []string{"email", "public_profile"},
			"https://graph.facebook.com/me?fields=id,name,email", false)
	}
	if cfg.OAuthX.Enabled() {
		out[ProviderX] = newLive(ProviderX, "X", cfg.OAuthX.ClientID, cfg.OAuthX.ClientSecret,
			xEndpoint, []string{"users.read", "tweet.read", "users.email", "offline.access"},
			"https://api.twitter.com/2/users/me?user.fields=confirmed_email,name", false)
	}
	return out
}

var _ Provider = (*liveProvider)(nil)
