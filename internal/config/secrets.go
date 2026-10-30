package config

import (
	"fmt"
	"net/url"
	"strings"
)

// minSecretLen is the shortest signing secret accepted outside local dev.
const minSecretLen = 32

// placeholderMarkers appear in the dev defaults shipped in git; a hosted API
// signing tokens with one of them can be forged by anyone who reads the repo.
var placeholderMarkers = []string{"local-dev-only", "change-me", "placeholder", "test-secret"}

// localDBHosts are the database hosts used by local dev and docker-compose.
var localDBHosts = map[string]bool{"": true, "localhost": true, "127.0.0.1": true, "::1": true, "postgres": true}

// RequiresStrongSecrets is true in production, and whenever the SQL database is
// not local (a hosted Neon/Render database with APP_ENV left unset still counts).
func (c Config) RequiresStrongSecrets() bool {
	switch strings.ToLower(appEnv()) {
	case "production", "prod":
		return true
	}
	return !isLocalDSN(c.DatabaseURL)
}

// ValidateSecrets refuses empty, short or dev-placeholder signing secrets when
// RequiresStrongSecrets. Local dev keeps working with the shipped defaults.
func (c Config) ValidateSecrets() error {
	if !c.RequiresStrongSecrets() {
		return nil
	}
	checks := []struct{ name, value string }{
		{"JWT_SECRET", c.JWTSecret},
		{"ENTITLEMENT_SIGNING_KEY", c.EntitlementSigningKey},
	}
	if c.Session.Enabled() {
		checks = append(checks, struct{ name, value string }{"SESSION_GATE_KEY", c.Session.GateKey})
	}
	for _, s := range checks {
		if reason := weakSecret(s.value); reason != "" {
			return fmt.Errorf("%s %s; set a random value of at least %d characters (openssl rand -base64 48)", s.name, reason, minSecretLen)
		}
	}
	if c.Session.Enabled() && strings.TrimSpace(c.Session.GateKey) == strings.TrimSpace(c.JWTSecret) {
		return fmt.Errorf("SESSION_GATE_KEY must differ from JWT_SECRET: the edge gate holds it")
	}
	return nil
}

func weakSecret(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "is empty"
	}
	lower := strings.ToLower(v)
	for _, m := range placeholderMarkers {
		if strings.Contains(lower, m) {
			return "is a dev placeholder"
		}
	}
	if len(v) < minSecretLen {
		return "is too short"
	}
	return ""
}

// isLocalDSN reports whether a URL-form DSN points at a local host. Key=value
// DSNs are treated as local so they never block local dev.
func isLocalDSN(dsn string) bool {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" || !strings.Contains(dsn, "://") {
		return true
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return true
	}
	return localDBHosts[strings.ToLower(u.Hostname())]
}
