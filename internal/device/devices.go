package device

import (
	"context"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
