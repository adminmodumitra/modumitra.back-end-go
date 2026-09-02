package billing

import (
	"context"
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) GetSubscription(ctx context.Context, actorID, orgID string) (*Subscription, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(m); err != nil {
		return nil, err
	}
	return s.getOrCreateTrial(ctx, o)
}

func (s *Service) getOrCreateTrial(ctx context.Context, o *org.Organization) (*Subscription, error) {
	sub, err := s.subs.GetByOrg(ctx, o.ID)
	if err == nil {
		return sub, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if o.TrialEndsAt == nil || !o.TrialEndsAt.After(now) {
		return nil, utils.E(utils.CodeNotFound)
	}
	plan, err := s.plans.GetByID(ctx, PlanIDTrial)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	sub = &Subscription{
		ID:                 s.ids.Subscription(),
		OrganizationID:     o.ID,
		PlanID:             plan.ID,
		PlanCode:           plan.Code,
		Status:             StatusTrialing,
		Provider:           s.provider.Name(),
		SeatsIncluded:      plan.SeatsIncluded,
		TrialEndsAt:        o.TrialEndsAt,
		CurrentPeriodStart: o.CreatedAt,
		CurrentPeriodEnd:   *o.TrialEndsAt,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.subs.Insert(ctx, sub); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return s.subs.GetByOrg(ctx, o.ID)
		}
		return nil, mapRepoErr(err)
	}
	return sub, nil
}

func (s *Service) Checkout(ctx context.Context, actorID, orgID, planID, successURL, cancelURL string) (*CheckoutSession, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	u, err := s.users.GetByID(ctx, actorID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !u.EmailVerified() {
		return nil, utils.E(utils.CodeEmailUnverified)
	}
	if !allowedRedirect(successURL) || !allowedRedirect(cancelURL) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, utils.E(utils.CodePlanNotFound)
	}
	if !plan.Sellable || !plan.Active {
		return nil, utils.E(utils.CodePlanNotFound)
	}
	existing, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if existing != nil && existing.PlanCode != PlanTrial && (existing.Status == StatusActive || existing.Status == StatusTrialing) {
		return nil, utils.E(utils.CodeSubscriptionActive)
	}
	sess, err := s.provider.CreateCheckout(ctx, CheckoutInput{
		OrganizationID: orgID,
		PlanID:         plan.ID,
		SuccessURL:     successURL,
		CancelURL:      cancelURL,
		CustomerID:     o.ProviderCustomerID,
		Email:          u.Email,
		Currency:       plan.Currency,
		AmountMinor:    plan.UnitAmountMinor,
		Seats:          plan.SeatsIncluded,
		PlanCode:       plan.Code,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}

func (s *Service) Portal(ctx context.Context, actorID, orgID, returnURL string) (*PortalSession, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	if o.ProviderCustomerID == "" {
		return nil, utils.E(utils.CodeBillingCustomerMissing)
	}
	if !allowedRedirect(returnURL) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	sess, err := s.provider.CreatePortal(ctx, PortalInput{
		OrganizationID: orgID,
		CustomerID:     o.ProviderCustomerID,
		ReturnURL:      returnURL,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}
