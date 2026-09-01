package device

import "time"

const (
	StatusActive      = "active"
	StatusDeactivated = "deactivated"

	PlatformMacOS   = "macos"
	PlatformWindows = "windows"
	PlatformLinux   = "linux"

	ReasonUser            = "user"
	ReasonReplaced        = "replaced"
	ReasonAdmin           = "admin"
	ReasonAccountDisabled = "account_disabled"

	PlanNone       = "none"
	PlanTrial      = "trial"
	PlanStudio     = "studio"
	PlanTeam       = "team"
	PlanEnterprise = "enterprise"

	SubNone       = "none"
	SubTrialing   = "trialing"
	SubActive     = "active"
	SubPastDue    = "past_due"
	SubGrace      = "grace"
	SubSuspended  = "suspended"
	SubCanceled   = "canceled"
	SubUnpaid     = "unpaid"
	SubIncomplete = "incomplete"

	PrefixDevice      = "dev_"
	PrefixEntitlement = "ent_"

	KeyIDV1       = "v1"
	SchemaVersion = 1
	DeviceLimit   = 3
	DefaultLimit  = 50
	MaxLimit      = 100
	RefreshAfter  = 12 * time.Hour
	ExpiresAfter  = 24 * time.Hour
	GraceAfter    = 7 * 24 * time.Hour

	DefaultSigningKey = "cabinet-studio-entitlement-local-dev-only-change-me"
)

type Device struct {
	ID                string
	UserID            string
	Name              string
	Platform          string
	Status            string
	AppVersion        string
	LastSeenAt        time.Time
	DeactivatedAt     *time.Time
	DeactivatedReason string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (d *Device) Active() bool {
	return d != nil && d.Status == StatusActive
}

type Capabilities struct {
	CloudProjects      bool `json:"cloud_projects"`
	OrganizationAssets bool `json:"organization_assets"`
	ProposalBranding   bool `json:"proposal_branding"`
	EngineeringHandoff bool `json:"engineering_handoff"`
	MaxSeats           int  `json:"max_seats"`
	MaxDevices         int  `json:"max_devices"`
	MaxCloudProjects   int  `json:"max_cloud_projects"`
}

type Snapshot struct {
	UserID             string
	OrganizationID     *string
	Plan               string
	SubscriptionStatus string
	Capabilities       Capabilities
	IssuedAt           time.Time
	RefreshAfter       time.Time
	ExpiresAt          time.Time
	GraceUntil         *time.Time
	DeviceID           *string
	SchemaVersion      int
	Signature          string
}

type Entitlement struct {
	ID                 string
	UserID             string
	OrganizationID     *string
	DeviceID           *string
	PlanID             *string
	Plan               string
	SubscriptionStatus string
	Capabilities       Capabilities
	IssuedAt           time.Time
	RefreshAfter       time.Time
	ExpiresAt          time.Time
	GraceUntil         *time.Time
	SchemaVersion      int
	KeyID              string
	Signature          string
	CreatedAt          time.Time
}

type CapabilitiesResult struct {
	UserID             string
	OrganizationID     string
	Role               string
	MembershipStatus   string
	Plan               string
	SubscriptionStatus string
	Capabilities       Capabilities
	IssuedAt           time.Time
}
