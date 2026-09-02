package org

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) CreateInvitation(ctx context.Context, actorID, orgID, email, role string) (*Invitation, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return nil, utils.E(utils.CodeValidation)
	}
	if role == RoleOwner {
		return nil, utils.E(utils.CodeRoleNotAssignable)
	}
	if !AssignableRole(role) {
		return nil, utils.E(utils.CodeValidation)
	}
	o, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	if _, err := s.invites.GetPendingByOrgEmail(ctx, orgID, email); err == nil {
		return nil, utils.E(utils.CodeInvitationPending)
	} else if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if s.users != nil {
		if u, err := s.users.GetByEmailNormalized(ctx, email); err == nil && u != nil {
			if mem, merr := s.members.GetByOrgUser(ctx, orgID, u.ID); merr == nil && mem.Active() {
				return nil, utils.E(utils.CodeAlreadyMember)
			}
		} else if err != nil && !errors.Is(err, identity.ErrNotFound) {
			return nil, mapRepoErr(err)
		}
	}
	now := s.clock.Now()
	raw, hash := s.tokens.NewInvite()
	inv := &Invitation{
		ID:              s.ids.Invitation(),
		OrganizationID:  orgID,
		Email:           email,
		Role:            role,
		Status:          InvitePending,
		TokenHash:       hash,
		InvitedByUserID: actorID,
		ExpiresAt:       now.Add(InviteTTL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.invites.Insert(ctx, inv); err != nil {
		return nil, mapRepoErr(err)
	}
	if s.mailer != nil {
		if err := s.mailer.SendInvite(ctx, email, raw, o.Name); err != nil {
			slog.ErrorContext(ctx, "send invite mail", "organization_id", orgID, "error", err)
		}
	}
	slog.InfoContext(ctx, "org.invitation_created", "organization_id", orgID, "invitation_id", inv.ID)
	return inv, nil
}

func (s *Service) ListInvitations(ctx context.Context, actorID, orgID, status, cursor string, limit int) ([]Invitation, string, error) {
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, "", err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, "", err
	}
	if status == "" {
		status = InvitePending
	}
	switch status {
	case InvitePending, InviteAccepted, InviteRevoked, InviteExpired:
	default:
		return nil, "", utils.E(utils.CodeValidation)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.invites.ListByOrg(ctx, orgID, status, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	now := s.clock.Now()
	for i := range items {
		if items[i].Status == InvitePending && !items[i].ExpiresAt.After(now) {
			items[i].Status = InviteExpired
		}
	}
	return items, next, nil
}
