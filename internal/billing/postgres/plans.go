package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type PlanRepo struct {
	db *gorm.DB
}

func NewPlanRepo(db *gorm.DB) *PlanRepo { return &PlanRepo{db: db} }

func (r *PlanRepo) ListSellable(ctx context.Context) ([]billing.Plan, error) {
	var rows []PlanModel
	if err := use(ctx, r.db).Where("sellable = ? AND active = ?", true, true).
		Order("sort_order ASC").Find(&rows).Error; err != nil {
		return nil, mapGorm(err)
	}
	out := make([]billing.Plan, 0, len(rows))
	for i := range rows {
		out = append(out, *planFromModel(&rows[i]))
	}
	return out, nil
}

func (r *PlanRepo) GetByID(ctx context.Context, id string) (*billing.Plan, error) {
	var m PlanModel
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return planFromModel(&m), nil
}

var _ billing.PlanRepo = (*PlanRepo)(nil)
