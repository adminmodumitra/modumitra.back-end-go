package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

type StateRepo struct {
	db *gorm.DB
}

func NewStateRepo(db *gorm.DB) *StateRepo { return &StateRepo{db: db} }

func (r *StateRepo) Insert(ctx context.Context, st *sso.State) error {
	m := stateToModel(st)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *StateRepo) Consume(ctx context.Context, state, provider string, now time.Time) (*sso.State, error) {
	var m StateModel
	db := use(ctx, r.db)
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("state = ? AND provider = ?", state, provider).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sso.ErrInvalidState
		}
		mapped := mapGorm(err)
		if mapped == sso.ErrNotFound {
			return nil, sso.ErrInvalidState
		}
		return nil, mapped
	}
	if m.ConsumedAt != nil || !m.ExpiresAt.After(now) {
		return nil, sso.ErrInvalidState
	}
	t := now.UTC()
	m.ConsumedAt = &t
	if err := db.Model(&StateModel{}).Where("public_id = ?", m.PublicID).Update("consumed_at", t).Error; err != nil {
		return nil, mapGorm(err)
	}
	return stateFromModel(&m), nil
}

var _ sso.StateRepo = (*StateRepo)(nil)
