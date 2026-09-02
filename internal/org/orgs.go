package org

import (
	"context"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) PrimaryForUser(ctx context.Context, userID string) (*Organization, error) {
	if userID == "" {
		return nil, nil
	}
	items, _, err := s.members.ListByUser(ctx, userID, "", 1)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	for _, it := range items {
		if it.Membership.Active() && !it.Organization.Deleted() {
			o := it.Organization
			return &o, nil
		}
	}
	return nil, nil
}

func (s *Service) Create(ctx context.Context, actorID, name string) (*Organization, *Membership, error) {
	name = strings.TrimSpace(name)
	if actorID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	if !validOrgName(name) {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	u, err := s.userByID(ctx, actorID)
	if err != nil {
		return nil, nil, err
	}
	if !u.EmailVerified() {
		return nil, nil, utils.E(utils.CodeEmailUnverified)
	}
	now := s.clock.Now()
	trial := now.Add(time.Duration(s.trialDays) * 24 * time.Hour)
	o := &Organization{
		ID:          s.ids.Org(),
		Name:        name,
		OwnerUserID: actorID,
		Status:      OrgActive,
		TrialEndsAt: &trial,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m := &Membership{
		ID:             s.ids.Member(),
		OrganizationID: o.ID,
		UserID:         actorID,
		Role:           RoleOwner,
		Status:         MemberActive,
		ActivatedAt:    &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.orgs.Insert(ctx, o); err != nil {
			return err
		}
		if err := s.members.Insert(ctx, m); err != nil {
			return err
		}
		if err := outbox.Enqueue(ctx, s.outbox, outbox.OrgCreated(ctx, o.ID, actorID, o.Name, now)); err != nil {
			return err
		}
		if err := outbox.Enqueue(ctx, s.outbox, outbox.MembershipCreated(ctx, m.ID, o.ID, actorID, m.Role, now)); err != nil {
			return err
		}
		if o.TrialEndsAt != nil {
			if err := outbox.Enqueue(ctx, s.outbox, outbox.BillingTrialStarted(ctx, o.ID, actorID, *o.TrialEndsAt, now)); err != nil {
				return err
			}
		}
		if s.provisioner != nil {
			return s.provisioner.EnsureOrganization(ctx, o.ID)
		}
		return nil
	})
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	l := obs.FromCtx(ctx)
	l.Info().Str("organization_id", o.ID).Str("user_id", actorID).Msg("org.created")
	return o, m, nil
}

func (s *Service) List(ctx context.Context, actorID, cursor string, limit int) ([]MembershipItem, string, error) {
	if actorID == "" {
		return nil, "", utils.E(utils.CodeUnauthenticated)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		return nil, "", utils.E(utils.CodeValidation)
	}
	items, next, err := s.members.ListByUser(ctx, actorID, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	return items, next, nil
}

func (s *Service) Get(ctx context.Context, actorID, orgID string) (*Organization, error) {
	o, _, err := s.loadActiveMember(ctx, orgID, actorID)
	return o, err
}

func (s *Service) UpdateName(ctx context.Context, actorID, orgID, name string) (*Organization, error) {
	name = strings.TrimSpace(name)
	if !validOrgName(name) {
		return nil, utils.E(utils.CodeValidation)
	}
	o, m, err := s.loadActiveMember(ctx, orgID, actorID)
	if err != nil {
		return nil, err
	}
	if err := requireRoles(m, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := s.orgs.UpdateName(ctx, orgID, name, now); err != nil {
		return nil, mapRepoErr(err)
	}
	o.Name = name
	o.UpdatedAt = now
	l := obs.FromCtx(ctx)
	l.Info().Str("organization_id", orgID).Str("user_id", actorID).Msg("org.updated")
	return o, nil
}
