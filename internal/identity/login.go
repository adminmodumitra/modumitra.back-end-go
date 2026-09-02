package identity

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Login(ctx context.Context, in LoginInput) (*User, *TokenPair, error) {
	if !validEmail(in.Email) || in.Password == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if dn := strings.TrimSpace(in.DeviceName); dn != "" && !validDisplayName(dn) {
		return nil, nil, utils.E(utils.CodeValidation)
	}

	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.hasher.DummyVerify(ctx)
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if u.Suspended() {
		return nil, nil, utils.E(utils.CodeAccountSuspended)
	}
	if u.PasswordHash == "" {
		s.hasher.DummyVerify(ctx)
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}
	ok, verr := s.hasher.Verify(ctx, u.PasswordHash, in.Password)
	if verr != nil {
		return nil, nil, utils.Generic(verr)
	}
	if !ok {
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}

	now := s.clock.Now()
	var pair *TokenPair
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		p, sess, err := s.issueSession(ctx, u.ID, strings.TrimSpace(in.DeviceName))
		if err != nil {
			return err
		}
		pair = p
		if err := s.users.SetLastLogin(ctx, u.ID, now); err != nil {
			return err
		}
		return outbox.Enqueue(ctx, s.outbox, outbox.AuthLogin(ctx, u.ID, sess.ID, now))
	})
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.login", "user_id", u.ID)
	u.PasswordHash = ""
	return u, pair, nil
}
