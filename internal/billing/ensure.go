package billing

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) EnsurePersonal(ctx context.Context, userID string) (*BillingAccount, *Subscription, error) {
	if userID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	return s.ensureOwner(ctx, OwnerUser, userID)
}

func (s *Service) EnsureOrganization(ctx context.Context, orgID string) (*BillingAccount, *Subscription, error) {
	if orgID == "" {
		return nil, nil, utils.E(utils.CodeValidation)
	}
	return s.ensureOwner(ctx, OwnerOrganization, orgID)
}

func (s *Service) ensureOwner(ctx context.Context, ownerType, ownerID string) (*BillingAccount, *Subscription, error) {
	acc, err := s.accounts.GetByOwner(ctx, ownerType, ownerID)
	if err == nil {
		sub, serr := s.subs.GetByAccountID(ctx, acc.ID)
		if serr == nil {
			return acc, sub, nil
		}
		if !errors.Is(serr, ErrNotFound) {
			return nil, nil, mapRepoErr(serr)
		}
		sub, serr = s.createTrial(ctx, acc)
		return acc, sub, serr
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	acc = &BillingAccount{
		ID: s.ids.Account(), OwnerType: ownerType, OwnerID: ownerID,
		Provider: s.provider.Name(), CreatedAt: now, UpdatedAt: now,
	}
	var sub *Subscription
	err = s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.accounts.Insert(ctx, acc); err != nil {
			return err
		}
		created, err := s.createTrial(ctx, acc)
		if err != nil {
			return err
		}
		sub = created
		return nil
	})
	if errors.Is(err, ErrDuplicate) {
		return s.ensureOwner(ctx, ownerType, ownerID)
	}
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	return acc, sub, nil
}

func (s *Service) createTrial(ctx context.Context, acc *BillingAccount) (*Subscription, error) {
	plan, err := s.plans.GetByID(ctx, PlanIDTrial)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	start := now
	days := plan.TrialDays
	if days <= 0 {
		days = 14
	}
	end := now.Add(time.Duration(days) * 24 * time.Hour)
	orgID := ""
	if acc.Organization() {
		orgID = acc.OwnerID
		if o, oerr := s.orgs.GetByID(ctx, acc.OwnerID); oerr == nil && o != nil {
			start = o.CreatedAt
			if o.TrialEndsAt != nil {
				end = *o.TrialEndsAt
			}
		}
	}
	sub := &Subscription{
		ID: s.ids.Subscription(), BillingAccountID: acc.ID, OrganizationID: orgID,
		PlanID: plan.ID, PlanCode: plan.Code, Status: StatusTrialing, Provider: s.provider.Name(),
		SeatsIncluded: plan.SeatsIncluded, TrialEndsAt: &end,
		CurrentPeriodStart: start, CurrentPeriodEnd: end, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.subs.Insert(ctx, sub); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return s.subs.GetByAccountID(ctx, acc.ID)
		}
		return nil, mapRepoErr(err)
	}
	return sub, nil
}
