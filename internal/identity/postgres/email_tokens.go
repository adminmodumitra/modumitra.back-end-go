package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

type EmailTokenRepo struct {
	db *gorm.DB
}

func NewEmailTokenRepo(db *gorm.DB) *EmailTokenRepo { return &EmailTokenRepo{db: db} }

func (r *EmailTokenRepo) Insert(ctx context.Context, t *identity.EmailToken) error {
	m := tokenToModel(t)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *EmailTokenRepo) GetByHash(ctx context.Context, hash string) (*identity.EmailToken, error) {
	var m EmailTokenModel
	if err := use(ctx, r.db).Where("token_hash = ?", hash).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return tokenFromModel(&m), nil
}

func (r *EmailTokenRepo) Consume(ctx context.Context, id string, at time.Time) error {
	res := use(ctx, r.db).Model(&EmailTokenModel{}).
		Where("public_id = ? AND consumed_at IS NULL", id).
		Update("consumed_at", at)
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *EmailTokenRepo) InvalidateOpen(ctx context.Context, userID, purpose string, at time.Time) error {
	if err := use(ctx, r.db).Model(&EmailTokenModel{}).
		Where("user_id = ? AND purpose = ? AND consumed_at IS NULL", userID, purpose).
		Update("consumed_at", at).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

var _ identity.EmailTokenRepo = (*EmailTokenRepo)(nil)
