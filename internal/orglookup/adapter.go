package orglookup

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

type Adapter struct{ *org.Service }

func (a Adapter) PrimaryForUser(ctx context.Context, userID string) (*device.OrgRef, error) {
	o, err := a.Service.PrimaryForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toOrgRef(o), nil
}

func (a Adapter) MembershipOf(ctx context.Context, orgID, userID string) (*device.OrgRef, *device.MemberRef, error) {
	o, m, err := a.Service.MembershipOf(ctx, orgID, userID)
	if err != nil {
		return nil, nil, err
	}
	return toOrgRef(o), toMemberRef(m), nil
}

func toOrgRef(o *org.Organization) *device.OrgRef {
	if o == nil {
		return nil
	}
	return &device.OrgRef{ID: o.ID, TrialEndsAt: o.TrialEndsAt}
}

func toMemberRef(m *org.Membership) *device.MemberRef {
	if m == nil {
		return nil
	}
	return &device.MemberRef{UserID: m.UserID, Role: m.Role, Status: m.Status}
}
