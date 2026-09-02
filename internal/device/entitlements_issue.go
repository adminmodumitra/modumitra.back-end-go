package device

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) issueLocked(ctx context.Context, userID, orgID, deviceID string, now time.Time) (*Snapshot, error) {
	var (
		o   *org.Organization
		m   *org.Membership
		err error
	)
	if s.orgs != nil {
		if orgID != "" {
			o, m, err = s.orgs.MembershipOf(ctx, orgID, userID)
			if err != nil {
				return nil, err
			}
			if !m.Active() {
				return nil, utils.E(utils.CodeNotFound)
			}
		} else {
			o, err = s.orgs.PrimaryForUser(ctx, userID)
			if err != nil {
				return nil, mapRepoErr(err)
			}
			if o != nil {
				_, m, err = s.orgs.MembershipOf(ctx, o.ID, userID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	var deviceIDPtr *string
	if deviceID != "" {
		d, err := s.devices.GetByID(ctx, deviceID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, utils.E(utils.CodeDeviceNotOwned)
			}
			return nil, mapRepoErr(err)
		}
		if d.UserID != userID || d.Status != StatusActive {
			return nil, utils.E(utils.CodeDeviceNotOwned)
		}
		id := d.ID
		deviceIDPtr = &id
	}

	plan, status := PlanNone, SubNone
	role := ""
	var orgIDPtr *string
	if o != nil {
		id := o.ID
		orgIDPtr = &id
		plan, status = planOf(o, now)
		if m != nil {
			role = m.Role
		}
	}

	snap := Snapshot{
		UserID:             userID,
		OrganizationID:     orgIDPtr,
		Plan:               plan,
		SubscriptionStatus: status,
		Capabilities:       Intersect(role, plan),
		IssuedAt:           now,
		RefreshAfter:       now.Add(RefreshAfter),
		ExpiresAt:          now.Add(ExpiresAfter),
		DeviceID:           deviceIDPtr,
		SchemaVersion:      SchemaVersion,
	}
	if plan != PlanNone {
		g := now.Add(GraceAfter)
		snap.GraceUntil = &g
	}

	canon, err := CanonicalSnapshot(snap)
	if err != nil {
		return nil, utils.Generic(err)
	}
	snap.Signature = s.signer.Sign(canon)

	row := &Entitlement{
		ID:                 s.ids.Entitlement(),
		UserID:             snap.UserID,
		OrganizationID:     snap.OrganizationID,
		DeviceID:           snap.DeviceID,
		Plan:               snap.Plan,
		SubscriptionStatus: snap.SubscriptionStatus,
		Capabilities:       snap.Capabilities,
		IssuedAt:           snap.IssuedAt,
		RefreshAfter:       snap.RefreshAfter,
		ExpiresAt:          snap.ExpiresAt,
		GraceUntil:         snap.GraceUntil,
		SchemaVersion:      snap.SchemaVersion,
		KeyID:              s.signer.KeyID(),
		Signature:          snap.Signature,
		CreatedAt:          now,
	}
	if err := s.ents.Insert(ctx, row); err != nil {
		return nil, mapRepoErr(err)
	}
	return &snap, nil
}
