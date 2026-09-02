package identity

import (
	"context"
	"errors"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	hash := s.tokens.Hash(raw)
	now := s.clock.Now()

	var pair *TokenPair
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		sess, err := s.sessions.GetByRefreshHashForUpdate(ctx, hash)
		if err != nil {
			return err
		}
		// Rotated/spent token (replaced_by set) or already flagged reuse.
		if sess.ReplacedBy != nil || sess.ReuseDetectedAt != nil {
			if markErr := s.sessions.MarkReuseAndRevokeFamily(ctx, sess.ID, sess.FamilyID, now); markErr != nil {
				return markErr
			}
			return utils.E(utils.CodeTokenReuse)
		}
		if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
			return utils.E(utils.CodeUnauthenticated)
		}

		rawRT, rtHash := s.tokens.NewRefresh()
		newSess := &Session{
			ID:          s.ids.Session(),
			UserID:      sess.UserID,
			RefreshHash: rtHash,
			FamilyID:    sess.FamilyID,
			ExpiresAt:   now.Add(s.cfg.RefreshTTL),
			DeviceName:  sess.DeviceName,
			CreatedAt:   now,
			UpdatedAt:   now,
			LastSeenAt:  now,
		}
		if err := s.sessions.Insert(ctx, newSess); err != nil {
			return err
		}
		if err := s.sessions.RevokeAndReplace(ctx, sess.ID, newSess.ID, now); err != nil {
			return err
		}
		access, err := s.tokens.IssueAccess(ctx, sess.UserID, newSess.ID)
		if err != nil {
			return err
		}
		pair = &TokenPair{
			AccessToken:  access,
			RefreshToken: rawRT,
			ExpiresIn:    s.expiresIn(),
			TokenType:    "Bearer",
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeUnauthenticated)
		}
		return nil, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Msg("auth.refresh")
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return utils.E(utils.CodeValidation)
	}
	hash := s.tokens.Hash(raw)
	sess, err := s.sessions.GetByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}
	if err := s.sessions.Revoke(ctx, sess.ID, s.clock.Now()); err != nil && !errors.Is(err, ErrNotFound) {
		return mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("session_id", sess.ID).Msg("auth.logout")
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, utils.E(utils.CodeUnauthenticated)
	}
	n, err := s.sessions.RevokeAllForUser(ctx, userID, s.clock.Now())
	if err != nil {
		return 0, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("user_id", userID).Int("revoked", n).Msg("auth.logout_all")
	return n, nil
}
