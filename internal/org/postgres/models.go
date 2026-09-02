package postgres

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func orgToModel(o *org.Organization) tables.Organization {
	return tables.Organization{
		PublicID:           o.ID,
		Name:               o.Name,
		OwnerUserID:        o.OwnerUserID,
		Status:             o.Status,
		Provider:           emptyToNil(o.Provider),
		ProviderCustomerID: emptyToNil(o.ProviderCustomerID),
		TrialEndsAt:        utcPtr(o.TrialEndsAt),
		DeletedAt:          utcPtr(o.DeletedAt),
		CreatedAt:          o.CreatedAt.UTC(),
		UpdatedAt:          o.UpdatedAt.UTC(),
	}
}

func orgFromModel(m *tables.Organization) *org.Organization {
	return &org.Organization{
		ID:                 m.PublicID,
		Name:               m.Name,
		OwnerUserID:        m.OwnerUserID,
		Status:             m.Status,
		Provider:           deref(m.Provider),
		ProviderCustomerID: deref(m.ProviderCustomerID),
		TrialEndsAt:        utcPtr(m.TrialEndsAt),
		DeletedAt:          utcPtr(m.DeletedAt),
		CreatedAt:          m.CreatedAt.UTC(),
		UpdatedAt:          m.UpdatedAt.UTC(),
	}
}

func memberToModel(m *org.Membership) tables.Membership {
	return tables.Membership{
		PublicID:       m.ID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Role:           m.Role,
		State:          m.Status,
		InvitedAt:      utcPtr(m.InvitedAt),
		ActivatedAt:    utcPtr(m.ActivatedAt),
		SuspendedAt:    utcPtr(m.SuspendedAt),
		RemovedAt:      utcPtr(m.RemovedAt),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
	}
}

func memberFromModel(m *tables.Membership) *org.Membership {
	return &org.Membership{
		ID:             m.PublicID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Role:           m.Role,
		Status:         m.State,
		InvitedAt:      utcPtr(m.InvitedAt),
		ActivatedAt:    utcPtr(m.ActivatedAt),
		SuspendedAt:    utcPtr(m.SuspendedAt),
		RemovedAt:      utcPtr(m.RemovedAt),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
	}
}

func inviteToModel(inv *org.Invitation) tables.Invitation {
	return tables.Invitation{
		PublicID:         inv.ID,
		OrganizationID:   inv.OrganizationID,
		Email:            inv.Email,
		Role:             inv.Role,
		Status:           inv.Status,
		TokenHash:        inv.TokenHash,
		InvitedByUserID:  inv.InvitedByUserID,
		ExpiresAt:        inv.ExpiresAt.UTC(),
		AcceptedAt:       utcPtr(inv.AcceptedAt),
		AcceptedByUserID: inv.AcceptedByUserID,
		RevokedAt:        utcPtr(inv.RevokedAt),
		RevokedByUserID:  inv.RevokedByUserID,
		ResentAt:         utcPtr(inv.ResentAt),
		CreatedAt:        inv.CreatedAt.UTC(),
		UpdatedAt:        inv.UpdatedAt.UTC(),
	}
}

func inviteFromModel(m *tables.Invitation) *org.Invitation {
	return &org.Invitation{
		ID:               m.PublicID,
		OrganizationID:   m.OrganizationID,
		Email:            m.Email,
		Role:             m.Role,
		Status:           m.Status,
		TokenHash:        m.TokenHash,
		InvitedByUserID:  m.InvitedByUserID,
		ExpiresAt:        m.ExpiresAt.UTC(),
		AcceptedAt:       utcPtr(m.AcceptedAt),
		AcceptedByUserID: m.AcceptedByUserID,
		RevokedAt:        utcPtr(m.RevokedAt),
		RevokedByUserID:  m.RevokedByUserID,
		ResentAt:         utcPtr(m.ResentAt),
		CreatedAt:        m.CreatedAt.UTC(),
		UpdatedAt:        m.UpdatedAt.UTC(),
	}
}
