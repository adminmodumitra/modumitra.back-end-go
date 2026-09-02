package org

import (
	"context"
	"errors"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) lookupPendingInvite(ctx context.Context, rawToken string, forAccept bool) (*Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || !strings.HasPrefix(rawToken, PrefixInviteTok) {
		return nil, utils.E(utils.CodeNotFound)
	}
	inv, err := s.invites.GetByTokenHash(ctx, s.tokens.Hash(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	expired := !inv.ExpiresAt.After(now)
	if !forAccept {
		if inv.Status != InvitePending || expired {
			return nil, utils.E(utils.CodeNotFound)
		}
		return inv, nil
	}
	return inv, nil
}

func (s *Service) lookupInviteForAccept(ctx context.Context, rawToken string) (*Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || !strings.HasPrefix(rawToken, PrefixInviteTok) {
		return nil, utils.E(utils.CodeNotFound)
	}
	inv, err := s.invites.GetByTokenHash(ctx, s.tokens.Hash(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	switch {
	case inv.Status == InviteRevoked:
		return nil, utils.E(utils.CodeInvitationRevoked)
	case inv.Status == InviteAccepted:
		return nil, utils.E(utils.CodeInvitationAccepted)
	case inv.Status == InviteExpired || !inv.ExpiresAt.After(now):
		return nil, utils.E(utils.CodeInvitationExpired)
	case inv.Status != InvitePending:
		return nil, utils.E(utils.CodeNotFound)
	}
	return inv, nil
}
