package identity

import (
	"context"
	"errors"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// RequestOTP starts (or resends) a one-time login. Unknown emails are a no-op
// so callers can always return 202 without leaking account existence.
func (s *Service) RequestOTP(ctx context.Context, email string) error {
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
	otp := NewOTP()
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeOTPLogin,
		TokenHash: s.tokens.HashOTP(u.ID, otp),
		ExpiresAt: now.Add(s.cfg.OTPTTL),
		CreatedAt: now,
	}
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.InvalidateOpen(ctx, u.ID, PurposeOTPLogin, now); err != nil {
			return err
		}
		return s.emails.Insert(ctx, et)
	})
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendOTP(ctx, u.Email, otp); err != nil {
		l := obs.FromCtx(ctx)
		l.Error().Err(err).Str("user_id", u.ID).Msg("send otp mail")
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", u.ID).Msg("auth.otp_requested")
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, email, otp, deviceName string) (*User, *TokenPair, error) {
	if !validEmail(email) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if !validOTP(otp) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	if dn := strings.TrimSpace(deviceName); dn != "" && !validDisplayName(dn) {
		return nil, nil, utils.E(utils.CodeValidation)
	}

	u, err := s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
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

	tok, err := s.emails.GetByHash(ctx, s.tokens.HashOTP(u.ID, otp))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if tok.Purpose != PurposeOTPLogin || tok.ConsumedAt != nil || !tok.ExpiresAt.After(s.clock.Now()) || tok.UserID != u.ID {
		return nil, nil, utils.E(utils.CodeInvalidCredentials)
	}

	var pair *TokenPair
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		p, _, err := s.issueSession(ctx, u.ID, strings.TrimSpace(deviceName))
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeInvalidCredentials)
		}
		return nil, nil, mapRepoErr(err)
	}
	if err := s.users.SetLastLogin(ctx, u.ID, now); err != nil {
		l := obs.FromCtx(ctx)
		l.Error().Err(err).Str("user_id", u.ID).Msg("set last_login_at")
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", u.ID).Msg("auth.otp_login")
	u.PasswordHash = ""
	return u, pair, nil
}
