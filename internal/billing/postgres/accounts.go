package postgres

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo { return &AccountRepo{db: db} }

func (r *AccountRepo) GetByID(ctx context.Context, id string) (*billing.BillingAccount, error) {
	var m tables.BillingAccount
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return accountFromModel(&m), nil
}

func (r *AccountRepo) GetByOwner(ctx context.Context, ownerType, ownerID string) (*billing.BillingAccount, error) {
	var m tables.BillingAccount
	if err := use(ctx, r.db).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return accountFromModel(&m), nil
}

func (r *AccountRepo) Insert(ctx context.Context, a *billing.BillingAccount) error {
	m := accountToModel(a)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *AccountRepo) SetProviderCustomer(ctx context.Context, id, customerID string, at time.Time) error {
	res := use(ctx, r.db).Model(&tables.BillingAccount{}).Where("public_id = ?", id).Updates(map[string]any{
		"provider_customer_id": customerID,
		"updated_at":           at.UTC(),
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return billing.ErrNotFound
	}
	return nil
}

func accountToModel(a *billing.BillingAccount) tables.BillingAccount {
	return tables.BillingAccount{
		PublicID: a.ID, OwnerType: a.OwnerType, OwnerID: a.OwnerID, Provider: a.Provider,
		ProviderCustomerID: a.ProviderCustomerID, CreatedAt: a.CreatedAt.UTC(), UpdatedAt: a.UpdatedAt.UTC(),
	}
}

func accountFromModel(m *tables.BillingAccount) *billing.BillingAccount {
	return &billing.BillingAccount{
		ID: m.PublicID, OwnerType: m.OwnerType, OwnerID: m.OwnerID, Provider: m.Provider,
		ProviderCustomerID: m.ProviderCustomerID, CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

var _ billing.AccountRepo = (*AccountRepo)(nil)
