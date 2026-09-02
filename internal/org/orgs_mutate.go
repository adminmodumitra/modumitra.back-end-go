package org

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Delete(ctx context.Context, actorID, orgID string) error {
	_, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if err := requireRoles(m, RoleOwner); err != nil {
		return err
	}
	now := s.clock.Now()
	if err := s.orgs.SoftDelete(ctx, orgID, now); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.deleted", "organization_id", orgID, "user_id", actorID)
	return nil
}

func (s *Service) Leave(ctx context.Context, actorID, orgID string) error {
	_, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if m.Role == RoleOwner {
		n, err := s.members.CountActiveOwners(ctx, orgID)
		if err != nil {
			return mapRepoErr(err)
		}
		if n <= 1 {
			return utils.E(utils.CodeLastOwner)
		}
	}
	now := s.clock.Now()
	m.Status = MemberRemoved
	m.RemovedAt = &now
	m.UpdatedAt = now
	if err := s.members.Update(ctx, m); err != nil {
		return mapRepoErr(err)
	}
	slog.InfoContext(ctx, "org.left", "organization_id", orgID, "user_id", actorID)
	return nil
}

func (s *Service) TransferOwnership(ctx context.Context, actorID, orgID, targetUserID string) (*Organization, error) {
	if targetUserID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	o, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(m, RoleOwner); err != nil {
		return nil, err
	}
	if targetUserID == actorID {
		return nil, utils.E(utils.CodeAlreadyOwner)
	}
	target, err := s.members.GetByOrgUser(ctx, orgID, targetUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	if !target.Active() {
		return nil, utils.E(utils.CodeNotFound)
	}
	if target.Role == RoleOwner {
		return nil, utils.E(utils.CodeAlreadyOwner)
	}
	now := s.clock.Now()
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		m.Role = RoleAdmin
		m.UpdatedAt = now
		if err := s.members.Update(ctx, m); err != nil {
			return err
		}
		target.Role = RoleOwner
		target.UpdatedAt = now
		if err := s.members.Update(ctx, target); err != nil {
			return err
		}
		return s.orgs.SetOwner(ctx, orgID, targetUserID, now)
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	o.OwnerUserID = targetUserID
	o.UpdatedAt = now
	slog.InfoContext(ctx, "org.ownership_transferred", "organization_id", orgID, "from", actorID, "to", targetUserID)
	return o, nil
}

func (s *Service) Roles(ctx context.Context, actorID, orgID string) (map[string]map[string]bool, error) {
	if _, _, err := s.loadActiveMember(ctx, orgID, actorID); err != nil {
		return nil, err
	}
	return RolesCatalog(), nil
}
