package sso

import (
	"encoding/json"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Live adapters talk to Google / Facebook / X. Tests must use FakeProvider.

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
