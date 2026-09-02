package device

import (
	"context"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) Issue(ctx context.Context, userID, orgID, deviceID string) (*Snapshot, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if orgID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	now := s.clock.Now().UTC()
	var snap *Snapshot
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		var err error
		snap, err = s.issueLocked(ctx, userID, orgID, deviceID, now)
		return err
	})
	return snap, err
}

func (s *Service) Capabilities(ctx context.Context, userID, orgID string) (*CapabilitiesResult, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	now := s.clock.Now().UTC()
	personalPlan, personalStatus := PlanNone, SubNone
	if s.planLookup != nil {
		p, st, err := s.planLookup.PlanForOwner(ctx, "user", userID)
		if err != nil {
			return nil, err
		}
		personalPlan, personalStatus = p, st
	}
	personalCaps := Intersect("owner", personalPlan)
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return &CapabilitiesResult{
			UserID: userID, Plan: personalPlan, SubscriptionStatus: personalStatus,
			Capabilities: personalCaps, IssuedAt: now,
		}, nil
	}
	if s.orgs == nil {
		return nil, utils.E(utils.CodeServiceUnavailable)
	}
	o, m, err := s.orgs.MembershipOf(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if m.Status == MemberSuspended {
		return nil, utils.E(utils.CodeForbidden)
	}
	if !m.Active() {
		return nil, utils.E(utils.CodeNotFound)
	}
	plan, status := planOf(o, now)
	if s.planLookup != nil {
		if p, st, err := s.planLookup.PlanForOwner(ctx, "organization", orgID); err != nil {
			return nil, err
		} else if p != PlanNone {
			plan, status = p, st
		}
	}
	return &CapabilitiesResult{
		UserID: userID, OrganizationID: o.ID, Role: m.Role, MembershipStatus: m.Status,
		Plan: plan, SubscriptionStatus: status,
		Capabilities: Union(personalCaps, Intersect(m.Role, plan)), IssuedAt: now,
	}, nil
}
