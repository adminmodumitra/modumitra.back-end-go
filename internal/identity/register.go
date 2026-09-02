package identity

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Register(ctx context.Context, in RegisterInput) (*User, error) {
	email := strings.TrimSpace(in.Email)
	display := strings.TrimSpace(in.DisplayName)
	if !validEmail(email) || !validDisplayName(display) || in.Password == "" || in.TermsVersion == "" || in.PrivacyVersion == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	if in.TermsVersion != s.cfg.TermsVersion || in.PrivacyVersion != s.cfg.PrivacyVersion {
		return nil, utils.E(utils.CodeTermsOutdated)
	}
	if len(in.Password) < minPasswordLen {
		return nil, utils.E(utils.CodeWeakPassword)
	}

	hash, err := s.hasher.Hash(ctx, in.Password)
	if err != nil {
		return nil, utils.Generic(err)
	}

	now := s.clock.Now()
	u := &User{
		ID:              s.ids.User(),
		Email:           email,
		EmailNormalized: normalizeEmail(email),
		PasswordHash:    hash,
		DisplayName:     display,
		TermsVersion:    in.TermsVersion,
		PrivacyVersion:  in.PrivacyVersion,
		Status:          StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixVerify)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeEmailVerify,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}

	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.users.Insert(ctx, u); err != nil {
			return err
		}
		if err := s.emails.Insert(ctx, et); err != nil {
			return err
		}
		return outbox.Enqueue(ctx, s.outbox, outbox.UserCreated(ctx, u.ID, now))
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}

	if err := s.mailer.SendVerification(ctx, u.Email, rawTok); err != nil {
		slog.ErrorContext(ctx, "send verification mail", "user_id", u.ID, "error", err)
	}
	slog.InfoContext(ctx, "auth.register", "user_id", u.ID)
	out := *u
	out.PasswordHash = ""
	return &out, nil
}

func (s *Service) VerifyEmail(ctx context.Context, raw string) (*User, error) {
	tok, err := s.lookupEmailToken(ctx, raw, PrefixVerify, PurposeEmailVerify)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	var user *User
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.emails.Consume(ctx, tok.ID, now); err != nil {
			return err
		}
		if err := s.users.SetEmailVerified(ctx, tok.UserID, now); err != nil {
			return err
		}
		u, err := s.users.GetByID(ctx, tok.UserID)
		if err != nil {
			return err
		}
		user = u
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeInvalidToken)
		}
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "auth.email_verified", "user_id", user.ID)
	user.PasswordHash = ""
	return user, nil
}

func (s *Service) ResendVerification(ctx context.Context, email string) error {
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
	if u.EmailVerified() {
		return nil
	}
	now := s.clock.Now()
	rawTok, tokHash := s.tokens.NewEmailToken(PrefixVerify)
	et := &EmailToken{
		ID:        s.ids.Token(),
		UserID:    u.ID,
		Purpose:   PurposeEmailVerify,
		TokenHash: tokHash,
		ExpiresAt: now.Add(s.cfg.EmailTokenTTL),
		CreatedAt: now,
	}
	if err := s.emails.Insert(ctx, et); err != nil {
		return mapRepoErr(err)
	}
	if err := s.mailer.SendVerification(ctx, u.Email, rawTok); err != nil {
		slog.ErrorContext(ctx, "send verification mail", "user_id", u.ID, "error", err)
	}
	return nil
}
