package billing

import "context"

type CheckoutInput struct {
	BillingAccountID string
	OrganizationID   string
	PlanID           string
	SuccessURL       string
	CancelURL        string
	CustomerID       string
	Email            string
	Currency         string
	AmountMinor      int
	Seats            int
	PlanCode         string
}

type PortalInput struct {
	OrganizationID string
	CustomerID     string
	ReturnURL      string
}

type Provider interface {
	Name() string // "fake"
	CreateCheckout(ctx context.Context, in CheckoutInput) (CheckoutSession, error)
	CreatePortal(ctx context.Context, in PortalInput) (PortalSession, error)
	Cancel(ctx context.Context, providerSubID string, atPeriodEnd bool) error
	Resume(ctx context.Context, providerSubID string) error
}
