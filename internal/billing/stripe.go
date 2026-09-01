package billing

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// StripeProvider is a stub. It never calls Stripe APIs.
type StripeProvider struct{}

func NewStripeProvider() StripeProvider { return StripeProvider{} }

func (StripeProvider) Name() string { return ProviderStripe }

func (StripeProvider) CreateCheckout(context.Context, CheckoutInput) (CheckoutSession, error) {
	return CheckoutSession{}, utils.E(utils.CodeServiceUnavailable).WithMessage("stripe checkout is not implemented")
}

func (StripeProvider) CreatePortal(context.Context, PortalInput) (PortalSession, error) {
	return PortalSession{}, utils.E(utils.CodeServiceUnavailable).WithMessage("stripe portal is not implemented")
}

func (StripeProvider) Cancel(context.Context, string, bool) error {
	return utils.E(utils.CodeServiceUnavailable).WithMessage("stripe cancel is not implemented")
}

func (StripeProvider) Resume(context.Context, string) error {
	return utils.E(utils.CodeServiceUnavailable).WithMessage("stripe resume is not implemented")
}

var _ Provider = StripeProvider{}
