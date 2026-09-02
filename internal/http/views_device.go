package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
)

func deviceView(d *device.Device) dto.DeviceView {
	if d == nil {
		return dto.DeviceView{}
	}
	return dto.DeviceView{
		ID:            d.ID,
		Name:          d.Name,
		Platform:      d.Platform,
		CreatedAt:     rfcTime(d.CreatedAt),
		LastSeenAt:    rfcTime(d.LastSeenAt),
		DeactivatedAt: rfcTimePtr(d.DeactivatedAt),
	}
}

func snapshotView(s *device.Snapshot) dto.EntitlementSnapshotView {
	if s == nil {
		return dto.EntitlementSnapshotView{}
	}
	return dto.EntitlementSnapshotView{
		UserID:             s.UserID,
		OrganizationID:     s.OrganizationID,
		Plan:               s.Plan,
		SubscriptionStatus: s.SubscriptionStatus,
		Capabilities:       capabilitiesView(s.Capabilities),
		IssuedAt:           rfcTime(s.IssuedAt),
		RefreshAfter:       rfcTime(s.RefreshAfter),
		ExpiresAt:          rfcTime(s.ExpiresAt),
		GraceUntil:         rfcTimePtr(s.GraceUntil),
		DeviceID:           s.DeviceID,
		SchemaVersion:      s.SchemaVersion,
		Signature:          s.Signature,
	}
}
