package org

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func ParseLimit(raw string) (int, error) {
	if raw == "" {
		return DefaultLimit, nil
	}
	n := 0
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, utils.E(utils.CodeValidation)
		}
		n = n*10 + int(c-'0')
		if n > MaxLimit {
			return 0, utils.E(utils.CodeValidation)
		}
	}
	if n < 1 {
		return 0, utils.E(utils.CodeValidation)
	}
	return n, nil
}

func EncodeCursor(t time.Time, id string) string {
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(s string) (time.Time, string, error) {
	if s == "" {
		return time.Time{}, "", nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, "", utils.E(utils.CodeValidation)
	}
	ts, id, ok := strings.Cut(string(b), "|")
	if !ok || id == "" {
		return time.Time{}, "", utils.E(utils.CodeValidation)
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return time.Time{}, "", utils.E(utils.CodeValidation)
	}
	return t, id, nil
}

func BeforeCursor(created time.Time, id string, cursorT time.Time, cursorID string) bool {
	if cursorID == "" {
		return true
	}
	if created.Equal(cursorT) {
		return id < cursorID
	}
	return created.Before(cursorT)
}
