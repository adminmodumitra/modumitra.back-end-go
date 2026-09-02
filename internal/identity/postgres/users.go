package postgres

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByID(ctx context.Context, id string) (*identity.User, error) {
	var m tables.User
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return userFromModel(&m), nil
}

func (r *UserRepo) GetByEmailNormalized(ctx context.Context, email string) (*identity.User, error) {
	var m tables.User
	if err := use(ctx, r.db).Where("email_normalized = ?", email).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return userFromModel(&m), nil
}

func (r *UserRepo) Insert(ctx context.Context, u *identity.User) error {
	m := userToModel(u)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		if isUnique(err) {
			return utils.E(utils.CodeEmailTaken)
		}
		return mapGorm(err)
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, hash string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.User{}).Where("public_id = ?", id).Updates(map[string]any{
		"password_hash":       hash,
		"password_updated_at": at,
		"updated_at":          at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *UserRepo) SetEmailVerified(ctx context.Context, id string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.User{}).Where("public_id = ?", id).Updates(map[string]any{
		"email_verified_at": at,
		"updated_at":        at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *UserRepo) SetLastLogin(ctx context.Context, id string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.User{}).Where("public_id = ?", id).Updates(map[string]any{
		"last_login_at": at,
		"updated_at":    at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

var _ identity.UserRepo = (*UserRepo)(nil)
