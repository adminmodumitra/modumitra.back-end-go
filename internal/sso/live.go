package sso

import (
	"context"
	"io"
	"net/http"

	"golang.org/x/oauth2"

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

var _ Provider = (*liveProvider)(nil)
