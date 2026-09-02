package billing

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) ListInvoices(ctx context.Context, actorID, orgID, status, cursor string, limit int) ([]Invoice, string, error) {
	a, err := s.resolveAccess(ctx, actorID, orgID)
	if err != nil {
		return nil, "", err
	}
	if err := requireRead(a); err != nil {
		return nil, "", err
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.invoices.ListByAccount(ctx, a.Account.ID, status, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	if items == nil {
		items = []Invoice{}
	}
	return items, next, nil
}

func (s *Service) GetInvoice(ctx context.Context, actorID, orgID, invoiceID string) (*Invoice, error) {
	a, err := s.resolveAccess(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(a); err != nil {
		return nil, err
	}
	inv, err := s.invoices.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if inv.BillingAccountID != a.Account.ID {
		return nil, utils.E(utils.CodeNotFound)
	}
	return inv, nil
}
