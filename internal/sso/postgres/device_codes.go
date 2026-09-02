package postgres

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

type DeviceCodeRepo struct {
	db *gorm.DB
}

func NewDeviceCodeRepo(db *gorm.DB) *DeviceCodeRepo { return &DeviceCodeRepo{db: db} }

func (r *DeviceCodeRepo) Insert(ctx context.Context, d *sso.DeviceCode) error {
	m := deviceToModel(d)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *DeviceCodeRepo) GetByDeviceCode(ctx context.Context, deviceCode string) (*sso.DeviceCode, error) {
	var m DeviceCodeModel
	if err := use(ctx, r.db).Where("device_code = ?", deviceCode).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return deviceFromModel(&m), nil
}

func (r *DeviceCodeRepo) GetByUserCode(ctx context.Context, userCode string) (*sso.DeviceCode, error) {
	var m DeviceCodeModel
	if err := use(ctx, r.db).Where("user_code = ?", strings.ToUpper(userCode)).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return deviceFromModel(&m), nil
}

func (r *DeviceCodeRepo) Update(ctx context.Context, d *sso.DeviceCode) error {
	m := deviceToModel(d)
	res := use(ctx, r.db).Model(&DeviceCodeModel{}).Where("public_id = ?", d.ID).Updates(map[string]any{
		"status":        m.Status,
		"user_id":       m.UserID,
		"last_poll_at":  m.LastPollAt,
		"authorized_at": m.AuthorizedAt,
		"consumed_at":   m.ConsumedAt,
		"updated_at":    m.UpdatedAt,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return sso.ErrNotFound
	}
	return nil
}

var _ sso.DeviceCodeRepo = (*DeviceCodeRepo)(nil)
