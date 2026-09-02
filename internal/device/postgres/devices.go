package postgres

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo { return &DeviceRepo{db: db} }

func (r *DeviceRepo) Insert(ctx context.Context, d *device.Device) error {
	m := deviceToModel(d)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *DeviceRepo) GetByID(ctx context.Context, id string) (*device.Device, error) {
	var m tables.Device
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return deviceFromModel(&m), nil
}

func (r *DeviceRepo) ListByUser(ctx context.Context, userID string, includeDeactivated bool, cursor string, limit int) ([]device.Device, string, error) {
	ct, cid, err := device.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := use(ctx, r.db).Model(&tables.Device{}).Where("user_id = ?", userID)
	if !includeDeactivated {
		q = q.Where("status = ?", device.StatusActive)
	}
	if cid != "" {
		q = q.Where("(created_at, public_id) < (?, ?)", ct, cid)
	}
	var rows []tables.Device
	if err := q.Order("created_at DESC, public_id DESC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, "", mapGorm(err)
	}
	out := make([]device.Device, 0, len(rows))
	for i := range rows {
		out = append(out, *deviceFromModel(&rows[i]))
	}
	next := ""
	if limit > 0 && len(out) > limit {
		last := out[limit-1]
		next = device.EncodeCursor(last.CreatedAt, last.ID)
		out = out[:limit]
	}
	return out, next, nil
}

func (r *DeviceRepo) CountActive(ctx context.Context, userID string) (int, error) {
	var n int64
	if err := use(ctx, r.db).Model(&tables.Device{}).
		Where("user_id = ? AND status = ?", userID, device.StatusActive).
		Count(&n).Error; err != nil {
		return 0, mapGorm(err)
	}
	return int(n), nil
}

func (r *DeviceRepo) UpdateName(ctx context.Context, id, name string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.Device{}).Where("public_id = ?", id).Updates(map[string]any{
		"name":       name,
		"updated_at": at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return device.ErrNotFound
	}
	return nil
}

func (r *DeviceRepo) Deactivate(ctx context.Context, id, reason string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.Device{}).Where("public_id = ?", id).Updates(map[string]any{
		"status":             device.StatusDeactivated,
		"deactivated_at":     at,
		"deactivated_reason": reason,
		"updated_at":         at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return device.ErrNotFound
	}
	return nil
}

var _ device.DeviceRepo = (*DeviceRepo)(nil)
