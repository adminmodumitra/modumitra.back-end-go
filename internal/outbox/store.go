package outbox

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

// SQLStore claims unpublished rows with FOR UPDATE SKIP LOCKED.
type SQLStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) ClaimDue(ctx context.Context, now time.Time, limit int) ([]Row, error) {
	if limit <= 0 {
		limit = DefaultBatch
	}
	var models []tables.Outbox
	err := store.GormDB(ctx, s.db).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("published_at IS NULL AND available_at <= ?", now.UTC()).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]Row, 0, len(models))
	for _, m := range models {
		out = append(out, rowFrom(m))
	}
	return out, nil
}

func (s *SQLStore) MarkPublished(ctx context.Context, publicID string, at time.Time) error {
	return store.GormDB(ctx, s.db).Model(&tables.Outbox{}).
		Where("public_id = ?", publicID).
		Updates(map[string]any{
			"published_at": at.UTC(),
			"last_error":   nil,
		}).Error
}

func (s *SQLStore) MarkFailed(ctx context.Context, publicID string, attempts int, lastErr string, availableAt time.Time) error {
	if len(lastErr) > 2000 {
		lastErr = lastErr[:2000]
	}
	return store.GormDB(ctx, s.db).Model(&tables.Outbox{}).
		Where("public_id = ?", publicID).
		Updates(map[string]any{
			"attempts":     attempts,
			"last_error":   lastErr,
			"available_at": availableAt.UTC(),
		}).Error
}

var _ Store = (*SQLStore)(nil)
