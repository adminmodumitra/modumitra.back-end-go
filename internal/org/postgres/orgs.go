package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

type OrgRepo struct {
	db *gorm.DB
}

func NewOrgRepo(db *gorm.DB) *OrgRepo { return &OrgRepo{db: db} }

func (r *OrgRepo) Insert(ctx context.Context, o *org.Organization) error {
	m := orgToModel(o)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *OrgRepo) GetByID(ctx context.Context, id string) (*org.Organization, error) {
	var m OrganizationModel
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return orgFromModel(&m), nil
}

func (r *OrgRepo) UpdateName(ctx context.Context, id, name string, at time.Time) error {
	res := use(ctx, r.db).Model(&OrganizationModel{}).
		Where("public_id = ? AND status <> ?", id, org.OrgDeleted).
		Updates(map[string]any{
			"name":       name,
			"updated_at": at,
		})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *OrgRepo) SoftDelete(ctx context.Context, id string, at time.Time) error {
	res := use(ctx, r.db).Model(&OrganizationModel{}).
		Where("public_id = ? AND status <> ?", id, org.OrgDeleted).
		Updates(map[string]any{
			"status":     org.OrgDeleted,
			"deleted_at": at,
			"updated_at": at,
		})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return org.ErrNotFound
	}
	return nil
}

func (r *OrgRepo) SetOwner(ctx context.Context, id, ownerUserID string, at time.Time) error {
	res := use(ctx, r.db).Model(&OrganizationModel{}).
		Where("public_id = ? AND status <> ?", id, org.OrgDeleted).
		Updates(map[string]any{
			"owner_user_id": ownerUserID,
			"updated_at":    at,
		})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return org.ErrNotFound
	}
	return nil
}

var _ org.OrgRepo = (*OrgRepo)(nil)
