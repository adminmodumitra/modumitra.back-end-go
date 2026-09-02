package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func organizationView(o *org.Organization) dto.OrganizationView {
	if o == nil {
		return dto.OrganizationView{}
	}
	return dto.OrganizationView{
		ID:          o.ID,
		Name:        o.Name,
		OwnerUserID: o.OwnerUserID,
		CreatedAt:   rfcTime(o.CreatedAt),
		UpdatedAt:   rfcTime(o.UpdatedAt),
		TrialEndsAt: rfcTimePtr(o.TrialEndsAt),
	}
}

func membershipView(m *org.Membership, u *dto.MemberUserView) dto.MembershipView {
	if m == nil {
		return dto.MembershipView{}
	}
	v := dto.MembershipView{
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

func invitationView(inv *org.Invitation) dto.InvitationView {
	if inv == nil {
		return dto.InvitationView{}
	}
	return dto.InvitationView{
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

func transferTarget(b dto.TransferBody) string {
	if b.UserID != "" {
		return b.UserID
	}
	return b.NewOwnerUserID
}
