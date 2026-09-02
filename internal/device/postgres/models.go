package postgres

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
)

func deviceToModel(d *device.Device) tables.Device {
	return tables.Device{
		PublicID:          d.ID,
		UserID:            d.UserID,
		Name:              d.Name,
		Platform:          d.Platform,
		Status:            d.Status,
		AppVersion:        emptyToNil(d.AppVersion),
		LastSeenAt:        d.LastSeenAt.UTC(),
		DeactivatedAt:     utcPtr(d.DeactivatedAt),
		DeactivatedReason: emptyToNil(d.DeactivatedReason),
		CreatedAt:         d.CreatedAt.UTC(),
		UpdatedAt:         d.UpdatedAt.UTC(),
	}
}

func deviceFromModel(m *tables.Device) *device.Device {
	return &device.Device{
		ID:                m.PublicID,
		UserID:            m.UserID,
		Name:              m.Name,
		Platform:          m.Platform,
		Status:            m.Status,
		AppVersion:        deref(m.AppVersion),
		LastSeenAt:        m.LastSeenAt.UTC(),
		DeactivatedAt:     utcPtr(m.DeactivatedAt),
		DeactivatedReason: deref(m.DeactivatedReason),
		CreatedAt:         m.CreatedAt.UTC(),
		UpdatedAt:         m.UpdatedAt.UTC(),
	}
}

func entitlementToModel(e *device.Entitlement) tables.Entitlement {
	return tables.Entitlement{
		PublicID:           e.ID,
		UserID:             e.UserID,
		OrganizationID:     e.OrganizationID,
		DeviceID:           e.DeviceID,
		PlanID:             e.PlanID,
		Plan:               e.Plan,
		SubscriptionStatus: e.SubscriptionStatus,
		Capabilities:       deviceCaps(e.Capabilities),
		IssuedAt:           e.IssuedAt.UTC(),
		RefreshAfter:       e.RefreshAfter.UTC(),
		ExpiresAt:          e.ExpiresAt.UTC(),
		GraceUntil:         utcPtr(e.GraceUntil),
		SchemaVersion:      e.SchemaVersion,
		KeyID:              e.KeyID,
		Signature:          e.Signature,
		CreatedAt:          e.CreatedAt.UTC(),
	}
}

func entitlementFromModel(m *tables.Entitlement) *device.Entitlement {
	return &device.Entitlement{
		ID:                 m.PublicID,
		UserID:             m.UserID,
		OrganizationID:     m.OrganizationID,
		DeviceID:           m.DeviceID,
		PlanID:             m.PlanID,
		Plan:               m.Plan,
		SubscriptionStatus: m.SubscriptionStatus,
		Capabilities:       deviceCapsFrom(m.Capabilities),
		IssuedAt:           m.IssuedAt.UTC(),
		RefreshAfter:       m.RefreshAfter.UTC(),
		ExpiresAt:          m.ExpiresAt.UTC(),
		GraceUntil:         utcPtr(m.GraceUntil),
		SchemaVersion:      m.SchemaVersion,
		KeyID:              m.KeyID,
		Signature:          m.Signature,
		CreatedAt:          m.CreatedAt.UTC(),
	}
}

func deviceCaps(c device.Capabilities) tables.DeviceCapabilities {
	return tables.DeviceCapabilities{
		CloudProjects: c.CloudProjects, OrganizationAssets: c.OrganizationAssets,
		ProposalBranding: c.ProposalBranding, EngineeringHandoff: c.EngineeringHandoff,
		MaxSeats: c.MaxSeats, MaxDevices: c.MaxDevices, MaxCloudProjects: c.MaxCloudProjects,
	}
}

func deviceCapsFrom(c tables.DeviceCapabilities) device.Capabilities {
	return device.Capabilities{
		CloudProjects: c.CloudProjects, OrganizationAssets: c.OrganizationAssets,
		ProposalBranding: c.ProposalBranding, EngineeringHandoff: c.EngineeringHandoff,
		MaxSeats: c.MaxSeats, MaxDevices: c.MaxDevices, MaxCloudProjects: c.MaxCloudProjects,
	}
}
