package httpapi

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func metaFrom(c *gin.Context) utils.Meta {
	return utils.Meta{
		RequestID: middleware.RequestIDFrom(c),
		TraceID:   middleware.TraceIDFrom(c),
	}
}

type HealthOK struct {
	utils.Meta
	Status string `json:"status"`
}

type ReadyOK struct {
	utils.Meta
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

type PublicConfigOK struct {
	utils.Meta
	TermsVersion   string   `json:"terms_version"`
	PrivacyVersion string   `json:"privacy_version"`
	OAuthProviders []string `json:"oauth_providers"`
	MinAppVersion  string   `json:"min_app_version"`
	TrialDays      int      `json:"trial_days"`
}

type UserView struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"display_name"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type UserOK struct {
	utils.Meta
	User UserView `json:"user"`
}

// StatusOK is a success body with a status string (accepted, logged_out, ...).
type StatusOK struct {
	utils.Meta
	Status string `json:"status"`
}

type LogoutAllOK struct {
	utils.Meta
	Status          string `json:"status"`
	RevokedSessions int    `json:"revoked_sessions"`
}

type TokenOK struct {
	utils.Meta
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type OrganizationView struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	OwnerUserID string  `json:"owner_user_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	TrialEndsAt *string `json:"trial_ends_at"`
}

type MembershipView struct {
	UserID         string          `json:"user_id"`
	OrganizationID string          `json:"organization_id"`
	Role           string          `json:"role"`
	Status         string          `json:"status"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
	User           *MemberUserView `json:"user,omitempty"`
}

type MemberUserView struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
}

type OrgCreatedOK struct {
	utils.Meta
	Organization OrganizationView `json:"organization"`
	Membership   MembershipView   `json:"membership"`
	TrialEndsAt  string           `json:"trial_ends_at"`
}

type OrgListItem struct {
	Organization     OrganizationView `json:"organization"`
	Role             string           `json:"role"`
	MembershipStatus string           `json:"membership_status"`
}

type OrgListOK struct {
	utils.Meta
	Items      []OrgListItem `json:"items"`
	NextCursor *string       `json:"next_cursor"`
}

type OrgOK struct {
	utils.Meta
	Organization OrganizationView `json:"organization"`
}

type OrgDeletedOK struct {
	utils.Meta
	Status string `json:"status"`
	ID     string `json:"id"`
}

type LeaveOK struct {
	utils.Meta
	Status         string `json:"status"`
	OrganizationID string `json:"organization_id"`
}

type TransferOK struct {
	utils.Meta
	Organization        OrganizationView `json:"organization"`
	PreviousOwnerUserID string           `json:"previous_owner_user_id"`
	NewOwnerUserID      string           `json:"new_owner_user_id"`
}

type RoleCatalogOK struct {
	utils.Meta
	SchemaVersion int                        `json:"schema_version"`
	Roles         map[string]map[string]bool `json:"roles"`
}

type MemberListOK struct {
	utils.Meta
	Items      []MembershipView `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

type MemberOK struct {
	utils.Meta
	Membership MembershipView `json:"membership"`
}

type MemberRemovedOK struct {
	utils.Meta
	Status         string `json:"status"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

type InvitationView struct {
	ID              string  `json:"id"`
	OrganizationID  string  `json:"organization_id"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	Status          string  `json:"status"`
	InvitedByUserID string  `json:"invited_by_user_id"`
	CreatedAt       string  `json:"created_at"`
	ExpiresAt       string  `json:"expires_at"`
	AcceptedAt      *string `json:"accepted_at"`
}

type InvitationOK struct {
	utils.Meta
	Invitation InvitationView `json:"invitation"`
}

type InvitationListOK struct {
	utils.Meta
	Items      []InvitationView `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

type InvitationRevokedOK struct {
	utils.Meta
	Status string `json:"status"`
	ID     string `json:"id"`
}

type InvitationResentOK struct {
	utils.Meta
	Status    string `json:"status"`
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

type InvitePreviewOK struct {
	utils.Meta
	OrganizationName   string `json:"organization_name"`
	Role               string `json:"role"`
	InviterDisplayName string `json:"inviter_display_name"`
	Email              string `json:"email"`
	ExpiresAt          string `json:"expires_at"`
	Status             string `json:"status"`
}

type InviteAcceptOK struct {
	utils.Meta
	Membership   MembershipView   `json:"membership"`
	Organization OrganizationView `json:"organization"`
}

type LoginOK struct {
	TokenOK
	User         UserView          `json:"user"`
	Organization *OrganizationView `json:"organization"`
}

func userView(u *identity.User) UserView {
	if u == nil {
		return UserView{}
	}
	return UserView{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified(),
		DisplayName:   u.DisplayName,
		CreatedAt:     u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func rfcTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func rfcTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func organizationView(o *org.Organization) OrganizationView {
	if o == nil {
		return OrganizationView{}
	}
	return OrganizationView{
		ID:          o.ID,
		Name:        o.Name,
		OwnerUserID: o.OwnerUserID,
		CreatedAt:   rfcTime(o.CreatedAt),
		UpdatedAt:   rfcTime(o.UpdatedAt),
		TrialEndsAt: rfcTimePtr(o.TrialEndsAt),
	}
}

func membershipView(m *org.Membership, u *MemberUserView) MembershipView {
	if m == nil {
		return MembershipView{}
	}
	v := MembershipView{
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		Role:           m.Role,
		Status:         m.Status,
		CreatedAt:      rfcTime(m.CreatedAt),
		UpdatedAt:      rfcTime(m.UpdatedAt),
	}
	if u != nil {
		v.User = u
	}
	return v
}

func invitationView(inv *org.Invitation) InvitationView {
	if inv == nil {
		return InvitationView{}
	}
	return InvitationView{
		ID:              inv.ID,
		OrganizationID:  inv.OrganizationID,
		Email:           inv.Email,
		Role:            inv.Role,
		Status:          inv.Status,
		InvitedByUserID: inv.InvitedByUserID,
		CreatedAt:       rfcTime(inv.CreatedAt),
		ExpiresAt:       rfcTime(inv.ExpiresAt),
		AcceptedAt:      rfcTimePtr(inv.AcceptedAt),
	}
}

func cursorPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
	UserID             string              `json:"user_id"`
	OrganizationID     *string             `json:"organization_id"`
	Plan               string              `json:"plan"`
	SubscriptionStatus string              `json:"subscription_status"`
	Capabilities       device.Capabilities `json:"capabilities"`
	IssuedAt           string              `json:"issued_at"`
	RefreshAfter       string              `json:"refresh_after"`
	ExpiresAt          string              `json:"expires_at"`
	GraceUntil         *string             `json:"grace_until"`
	DeviceID           *string             `json:"device_id"`
	SchemaVersion      int                 `json:"schema_version"`
	Signature          string              `json:"signature"`
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
	UserID             string              `json:"user_id"`
	OrganizationID     string              `json:"organization_id"`
	Role               string              `json:"role"`
	MembershipStatus   string              `json:"membership_status"`
	Plan               string              `json:"plan"`
	SubscriptionStatus string              `json:"subscription_status"`
	Capabilities       device.Capabilities `json:"capabilities"`
	IssuedAt           string              `json:"issued_at"`
}

func deviceView(d *device.Device) DeviceView {
	if d == nil {
		return DeviceView{}
	}
	return DeviceView{
		ID:            d.ID,
		Name:          d.Name,
		Platform:      d.Platform,
		CreatedAt:     rfcTime(d.CreatedAt),
		LastSeenAt:    rfcTime(d.LastSeenAt),
		DeactivatedAt: rfcTimePtr(d.DeactivatedAt),
	}
}

func snapshotView(s *device.Snapshot) EntitlementSnapshotView {
	if s == nil {
		return EntitlementSnapshotView{}
	}
	return EntitlementSnapshotView{
		UserID:             s.UserID,
		OrganizationID:     s.OrganizationID,
		Plan:               s.Plan,
		SubscriptionStatus: s.SubscriptionStatus,
		Capabilities:       s.Capabilities,
		IssuedAt:           rfcTime(s.IssuedAt),
		RefreshAfter:       rfcTime(s.RefreshAfter),
		ExpiresAt:          rfcTime(s.ExpiresAt),
		GraceUntil:         rfcTimePtr(s.GraceUntil),
		DeviceID:           s.DeviceID,
		SchemaVersion:      s.SchemaVersion,
		Signature:          s.Signature,
	}
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

type WebhookOK struct {
	utils.Meta
	Received bool `json:"received"`
}
