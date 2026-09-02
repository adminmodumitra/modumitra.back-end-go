package device

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	devices DeviceRepo
	ents    EntitlementRepo
	tx      TxRunner
	orgs    OrgLookup
	signer  Signer
	clock   Clock
	ids     IDGen
	outbox  outbox.Writer
}

func NewService(
	devices DeviceRepo,
	ents EntitlementRepo,
	tx TxRunner,
	orgs OrgLookup,
	signer Signer,
	clock Clock,
	ids IDGen,
	ob outbox.Writer,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if signer == nil {
		s := NewHMACSigner(DefaultSigningKey, KeyIDV1)
		signer = s
	}
	return &Service{
		devices: devices,
		ents:    ents,
		tx:      tx,
		orgs:    orgs,
		signer:  signer,
		clock:   clock,
		ids:     ids,
		outbox:  ob,
	}
}

func NewMemoryService(orgs OrgLookup, signingKey string) (*Service, *Memory) {
	mem := NewMemory()
	svc := NewService(mem.Devices(), mem.Entitlements(), mem, orgs, NewHMACSigner(signingKey, KeyIDV1), SystemClock{}, NewIDGen(), outbox.NewMemoryWriter())
	return svc, mem
}

func validName(s string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= 1 && n <= 80
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}

func (s *Service) ownedDevice(ctx context.Context, userID, deviceID string) (*Device, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if deviceID == "" {
		return nil, utils.E(utils.CodeNotFound)
	}
	d, err := s.devices.GetByID(ctx, deviceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if d.UserID != userID {
		return nil, utils.E(utils.CodeNotFound)
	}
	return d, nil
}

func (s *Service) List(ctx context.Context, userID string, includeDeactivated bool, cursor string, limit int) ([]Device, string, int, error) {
	if userID == "" {
		return nil, "", 0, utils.E(utils.CodeUnauthenticated)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.devices.ListByUser(ctx, userID, includeDeactivated, cursor, limit)
	if err != nil {
		return nil, "", 0, mapRepoErr(err)
	}
	n, err := s.devices.CountActive(ctx, userID)
	if err != nil {
		return nil, "", 0, mapRepoErr(err)
	}
	if items == nil {
		items = []Device{}
	}
	return items, next, n, nil
}

func (s *Service) Activate(ctx context.Context, userID, name, platform string) (*Device, *Snapshot, error) {
	if userID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	name = strings.TrimSpace(name)
	if !validName(name) {
		return nil, nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "name"})
	}
	if !ValidPlatform(platform) {
		return nil, nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "platform"})
	}
	now := s.clock.Now().UTC()
	var outDev *Device
	var outSnap *Snapshot
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		n, err := s.devices.CountActive(ctx, userID)
		if err != nil {
			return mapRepoErr(err)
		}
		if n >= DeviceLimit {
			return utils.E(utils.CodeDeviceLimit)
		}
		d := &Device{
			ID:         s.ids.Device(),
			UserID:     userID,
			Name:       name,
			Platform:   platform,
			Status:     StatusActive,
			LastSeenAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.devices.Insert(ctx, d); err != nil {
			return mapRepoErr(err)
		}
		snap, err := s.issueLocked(ctx, userID, "", d.ID, now)
		if err != nil {
			return err
		}
		outDev = d
		outSnap = snap
		return outbox.Enqueue(ctx, s.outbox, outbox.DeviceActivated(ctx, d.ID, userID, d.Platform, now))
	})
	if err != nil {
		return nil, nil, err
	}
	return outDev, outSnap, nil
}

func (s *Service) Rename(ctx context.Context, userID, deviceID, name string) (*Device, error) {
	name = strings.TrimSpace(name)
	if !validName(name) {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "name"})
	}
	d, err := s.ownedDevice(ctx, userID, deviceID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	if err := s.devices.UpdateName(ctx, d.ID, name, now); err != nil {
		return nil, mapRepoErr(err)
	}
	d.Name = name
	d.UpdatedAt = now
	return d, nil
}

func (s *Service) Deactivate(ctx context.Context, userID, deviceID string) (*Device, error) {
	d, err := s.ownedDevice(ctx, userID, deviceID)
	if err != nil {
		return nil, err
	}
	if d.Status == StatusDeactivated {
		return d, nil
	}
	now := s.clock.Now().UTC()
	if err := s.devices.Deactivate(ctx, d.ID, ReasonUser, now); err != nil {
		return nil, mapRepoErr(err)
	}
	d.Status = StatusDeactivated
	d.DeactivatedAt = &now
	d.DeactivatedReason = ReasonUser
	d.UpdatedAt = now
	return d, nil
}

func (s *Service) Issue(ctx context.Context, userID, orgID, deviceID string) (*Snapshot, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if orgID == "" {
		return nil, utils.E(utils.CodeValidation)
	}
	now := s.clock.Now().UTC()
	var snap *Snapshot
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		var err error
		snap, err = s.issueLocked(ctx, userID, orgID, deviceID, now)
		return err
	})
	return snap, err
}

func (s *Service) Capabilities(ctx context.Context, userID, orgID string) (*CapabilitiesResult, error) {
	if userID == "" {
		return nil, utils.E(utils.CodeUnauthenticated)
	}
	if strings.TrimSpace(orgID) == "" {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "organization_id"})
	}
	if s.orgs == nil {
		return nil, utils.E(utils.CodeServiceUnavailable)
	}
	o, m, err := s.orgs.MembershipOf(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if m.Status == org.MemberSuspended {
		return nil, utils.E(utils.CodeForbidden)
	}
	if !m.Active() {
		return nil, utils.E(utils.CodeNotFound)
	}
	now := s.clock.Now().UTC()
	plan, status := planOf(o, now)
	return &CapabilitiesResult{
		UserID:             userID,
		OrganizationID:     o.ID,
		Role:               m.Role,
		MembershipStatus:   m.Status,
		Plan:               plan,
		SubscriptionStatus: status,
		Capabilities:       Intersect(m.Role, plan),
		IssuedAt:           now,
	}, nil
}

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
