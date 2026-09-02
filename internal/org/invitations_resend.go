package org

import (
	"context"
	"log/slog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) RevokeInvitation(ctx context.Context, actorID, orgID, invitationID string) error {
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return err
	}
	inv, err := s.invites.GetByID(ctx, invitationID)
	if err != nil {
		return mapRepoErr(err)
	}
	if inv.OrganizationID != orgID {
		return utils.E(utils.CodeNotFound)
	}
	now := s.clock.Now()
	if inv.Status != InvitePending || !inv.ExpiresAt.After(now) {
		return utils.E(utils.CodeInvitationNotPending)
	}
	inv.Status = InviteRevoked
	inv.RevokedAt = &now
	inv.RevokedByUserID = &actorID
	inv.UpdatedAt = now
	if err := s.invites.Update(ctx, inv); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.invitation_revoked", "invitation_id", invitationID)
	return nil
}

func (s *Service) ResendInvitation(ctx context.Context, actorID, orgID, invitationID string) (*Invitation, error) {
	o, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	inv, err := s.invites.GetByID(ctx, invitationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if inv.OrganizationID != orgID {
		return nil, utils.E(utils.CodeNotFound)
	}
	now := s.clock.Now()
	if inv.Status != InvitePending || !inv.ExpiresAt.After(now) {
		return nil, utils.E(utils.CodeInvitationNotPending)
	}
	raw, hash := s.tokens.NewInvite()
	inv.TokenHash = hash
	inv.ExpiresAt = now.Add(InviteTTL)
	inv.ResentAt = &now
	inv.UpdatedAt = now
	if err := s.invites.Update(ctx, inv); err != nil {
		return nil, mapRepoErr(err)
	}
	if s.mailer != nil {
		if err := s.mailer.SendInvite(ctx, inv.Email, raw, o.Name); err != nil {
			slog.ErrorContext(ctx, "resend invite mail", "invitation_id", invitationID, "error", err)
		}
	}
	slog.InfoContext(ctx, "org.invitation_resent", "invitation_id", invitationID)
	return inv, nil
}
