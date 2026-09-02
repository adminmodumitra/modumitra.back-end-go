package identity

import (
	"context"
	"errors"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	if !validEmail(email) {
		return utils.E(utils.CodeValidation)
	}
	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixReset)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposePasswordReset,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}
	if err := s.emails.Insert(ctx, et); err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendPasswordReset(ctx, u.Email, rawTok); err != nil {
		l := obs.FromCtx(ctx)
		l.Error().Err(err).Str("user_id", u.ID).Msg("send password reset mail")
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", u.ID).Msg("auth.password_reset_requested")
	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if strings.TrimSpace(rawToken) == "" {
		return utils.E(utils.CodeInvalidToken)
	}
	if len(newPassword) < minPasswordLen {
		return utils.E(utils.CodeWeakPassword)
	}
	tok, err := s.lookupEmailToken(ctx, rawToken, PrefixReset, PurposePasswordReset)
	if err != nil {
		return err
	}
	hash, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return utils.Generic(err)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		if err := s.users.UpdatePassword(ctx, tok.UserID, hash, now); err != nil {
			return err
		}
		_, err := s.sessions.RevokeAllForUser(ctx, tok.UserID, now)
		return err
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return utils.E(utils.CodeInvalidToken)
		}
		return mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", tok.UserID).Msg("auth.password_reset_confirm")
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID, current, newPassword string) error {
	if userID == "" {
		return utils.E(utils.CodeUnauthenticated)
	}
	if current == "" || newPassword == "" {
		return utils.E(utils.CodeValidation)
	}
	if len(newPassword) < minPasswordLen || newPassword == current {
		return utils.E(utils.CodeWeakPassword)
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return utils.E(utils.CodeNotFound)
		}
		return mapRepoErr(err)
	}
	if u.PasswordHash == "" {
		s.hasher.DummyVerify(ctx)
		return utils.E(utils.CodeInvalidCredentials)
	}
	ok, verr := s.hasher.Verify(ctx, u.PasswordHash, current)
	if verr != nil {
		return utils.Generic(verr)
	}
	if !ok {
		return utils.E(utils.CodeInvalidCredentials)
	}
	hash, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return utils.Generic(err)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.users.UpdatePassword(ctx, userID, hash, now); err != nil {
			return err
		}
		_, err := s.sessions.RevokeAllForUser(ctx, userID, now)
		return err
	})
	if err != nil {
		return mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", userID).Msg("auth.password_change")
	return nil
}
