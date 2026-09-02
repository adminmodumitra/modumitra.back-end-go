package postgres

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type SubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) *SubscriptionRepo { return &SubscriptionRepo{db: db} }

func (r *SubscriptionRepo) GetByOrg(ctx context.Context, orgID string) (*billing.Subscription, error) {
	var m tables.Subscription
	if err := use(ctx, r.db).Where("organization_id = ?", orgID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return subFromModel(&m), nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id string) (*billing.Subscription, error) {
	var m tables.Subscription
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return subFromModel(&m), nil
}

func (r *SubscriptionRepo) Insert(ctx context.Context, s *billing.Subscription) error {
	m := subToModel(s)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, s *billing.Subscription) error {
	m := subToModel(s)
	res := use(ctx, r.db).Model(&tables.Subscription{}).Where("public_id = ?", s.ID).Updates(map[string]any{
		"plan_id":                  m.PlanID,
		"plan_code":                m.PlanCode,
		"status":                   m.Status,
		"provider":                 m.Provider,
		"provider_subscription_id": m.ProviderSubscriptionID,
		"seats_included":           m.SeatsIncluded,
		"trial_ends_at":            m.TrialEndsAt,
		"current_period_start":     m.CurrentPeriodStart,
		"current_period_end":       m.CurrentPeriodEnd,
		"cancel_at_period_end":     m.CancelAtPeriodEnd,
		"canceled_at":              m.CanceledAt,
		"grace_until":              m.GraceUntil,
		"updated_at":               m.UpdatedAt,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return billing.ErrNotFound
	}
	return nil
}

var _ billing.SubscriptionRepo = (*SubscriptionRepo)(nil)
