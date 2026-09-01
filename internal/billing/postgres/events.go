package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo { return &EventRepo{db: db} }

func (r *EventRepo) Insert(ctx context.Context, e *billing.Event) error {
	m := eventToModel(e)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *EventRepo) GetByProviderEvent(ctx context.Context, provider, providerEventID string) (*billing.Event, error) {
	var m EventModel
	if err := use(ctx, r.db).Where("provider = ? AND provider_event_id = ?", provider, providerEventID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return eventFromModel(&m), nil
}

func (r *EventRepo) MarkProcessed(ctx context.Context, id string, orgID, subID *string, at time.Time) error {
	res := use(ctx, r.db).Model(&EventModel{}).Where("public_id = ?", id).Updates(map[string]any{
		"status":          billing.EventProcessed,
		"organization_id": orgID,
		"subscription_id": subID,
		"processed_at":    at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return billing.ErrNotFound
	}
	return nil
}

func (r *EventRepo) MarkFailed(ctx context.Context, id, errMsg string, at time.Time) error {
	res := use(ctx, r.db).Model(&EventModel{}).Where("public_id = ?", id).Updates(map[string]any{
		"status":       billing.EventFailed,
		"error":        errMsg,
		"processed_at": at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return billing.ErrNotFound
	}
	return nil
}

var _ billing.EventRepo = (*EventRepo)(nil)
