package identity

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validEmail(raw string) bool {
	s := strings.TrimSpace(raw)
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address == "" {
		return false
	}
	return true
}

func validDisplayName(s string) bool {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return n >= 1 && n <= 80
}

func validOTP(otp string) bool {
	if len(otp) != 6 {
		return false
	}
	for i := 0; i < 6; i++ {
		if otp[i] < '0' || otp[i] > '9' {
			return false
		}
	}
	return true
}

func (s *Service) lookupEmailToken(ctx context.Context, raw, prefix, purpose string) (*EmailToken, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || (prefix != "" && !strings.HasPrefix(raw, prefix)) {
		return nil, utils.E(utils.CodeInvalidToken)
	}
	tok, err := s.emails.GetByHash(ctx, s.tokens.Hash(raw))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeInvalidToken)
		}
		return nil, mapRepoErr(err)
	}
	if tok.Purpose != purpose || tok.ConsumedAt != nil || !tok.ExpiresAt.After(s.clock.Now()) {
		return nil, utils.E(utils.CodeInvalidToken)
	}
	return tok, nil
}
