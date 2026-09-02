package org

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) SuspendMember(ctx context.Context, actorID, orgID, targetUserID string) (*Membership, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	if targetUserID == actorID {
		return nil, utils.E(utils.CodeCannotSuspendSelf)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil, utils.E(utils.CodeNotFound)
	}
	if target.Role == RoleOwner {
		return nil, utils.E(utils.CodeLastOwner)
	}
	if actor.Role == RoleAdmin && (target.Role == RoleAdmin || target.Role == RoleOwner) {
		return nil, utils.E(utils.CodeForbidden)
	}
	if target.Status == MemberSuspended {
		return target, nil
	}
	now := s.clock.Now()
	target.Status = MemberSuspended
	target.SuspendedAt = &now
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return nil, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("organization_id", orgID).Str("user_id", targetUserID).Msg("org.member_suspended")
	return target, nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, orgID, targetUserID string) error {
	if targetUserID == "" {
		return utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return err
	}
	if targetUserID == actorID {
		return utils.E(utils.CodeCannotRemoveSelf)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil
	}
	if target.Role == RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return mapRepoErr(err)
		}
		if n <= 1 {
			return utils.E(utils.CodeLastOwner)
		}
	}
	if actor.Role == RoleAdmin && (target.Role == RoleAdmin || target.Role == RoleOwner) {
		return utils.E(utils.CodeForbidden)
	}
	now := s.clock.Now()
	target.Status = MemberRemoved
	target.RemovedAt = &now
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("organization_id", orgID).Str("user_id", targetUserID).Msg("org.member_removed")
	return nil
}
