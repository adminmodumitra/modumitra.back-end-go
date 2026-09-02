package billing

import (
	"context"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type access struct {
	Account *BillingAccount
	Member  *org.Membership
}

func (s *Service) Resolve(ctx context.Context, userID, orgID string) (*BillingAccount, error) {
	a, err := s.resolveAccess(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	return a.Account, nil
}

func (s *Service) resolveAccess(ctx context.Context, actorID, orgID string) (*access, error) {
	if actorID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if strings.TrimSpace(orgID) == "" {
		acc, _, err := s.EnsurePersonal(ctx, actorID)
		if err != nil {
			return nil, err
		}
		return &access{Account: acc}, nil
	}
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	acc, _, err := s.EnsureOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &access{Account: acc, Member: m}, nil
}

func (s *Service) loadVisible(ctx context.Context, actorID, orgID string) (*org.Organization, *org.Membership, error) {
	if actorID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	o, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if o.Deleted() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	m, err := s.members.GetByOrgUser(ctx, orgID, actorID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if !m.Active() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	return o, m, nil
}

func requireRead(a *access) error {
	if a == nil || a.Account == nil {
		return utils.E(utils.CodeNotFound)
	}
	if a.Account.Personal() {
		return nil
	}
	return requireMemberRead(a.Member)
}

func requireWrite(a *access) error {
	if a == nil || a.Account == nil {
		return utils.E(utils.CodeNotFound)
	}
	if a.Account.Personal() {
		return nil
	}
	return requireMemberWrite(a.Member)
}

func requireMemberRead(m *org.Membership) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	switch m.Role {
	case org.RoleOwner, org.RoleAdmin, org.RoleBillingAdmin:
		return nil
	default:
		return utils.E(utils.CodeForbidden)
	}
}

func requireMemberWrite(m *org.Membership) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	switch m.Role {
	case org.RoleOwner, org.RoleBillingAdmin:
		return nil
	default:
		return utils.E(utils.CodeForbidden)
	}
}
