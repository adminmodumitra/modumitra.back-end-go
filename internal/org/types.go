package org

import "time"

const (
	RoleOwner         = "owner"
	RoleAdmin         = "admin"
	RoleBillingAdmin  = "billing_admin"
	RoleSalesDesigner = "sales_designer"
	RoleEngineer      = "engineer"
	RoleProduction    = "production"
	RoleViewer        = "viewer"

	MemberActive    = "active"
	MemberSuspended = "suspended"
	MemberInvited   = "invited"
	MemberRemoved   = "removed"

	OrgActive          = "active"
	OrgPendingDeletion = "pending_deletion"
	OrgDeleted         = "deleted"

	InvitePending  = "pending"
	InviteAccepted = "accepted"
	InviteRevoked  = "revoked"
	InviteExpired  = "expired"

	PrefixOrg        = "org_"
	PrefixMember     = "mbr_"
	PrefixInvitation = "inv_"
	PrefixInviteTok  = "invt_"

	InviteTTL    = 7 * 24 * time.Hour
	DefaultLimit = 50
	MaxLimit     = 100
)

type Organization struct {
	ID                 string
	Name               string
	OwnerUserID        string
	Status             string
	TrialEndsAt        *time.Time
	Provider           string
	ProviderCustomerID string
	DeletedAt          *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (o *Organization) Deleted() bool {
	return o == nil || o.Status == OrgDeleted || o.DeletedAt != nil
}

type Membership struct {
	ID             string
	OrganizationID string
	UserID         string
	Role           string
	Status         string
	InvitedAt      *time.Time
	ActivatedAt    *time.Time
	SuspendedAt    *time.Time
	RemovedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m *Membership) Active() bool {
	return m != nil && m.Status == MemberActive
}

type Invitation struct {
	ID               string
	OrganizationID   string
	Email            string
	Role             string
	Status           string
	TokenHash        string
	InvitedByUserID  string
	ExpiresAt        time.Time
	AcceptedAt       *time.Time
	AcceptedByUserID *string
	RevokedAt        *time.Time
	RevokedByUserID  *string
	ResentAt         *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MembershipItem struct {
	Organization Organization
	Membership   Membership
}

type InvitePreview struct {
	OrganizationName   string
	Role               string
	InviterDisplayName string
	Email              string
	ExpiresAt          time.Time
	Status             string
}

type Settings struct {
	TrialDays int
}

func ValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleBillingAdmin, RoleSalesDesigner, RoleEngineer, RoleProduction, RoleViewer:
		return true
	default:
		return false
	}
}

func AssignableRole(role string) bool {
	return ValidRole(role) && role != RoleOwner
}
