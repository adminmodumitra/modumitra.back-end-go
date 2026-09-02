package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type CheckoutBody struct {
	PlanID     string `json:"plan_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type PortalBody struct {
	ReturnURL string `json:"return_url"`
}

type CancelBody struct {
	AtPeriodEnd *bool `json:"at_period_end"`
}

type PreviewBody struct {
	PlanID string `json:"plan_id"`
	Seats  *int   `json:"seats"`
}

type PlanCapabilitiesView struct {
	CloudProjects      bool `json:"cloud_projects"`
	OrganizationAssets bool `json:"organization_assets"`
	ProposalBranding   bool `json:"proposal_branding"`
	EngineeringHandoff bool `json:"engineering_handoff"`
}

type PlanView struct {
	ID                string               `json:"id"`
	Plan              string               `json:"plan"`
	Interval          string               `json:"interval"`
	Currency          string               `json:"currency"`
	UnitAmountMinor   int                  `json:"unit_amount_minor"`
	IncludedSeats     int                  `json:"included_seats"`
	MaxDevicesPerUser int                  `json:"max_devices_per_user"`
	Capabilities      PlanCapabilitiesView `json:"capabilities"`
	TrialDays         int                  `json:"trial_days"`
}

type PlanListOK struct {
	utils.Meta
	Items      []PlanView `json:"items"`
	NextCursor *string    `json:"next_cursor"`
}

type SubscriptionView struct {
	ID                     string  `json:"id"`
	OrganizationID         string  `json:"organization_id"`
	PlanID                 string  `json:"plan_id"`
	Plan                   string  `json:"plan"`
	Status                 string  `json:"status"`
	Seats                  int     `json:"seats"`
	CurrentPeriodStart     string  `json:"current_period_start"`
	CurrentPeriodEnd       string  `json:"current_period_end"`
	TrialEndsAt            *string `json:"trial_ends_at"`
	CancelAtPeriodEnd      bool    `json:"cancel_at_period_end"`
	CanceledAt             *string `json:"canceled_at"`
	GraceUntil             *string `json:"grace_until"`
	Provider               string  `json:"provider"`
	ProviderSubscriptionID *string `json:"provider_subscription_id"`
}

type SubscriptionOK struct {
	utils.Meta
	Subscription SubscriptionView `json:"subscription"`
}

type CheckoutOK struct {
	utils.Meta
	CheckoutSessionID string `json:"checkout_session_id"`
	URL               string `json:"url"`
	ExpiresAt         string `json:"expires_at"`
}

type PortalOK struct {
	utils.Meta
	PortalSessionID string `json:"portal_session_id"`
	URL             string `json:"url"`
	ExpiresAt       string `json:"expires_at"`
}

type InvoiceView struct {
	ID               string  `json:"id"`
	OrganizationID   string  `json:"organization_id"`
	Number           *string `json:"number"`
	Status           string  `json:"status"`
	Currency         string  `json:"currency"`
	TotalMinor       int     `json:"total_minor"`
	AmountDueMinor   int     `json:"amount_due_minor"`
	HostedInvoiceURL *string `json:"hosted_invoice_url"`
	InvoicePDFURL    *string `json:"invoice_pdf_url"`
	PeriodStart      *string `json:"period_start"`
	PeriodEnd        *string `json:"period_end"`
	CreatedAt        string  `json:"created_at"`
	PaidAt           *string `json:"paid_at"`
}

type InvoiceListOK struct {
	utils.Meta
	Items      []InvoiceView `json:"items"`
	NextCursor *string       `json:"next_cursor"`
}

type InvoiceOK struct {
	utils.Meta
	Invoice InvoiceView `json:"invoice"`
}

type SeatMemberView struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	ConsumesSeat bool   `json:"consumes_seat"`
}

type SeatsOK struct {
	utils.Meta
	PurchasedSeats     int              `json:"purchased_seats"`
	UsedSeats          int              `json:"used_seats"`
	PendingInvitations int              `json:"pending_invitations"`
	AvailableSeats     int              `json:"available_seats"`
	Members            []SeatMemberView `json:"members"`
}

type PreviewDowngradeOK struct {
	utils.Meta
	PlanID            string           `json:"plan_id"`
	Plan              string           `json:"plan"`
	PurchasedSeats    int              `json:"purchased_seats"`
	TargetSeats       int              `json:"target_seats"`
	SeatShortfall     int              `json:"seat_shortfall"`
	MembersOverTarget []SeatMemberView `json:"members_over_target"`
	CapabilityLoss    []string         `json:"capability_loss"`
	Blocked           bool             `json:"blocked"`
	Message           string           `json:"message"`
}
