package config

import "strings"

// liveValue is false for empty / whitespace / values containing "placeholder".
func liveValue(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && !strings.Contains(strings.ToLower(s), "placeholder")
}
