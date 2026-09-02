package postgres

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type InvoiceRepo struct {
	db *gorm.DB
}

func NewInvoiceRepo(db *gorm.DB) *InvoiceRepo { return &InvoiceRepo{db: db} }

func (r *InvoiceRepo) Insert(ctx context.Context, inv *billing.Invoice) error {
	m := invToModel(inv)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id string) (*billing.Invoice, error) {
	var m tables.Invoice
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return invFromModel(&m), nil
}

func (r *InvoiceRepo) GetByProviderInvoiceID(ctx context.Context, providerInvoiceID string) (*billing.Invoice, error) {
	var m tables.Invoice
	if err := use(ctx, r.db).Where("provider_invoice_id = ?", providerInvoiceID).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return invFromModel(&m), nil
}

func (r *InvoiceRepo) ListByAccount(ctx context.Context, accountID, status, cursor string, limit int) ([]billing.Invoice, string, error) {
	return r.list(ctx, "billing_account_id = ?", accountID, status, cursor, limit)
}

func (r *InvoiceRepo) ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]billing.Invoice, string, error) {
	return r.list(ctx, "organization_id = ?", orgID, status, cursor, limit)
}

func (r *InvoiceRepo) list(ctx context.Context, col, id, status, cursor string, limit int) ([]billing.Invoice, string, error) {
	ct, cid, err := billing.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := use(ctx, r.db).Model(&tables.Invoice{}).Where(col, id)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if cid != "" {
		q = q.Where("(created_at, public_id) < (?, ?)", ct, cid)
	}
	var rows []tables.Invoice
	if err := q.Order("created_at DESC, public_id DESC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, "", mapGorm(err)
	}
	out := make([]billing.Invoice, 0, len(rows))
	for i := range rows {
		out = append(out, *invFromModel(&rows[i]))
	}
	next := ""
	if limit > 0 && len(out) > limit {
		last := out[limit-1]
		next = billing.EncodeCursor(last.CreatedAt, last.ID)
		out = out[:limit]
	}
	return out, next, nil
}

var _ billing.InvoiceRepo = (*InvoiceRepo)(nil)
