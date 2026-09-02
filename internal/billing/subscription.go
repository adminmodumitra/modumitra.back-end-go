package billing

import (
	"context"
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) GetSubscription(ctx context.Context, actorID, orgID string) (*Subscription, error) {
	a, err := s.resolveAccess(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(a); err != nil {
		return nil, err
	}
	sub, err := s.subs.GetByAccountID(ctx, a.Account.ID)
	if err == nil {
		return sub, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	return s.createTrial(ctx, a.Account)
}

func (s *Service) Checkout(ctx context.Context, actorID, orgID, planID, successURL, cancelURL string) (*CheckoutSession, error) {
	a, err := s.resolveAccess(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(a); err != nil {
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
	existing, err := s.subs.GetByAccountID(ctx, a.Account.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if existing != nil && existing.PlanCode != PlanTrial && (existing.Status == StatusActive || existing.Status == StatusTrialing) {
		return nil, utils.E(utils.CodeSubscriptionActive)
	}
	orgRef := ""
	if a.Account.Organization() {
		orgRef = a.Account.OwnerID
	}
	sess, err := s.provider.CreateCheckout(ctx, CheckoutInput{
		BillingAccountID: a.Account.ID, OrganizationID: orgRef, PlanID: plan.ID,
		SuccessURL: successURL, CancelURL: cancelURL, CustomerID: a.Account.CustomerID(),
		Email: u.Email, Currency: plan.Currency, AmountMinor: plan.UnitAmountMinor,
		Seats: plan.SeatsIncluded, PlanCode: plan.Code,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}

func (s *Service) Portal(ctx context.Context, actorID, orgID, returnURL string) (*PortalSession, error) {
	a, err := s.resolveAccess(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(a); err != nil {
		return nil, err
	}
	if a.Account.CustomerID() == "" {
		return nil, utils.E(utils.CodeBillingCustomerMissing)
	}
	if !allowedRedirect(returnURL) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	orgRef := ""
	if a.Account.Organization() {
		orgRef = a.Account.OwnerID
	}
	sess, err := s.provider.CreatePortal(ctx, PortalInput{
		OrganizationID: orgRef, CustomerID: a.Account.CustomerID(), ReturnURL: returnURL,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}

func (s *Service) PlanForOwner(ctx context.Context, ownerType, ownerID string) (string, string, error) {
	acc, err := s.accounts.GetByOwner(ctx, ownerType, ownerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "none", "none", nil
		}
		return "", "", mapRepoErr(err)
	}
	sub, err := s.subs.GetByAccountID(ctx, acc.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "none", "none", nil
		}
		return "", "", mapRepoErr(err)
	}
	return sub.PlanCode, sub.Status, nil
}
