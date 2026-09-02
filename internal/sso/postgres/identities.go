package postgres

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type IdentityRepo struct {
	db *gorm.DB
}

func NewIdentityRepo(db *gorm.DB) *IdentityRepo { return &IdentityRepo{db: db} }

func (r *IdentityRepo) Insert(ctx context.Context, id *sso.Identity) error {
	m := identityToModel(id)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		if isUnique(err) {
			switch uniqueConstraint(err) {
			case "uniq_oauth_identities_user_provider":
				return utils.E(utils.CodeProviderAlreadyLinked)
			default:
				return utils.E(utils.CodeIdentityAlreadyLinked)
			}
		}
		return mapGorm(err)
	}
	return nil
}

func (r *IdentityRepo) GetByProviderSubject(ctx context.Context, provider, subject string) (*sso.Identity, error) {
	var m tables.OAuthIdentity
	if err := use(ctx, r.db).Where("provider = ? AND provider_user_id = ?", provider, subject).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return identityFromModel(&m), nil
}

func (r *IdentityRepo) GetByUserProvider(ctx context.Context, userID, provider string) (*sso.Identity, error) {
	var m tables.OAuthIdentity
	if err := use(ctx, r.db).Where("user_id = ? AND provider = ?", userID, provider).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return identityFromModel(&m), nil
}

func (r *IdentityRepo) ListByUser(ctx context.Context, userID string) ([]sso.Identity, error) {
	var rows []tables.OAuthIdentity
	if err := use(ctx, r.db).Where("user_id = ?", userID).Order("linked_at ASC").Find(&rows).Error; err != nil {
		return nil, mapGorm(err)
	}
	out := make([]sso.Identity, 0, len(rows))
	for i := range rows {
		out = append(out, *identityFromModel(&rows[i]))
	}
	return out, nil
}

func (r *IdentityRepo) DeleteByUserProvider(ctx context.Context, userID, provider string) error {
	res := use(ctx, r.db).Where("user_id = ? AND provider = ?", userID, provider).Delete(&tables.OAuthIdentity{})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return sso.ErrNotFound
	}
	return nil
}

var _ sso.IdentityRepo = (*IdentityRepo)(nil)
