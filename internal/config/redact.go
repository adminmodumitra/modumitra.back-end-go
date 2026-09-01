package config

import (
	"errors"
	"net/url"
	"strings"
)

// RedactDSN returns a log-safe form of a DSN, URI, or URL: userinfo is stripped
// so passwords never appear in logs. Non-URL values are returned unchanged when
// they contain no userinfo marker.
func RedactDSN(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return redactOpaque(raw)
	}
	if u.User != nil {
		u.User = url.User("redacted")
	}
	return u.String()
}

func redactOpaque(raw string) string {
	// host:port or similar — nothing to redact.
	if !strings.Contains(raw, "@") {
		return raw
	}
	// user:pass@host without a scheme
	at := strings.LastIndex(raw, "@")
	return "redacted@" + raw[at+1:]
}

// MaskSecrets replaces known secret substrings (raw DSN, password, URI) in err.
func MaskSecrets(err error, secrets ...string) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	for _, s := range secrets {
		if s == "" {
			continue
		}
		if strings.Contains(msg, s) {
			msg = strings.ReplaceAll(msg, s, RedactDSN(s))
		}
		u, perr := url.Parse(s)
		if perr == nil && u.User != nil {
			if pw, ok := u.User.Password(); ok && pw != "" && strings.Contains(msg, pw) {
				msg = strings.ReplaceAll(msg, pw, "*****")
			}
		}
	}
	if msg == err.Error() {
		return err
	}
	return errors.New(msg)
}
