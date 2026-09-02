package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type CreateOrgBody struct {
	Name string `json:"name"`
}

type PatchOrgBody struct {
	Name string `json:"name"`
}

type TransferBody struct {
	UserID         string `json:"user_id"`
	NewOwnerUserID string `json:"new_owner_user_id"`
}

type PatchMemberBody struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}

type CreateInviteBody struct {
	Email string `json:"email"`
	Role  string `json:"role"`
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
