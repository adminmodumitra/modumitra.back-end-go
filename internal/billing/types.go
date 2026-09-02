package billing

import "time"

const (
	ProviderFake   = "fake"
	ProviderStripe = "stripe"

	PlanTrial      = "trial"
	PlanStudio     = "studio"
	PlanTeam       = "team"
	PlanEnterprise = "enterprise"

	PlanIDTrial  = "plan_trial"
	PlanIDStudio = "plan_studio_monthly"
	PlanIDTeam   = "plan_team_monthly"

	IntervalNone  = "none"
	IntervalMonth = "month"
	IntervalYear  = "year"

	StatusTrialing  = "trialing"
	StatusActive    = "active"
	StatusPastDue   = "past_due"
	StatusGrace     = "grace"
	StatusSuspended = "suspended"
	StatusCanceled  = "canceled"

	InvoiceDraft         = "draft"
	InvoiceOpen          = "open"
	InvoicePaid          = "paid"
	InvoiceVoid          = "void"
	InvoiceUncollectible = "uncollectible"

	EventReceived  = "received"
	EventProcessed = "processed"
	EventIgnored   = "ignored"
	EventFailed    = "failed"

	EventCheckoutCompleted = "checkout.session.completed"
	EventInvoicePaid       = "invoice.paid"

	PrefixSub      = "sub_"
	PrefixInvoice  = "invc_"
	PrefixEvent    = "bev_"
	PrefixCheckout = "cs_"
	PrefixPortal   = "bps_"
	PrefixCustomer = "cus_fake_"
	PrefixProvSub  = "sub_fake_"
	PrefixProvInv  = "in_fake_"
	PrefixProvEvt  = "evt_fake_"

	DefaultLimit = 50
	MaxLimit     = 100
	SessionTTL   = time.Hour
)
