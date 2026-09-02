package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type ActivateDeviceBody struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

type PatchDeviceBody struct {
	Name *string `json:"name"`
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

type DeviceView struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Platform      string  `json:"platform"`
	CreatedAt     string  `json:"created_at"`
	LastSeenAt    string  `json:"last_seen_at"`
	DeactivatedAt *string `json:"deactivated_at"`
}

type DeviceListOK struct {
	utils.Meta
	Items       []DeviceView `json:"items"`
	NextCursor  *string      `json:"next_cursor"`
	ActiveCount int          `json:"active_count"`
	DeviceLimit int          `json:"device_limit"`
}

type DeviceOK struct {
	utils.Meta
	Device DeviceView `json:"device"`
}

type DeviceDeactivatedOK struct {
	utils.Meta
	Status string `json:"status"`
	ID     string `json:"id"`
}

type EntitlementSnapshotView struct {
	UserID             string       `json:"user_id"`
	OrganizationID     *string      `json:"organization_id"`
	Plan               string       `json:"plan"`
	SubscriptionStatus string       `json:"subscription_status"`
	Capabilities       Capabilities `json:"capabilities"`
	IssuedAt           string       `json:"issued_at"`
	RefreshAfter       string       `json:"refresh_after"`
	ExpiresAt          string       `json:"expires_at"`
	GraceUntil         *string      `json:"grace_until"`
	DeviceID           *string      `json:"device_id"`
	SchemaVersion      int          `json:"schema_version"`
	Signature          string       `json:"signature"`
}

type ActivateOK struct {
	utils.Meta
	Device              DeviceView              `json:"device"`
	EntitlementSnapshot EntitlementSnapshotView `json:"entitlement_snapshot"`
}

type EntitlementsOK struct {
	utils.Meta
	EntitlementSnapshot EntitlementSnapshotView `json:"entitlement_snapshot"`
}

type CapabilitiesOK struct {
	utils.Meta
	UserID             string       `json:"user_id"`
	OrganizationID     string       `json:"organization_id"`
	Role               string       `json:"role"`
	MembershipStatus   string       `json:"membership_status"`
	Plan               string       `json:"plan"`
	SubscriptionStatus string       `json:"subscription_status"`
	Capabilities       Capabilities `json:"capabilities"`
	IssuedAt           string       `json:"issued_at"`
}
