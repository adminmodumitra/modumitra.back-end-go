package org

const (
	ActionOrgRead        = "org.read"
	ActionOrgWrite       = "org.write"
	ActionOrgDelete      = "org.delete"
	ActionOrgLeave       = "org.leave"
	ActionOrgTransfer    = "org.transfer"
	ActionOrgRoles       = "org.roles"
	ActionMembersList    = "members.list"
	ActionMembersWrite   = "members.write"
	ActionMembersSuspend = "members.suspend"
	ActionMembersRemove  = "members.remove"
	ActionInvitesCreate  = "invites.create"
	ActionInvitesList    = "invites.list"
	ActionInvitesRevoke  = "invites.revoke"
	ActionInvitesResend  = "invites.resend"
)

// RolePermits reports whether an active membership role may perform action.
func RolePermits(role, action string) bool {
	switch action {
	case ActionOrgRead, ActionOrgRoles, ActionMembersList, ActionOrgLeave:
		return ValidRole(role)
	case ActionOrgWrite, ActionMembersWrite, ActionMembersSuspend, ActionMembersRemove,
		ActionInvitesCreate, ActionInvitesList, ActionInvitesRevoke, ActionInvitesResend:
		return role == RoleOwner || role == RoleAdmin
	case ActionOrgDelete, ActionOrgTransfer:
		return role == RoleOwner
	default:
		return false
	}
}

// RolesCatalog is the static role → capability map for GET .../roles (contract §4.3).
func RolesCatalog() map[string]map[string]bool {
	return map[string]map[string]bool{
		RoleOwner: {
			"cloud_projects": true, "organization_assets": true, "proposal_branding": true,
			"engineering_handoff": true, "manage_members": true, "manage_billing": true,
			"transfer_ownership": true, "delete_organization": true,
		},
		RoleAdmin: {
			"cloud_projects": true, "organization_assets": true, "proposal_branding": true,
			"engineering_handoff": true, "manage_members": true, "manage_billing": false,
			"transfer_ownership": false, "delete_organization": false,
		},
		RoleBillingAdmin: {
			"cloud_projects": false, "organization_assets": false, "proposal_branding": false,
			"engineering_handoff": false, "manage_members": false, "manage_billing": true,
			"transfer_ownership": false, "delete_organization": false,
		},
		RoleSalesDesigner: {
			"cloud_projects": true, "organization_assets": true, "proposal_branding": true,
			"engineering_handoff": false, "manage_members": false, "manage_billing": false,
			"transfer_ownership": false, "delete_organization": false,
		},
		RoleEngineer: {
			"cloud_projects": true, "organization_assets": true, "proposal_branding": false,
			"engineering_handoff": true, "manage_members": false, "manage_billing": false,
			"transfer_ownership": false, "delete_organization": false,
		},
		RoleProduction: {
			"cloud_projects": true, "organization_assets": true, "proposal_branding": false,
			"engineering_handoff": false, "manage_members": false, "manage_billing": false,
			"transfer_ownership": false, "delete_organization": false,
		},
		RoleViewer: {
			"cloud_projects": false, "organization_assets": false, "proposal_branding": false,
			"engineering_handoff": false, "manage_members": false, "manage_billing": false,
			"transfer_ownership": false, "delete_organization": false,
		},
	}
}
