package org

import (
	"context"
	"log/slog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) ListMembers(ctx context.Context, actorID, orgID, status, role, cursor string, limit int) ([]Membership, string, error) {
	if _, _, err := s.loadActiveMember(ctx, orgID, actorID); err != nil {
		return nil, "", err
	}
	if status != "" && status != MemberActive && status != MemberSuspended {
		return nil, "", utils.E(utils.CodeValidation)
	}
	if role != "" && !ValidRole(role) {
		return nil, "", utils.E(utils.CodeValidation)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.members.ListByOrg(ctx, orgID, status, role, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	return items, next, nil
}

func (s *Service) PatchMember(ctx context.Context, actorID, orgID, targetUserID, role, status string) (*Membership, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	if role == "" && status == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	_, actor, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(actor, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if target.Status == MemberRemoved {
		return nil, utils.E(utils.CodeNotFound)
	}
	if role != "" {
		if role == RoleOwner {
			return nil, utils.E(utils.CodeRoleNotAssignable)
		}
		if !AssignableRole(role) {
			return nil, utils.E(utils.CodeValidation)
		}
	}
	if actor.Role == RoleAdmin {
		if target.Role == RoleOwner || target.Role == RoleAdmin {
			return nil, utils.E(utils.CodeForbidden)
		}
		if role == RoleAdmin {
			return nil, utils.E(utils.CodeForbidden)
		}
	}
	if target.Role == RoleOwner && role != "" && role != RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		if n <= 1 {
			return nil, utils.E(utils.CodeLastOwner)
		}
	}
	now := s.clock.Now()
	if role != "" {
		target.Role = role
	}
	if status == MemberActive || (status == "" && role != "" && target.Status == MemberSuspended) {
		target.Status = MemberActive
		target.SuspendedAt = nil
		if target.ActivatedAt == nil {
			target.ActivatedAt = &now
		}
	} else if status != "" && status != MemberActive && status != MemberSuspended {
		return nil, utils.E(utils.CodeValidation)
	}
	target.UpdatedAt = now
	if err := s.members.Update(ctx, target); err != nil {
		return nil, mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.member_patched", "organization_id", orgID, "user_id", targetUserID, "actor", actorID)
	return target, nil
}
