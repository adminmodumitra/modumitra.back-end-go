package org

import (
	"context"
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) loadOrg(ctx context.Context, id string) (*Organization, error) {
	o, err := s.orgs.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if o.Deleted() {
		return nil, utils.E(utils.CodeNotFound)
	}
	return o, nil
}

func (s *Service) loadActiveMember(ctx context.Context, orgID, userID string) (*Organization, *Membership, error) {
	o, err := s.loadOrg(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}
	m, err := s.members.GetByOrgUser(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, utils.E(utils.CodeNotFound)
		}
		return nil, nil, mapRepoErr(err)
	}
	if !m.Active() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	return o, m, nil
}

func requireRoles(m *Membership, roles ...string) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	for _, r := range roles {
		if m.Role == r {
			return nil
		}
	}
	return utils.E(utils.CodeForbidden)
}

func (s *Service) userByID(ctx context.Context, id string) (*identity.User, error) {
	if s.users == nil {
		return nil, utils.E(utils.CodeServiceUnavailable)
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) {
			return nil, utils.E(utils.CodeNotFound)
		}
		return nil, mapRepoErr(err)
	}
	return u, nil
}

// Resolve is used by Tenant middleware: active membership role, or not_found.
func (s *Service) Resolve(ctx context.Context, orgID, userID string) (string, error) {
	_, m, err := s.loadActiveMember(ctx, orgID, userID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

// MembershipOf returns the caller's membership including suspended (not removed).
// Missing org/membership → not_found. Used by unsigned capabilities and snapshots.
func (s *Service) MembershipOf(ctx context.Context, orgID, userID string) (*Organization, *Membership, error) {
	o, err := s.loadOrg(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}
	m, err := s.members.GetByOrgUser(ctx, orgID, userID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if m.Status == MemberRemoved || m.Status == MemberInvited {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	return o, m, nil
}

func (s *Service) MemberUser(ctx context.Context, userID string) (*identity.User, error) {
	return s.userByID(ctx, userID)
}
