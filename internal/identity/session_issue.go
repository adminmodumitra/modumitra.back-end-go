package identity

import (
	"context"
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	u.PasswordHash = ""
	return u, nil
}

// ValidateAccess parses a Bearer JWT, loads the session, and requires it
// not revoked and not expired.
func (s *Service) ValidateAccess(ctx context.Context, raw string) (string, string, error) {
	userID, sessionID, err := s.tokens.ParseAccess(ctx, raw)
	if err != nil {
		return "", "", utils.E(utils.CodeUnauthenticated).With(err)
	}
	sess, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", utils.E(utils.CodeUnauthenticated)
		}
		return "", "", mapRepoErr(err)
	}
	now := s.clock.Now()
	if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) || sess.UserID != userID {
		return "", "", utils.E(utils.CodeUnauthenticated)
	}
	return userID, sessionID, nil
}

func (s *Service) issueSession(ctx context.Context, userID, deviceName string) (*TokenPair, *Session, error) {
	now := s.clock.Now()
	rawRT, rtHash := s.tokens.NewRefresh()
	sess := &Session{
		ID:          s.ids.Session(),
		UserID:      userID,
		RefreshHash: rtHash,
		FamilyID:    s.ids.Family(),
		ExpiresAt:   now.Add(s.cfg.RefreshTTL),
		DeviceName:  deviceName,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastSeenAt:  now,
	}
	if err := s.sessions.Insert(ctx, sess); err != nil {
		return nil, nil, mapRepoErr(err)
	}
	access, err := s.tokens.IssueAccess(ctx, userID, sess.ID)
	if err != nil {
		return nil, nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: rawRT,
		ExpiresIn:    s.expiresIn(),
		TokenType:    "Bearer",
	}, sess, nil
}
