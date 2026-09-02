package billing

import (
	"context"
	"sort"
)

func (m *MemInvoices) Insert(_ context.Context, inv *Invoice) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.invProv[inv.ProviderInvoiceID]; ok {
		return ErrDuplicate
	}
	stored := cloneInvoice(inv)
	m.invoices[inv.ID] = stored
	m.invProv[inv.ProviderInvoiceID] = stored
	return nil
}

func (m *MemInvoices) GetByID(_ context.Context, id string) (*Invoice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invoices[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneInvoice(inv), nil
}

func (m *MemInvoices) GetByProviderInvoiceID(_ context.Context, providerInvoiceID string) (*Invoice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invProv[providerInvoiceID]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneInvoice(inv), nil
}

func (m *MemInvoices) ListByAccount(_ context.Context, accountID, status, cursor string, limit int) ([]Invoice, string, error) {
	return m.list(func(inv *Invoice) bool { return inv.BillingAccountID == accountID }, status, cursor, limit)
}

func (m *MemInvoices) ListByOrg(_ context.Context, orgID, status, cursor string, limit int) ([]Invoice, string, error) {
	return m.list(func(inv *Invoice) bool { return orgID != "" && inv.OrganizationID == orgID }, status, cursor, limit)
}

func (m *MemInvoices) list(match func(*Invoice) bool, status, cursor string, limit int) ([]Invoice, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Invoice
	for _, inv := range m.invoices {
		if !match(inv) {
			continue
		}
		if status != "" && inv.Status != status {
			continue
		}
		if !BeforeCursor(inv.CreatedAt, inv.ID, ct, cid) {
			continue
		}
		out = append(out, *cloneInvoice(inv))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	next := ""
	if limit > 0 && len(out) > limit {
		last := out[limit-1]
		next = EncodeCursor(last.CreatedAt, last.ID)
		out = out[:limit]
	}
	if out == nil {
		out = []Invoice{}
	}
	return out, next, nil
}
