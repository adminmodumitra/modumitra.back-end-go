package billing

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) applyCheckoutCompleted(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	acc, err := s.accountFromEvent(ctx, obj)
	if err != nil {
		return "", "", err
	}
	planID := asString(obj, "plan_id")
	customer := asString(obj, "customer")
	provSub := asString(obj, "subscription")
	if planID == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return "", "", mapRepoErr(err)
	}
	if customer != "" {
		if err := s.accounts.SetProviderCustomer(ctx, acc.ID, customer, now); err != nil {
			return "", "", mapRepoErr(err)
		}
		if acc.Organization() {
			_ = s.orgs.SetProviderCustomer(ctx, acc.OwnerID, s.provider.Name(), customer, now)
		}
	}
	periodEnd := now.AddDate(0, 1, 0)
	sub, err := s.subs.GetByAccountID(ctx, acc.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	if sub == nil {
		sub = &Subscription{ID: s.ids.Subscription(), BillingAccountID: acc.ID, OrganizationID: accOrg(acc), CreatedAt: now}
	}
	sub.PlanID, sub.PlanCode, sub.Status = plan.ID, plan.Code, StatusActive
	sub.Provider, sub.ProviderSubscriptionID = s.provider.Name(), provSub
	sub.SeatsIncluded = plan.SeatsIncluded
	sub.CurrentPeriodStart, sub.CurrentPeriodEnd = now, periodEnd
	sub.CancelAtPeriodEnd, sub.CanceledAt, sub.UpdatedAt = false, nil, now
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	if _, err := s.subs.GetByID(ctx, sub.ID); err != nil {
		if err := s.subs.Insert(ctx, sub); err != nil && !errors.Is(err, ErrDuplicate) {
			return "", "", mapRepoErr(err)
		}
	} else if err := s.subs.Update(ctx, sub); err != nil {
		return "", "", mapRepoErr(err)
	}
	if err := s.ensurePaidInvoice(ctx, obj, acc, sub, plan, now, periodEnd); err != nil {
		return "", "", err
	}
	return accOrg(acc), sub.ID, nil
}

func accOrg(a *BillingAccount) string {
	if a != nil && a.Organization() {
		return a.OwnerID
	}
	return ""
}

func (s *Service) accountFromEvent(ctx context.Context, obj map[string]any) (*BillingAccount, error) {
	if id := asString(obj, "billing_account_id"); id != "" {
		acc, err := s.accounts.GetByID(ctx, id)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		return acc, nil
	}
	orgID := asString(obj, "organization_id", "client_reference_id")
	if orgID == "" {
		return nil, utils.E(utils.CodeInvalidPayload)
	}
	acc, _, err := s.EnsureOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
