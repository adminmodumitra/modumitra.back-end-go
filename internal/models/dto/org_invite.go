package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

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
