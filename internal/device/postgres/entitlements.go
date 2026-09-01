package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
)

type EntitlementRepo struct {
	db *gorm.DB
}

func NewEntitlementRepo(db *gorm.DB) *EntitlementRepo { return &EntitlementRepo{db: db} }

func (r *EntitlementRepo) Insert(ctx context.Context, e *device.Entitlement) error {
	m := entitlementToModel(e)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

var _ device.EntitlementRepo = (*EntitlementRepo)(nil)
