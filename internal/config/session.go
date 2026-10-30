package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Session turns on browser cookie sessions for the subdomain setup
// (app., login., register. on one parent domain). Off when either field is
// empty: login then answers with bearer tokens only, as before.
type Session struct {
	// CookieDomain is the parent domain the access cookie is shared on,
	// e.g. ".modumitrasoftware.com". Env SESSION_COOKIE_DOMAIN.
	CookieDomain string
	// Origins may use cookie sessions: CORS with credentials, and the only
	// Origins accepted on cookie-authenticated writes. Never list www.
	// Env SESSION_ORIGINS, a JSON array like CORS_ORIGINS.
	Origins []string
	// GateKey signs the ms_gate cookie the app. edge gate checks. Its own key,
	// so the edge never holds JWT_SECRET. Env SESSION_GATE_KEY.
	GateKey string
}

func (s Session) Enabled() bool {
	return strings.TrimSpace(s.CookieDomain) != "" && len(s.Origins) > 0
}

func applySessionEnv(s *Session) {
	s.CookieDomain = getenv("SESSION_COOKIE_DOMAIN", s.CookieDomain)
	s.GateKey = getenv("SESSION_GATE_KEY", s.GateKey)
	if v := strings.TrimSpace(os.Getenv("SESSION_ORIGINS")); v != "" {
		var origins []string
		if err := json.Unmarshal([]byte(v), &origins); err == nil && origins != nil {
			s.Origins = origins
		}
	}
}
