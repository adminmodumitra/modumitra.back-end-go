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

type Capabilities struct {
	CloudProjects      bool `json:"cloud_projects"`
	OrganizationAssets bool `json:"organization_assets"`
	ProposalBranding   bool `json:"proposal_branding"`
	EngineeringHandoff bool `json:"engineering_handoff"`
	MaxSeats           int  `json:"max_seats,omitempty"`
	MaxDevices         int  `json:"max_devices,omitempty"`
	MaxCloudProjects   int  `json:"max_cloud_projects,omitempty"`
}

type Plan struct {
	ID                string
	Code              string
	Name              string
	Interval          string
	Currency          string
	UnitAmountMinor   int
	SeatsIncluded     int
	MaxDevicesPerUser int
	TrialDays         int
	Sellable          bool
	Active            bool
	Capabilities      Capabilities
	Provider          string
	ProviderPriceID   string
	SortOrder         int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Subscription struct {
	ID                     string
	OrganizationID         string
	PlanID                 string
	PlanCode               string
	Status                 string
	Provider               string
	ProviderSubscriptionID string
	SeatsIncluded          int
	TrialEndsAt            *time.Time
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
	GraceUntil             *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Invoice struct {
	ID                string
	OrganizationID    string
	SubscriptionID    *string
	Provider          string
	ProviderInvoiceID string
	Number            *string
	Status            string
	Currency          string
	TotalMinor        int
	AmountDueMinor    int
	HostedInvoiceURL  *string
	InvoicePDFURL     *string
	PeriodStart       *time.Time
	PeriodEnd         *time.Time
	PaidAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Event struct {
	ID              string
	Provider        string
	ProviderEventID string
	Type            string
	OrganizationID  *string
	SubscriptionID  *string
	Status          string
	PayloadDigest   string
	Summary         map[string]any
	Error           string
	ProcessedAt     *time.Time
	CreatedAt       time.Time
}

type CheckoutSession struct {
	ID         string
	URL        string
	ExpiresAt  time.Time
	CustomerID string
}

type PortalSession struct {
	ID        string
	URL       string
	ExpiresAt time.Time
}

type SeatMember struct {
	UserID       string
	Role         string
	Status       string
	ConsumesSeat bool
}

type Seats struct {
	Purchased          int
	Used               int
	PendingInvitations int
	Available          int
	Members            []SeatMember
}

type Preview struct {
	PlanID            string
	Plan              string
	PurchasedSeats    int
	TargetSeats       int
	SeatShortfall     int
	MembersOverTarget []SeatMember
	CapabilityLoss    []string
	Blocked           bool
	Message           string
}

type Settings struct {
	Provider      string
	WebhookSecret string
}
