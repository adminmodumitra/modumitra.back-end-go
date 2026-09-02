package identity

import (
	"context"
	"errors"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// UserByID loads a user including PasswordHash (SSO uses it to detect passwordless accounts).
func (s *Service) UserByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, ErrNotFound
	}
	return s.users.GetByID(ctx, id)
}

// UserByEmail looks up by normalized email. Missing rows are ErrNotFound.
func (s *Service) UserByEmail(ctx context.Context, email string) (*User, error) {
	return s.users.GetByEmailNormalized(ctx, normalizeEmail(email))
}

// CreateUser inserts u and enqueues user.created. Caller sets ID, timestamps, and hashes.
func (s *Service) CreateUser(ctx context.Context, u *User) error {
	if u == nil || u.ID == "" {
		return utils.E(utils.CodeValidation)
	}
	now := s.clock.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	if u.Status == "" {
		u.Status = StatusActive
	}
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.users.Insert(ctx, u); err != nil {
			return err
		}
		return outbox.Enqueue(ctx, s.outbox, outbox.UserCreated(ctx, u.ID, now))
	})
	if err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Service) NewUserID() string { return s.ids.User() }

func (s *Service) TermsVersion() string   { return s.cfg.TermsVersion }
func (s *Service) PrivacyVersion() string { return s.cfg.PrivacyVersion }

// LoginAs issues a session for an already-authenticated user (SSO / device flow)
// and writes auth.login to the outbox in the same transaction.
func (s *Service) LoginAs(ctx context.Context, userID, deviceName string) (*User, *TokenPair, error) {
	if userID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	if dn := strings.TrimSpace(deviceName); dn != "" && !validDisplayName(dn) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeNotFound)
		}
		return nil, nil, mapRepoErr(err)
	}
	if u.Suspended() {
		return nil, nil, utils.E(utils.CodeAccountSuspended)
	}
	now := s.clock.Now()
	var pair *TokenPair
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		p, sess, err := s.issueSession(ctx, u.ID, strings.TrimSpace(deviceName))
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
	u.PasswordHash = ""
	return u, pair, nil
}
