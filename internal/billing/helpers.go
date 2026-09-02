package billing

import (
	"strings"
	"time"
)

func allowedRedirect(u string) bool {
	u = strings.TrimSpace(u)
	return strings.HasPrefix(u, "cabinetstudio://") || strings.HasPrefix(u, "https://")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	u := t.UTC()
	return &u
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
