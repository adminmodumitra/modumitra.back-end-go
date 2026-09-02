package config

import (
	"strings"
)

// OAuthApp is a single provider's confidential client credentials.
type OAuthApp struct {
	ClientID     string
	ClientSecret string
}

func (a OAuthApp) Enabled() bool {
	return strings.TrimSpace(a.ClientID) != "" && strings.TrimSpace(a.ClientSecret) != ""
}

// LiveOAuthProviders is the subset of google/facebook/x that have both client id and secret.
func (c Config) LiveOAuthProviders() []string {
	var out []string
	if c.OAuthGoogle.Enabled() {
		out = append(out, "google")
	}
	if c.OAuthFacebook.Enabled() {
		out = append(out, "facebook")
	}
	if c.OAuthX.Enabled() {
		out = append(out, "x")
	}
	return out
}
