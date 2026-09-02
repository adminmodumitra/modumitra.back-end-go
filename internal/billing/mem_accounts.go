package billing

import (
	"context"
	"time"
)

type MemAccounts struct{ *Memory }

func cloneAccount(a *BillingAccount) *BillingAccount {
	if a == nil {
		return nil
	}
	cp := *a
	if a.ProviderCustomerID != nil {
		s := *a.ProviderCustomerID
		cp.ProviderCustomerID = &s
	}
	return &cp
}

func ownerKey(ownerType, ownerID string) string { return ownerType + "\x00" + ownerID }

func (m *MemAccounts) GetByID(_ context.Context, id string) (*BillingAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneAccount(a), nil
}

func (m *MemAccounts) GetByOwner(_ context.Context, ownerType, ownerID string) (*BillingAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accountsOwner[ownerKey(ownerType, ownerID)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneAccount(a), nil
}

func (m *MemAccounts) Insert(_ context.Context, a *BillingAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := ownerKey(a.OwnerType, a.OwnerID)
	if _, ok := m.accountsOwner[key]; ok {
		return ErrDuplicate
	}
	if _, ok := m.accounts[a.ID]; ok {
		return ErrDuplicate
	}
	stored := cloneAccount(a)
	m.accounts[a.ID] = stored
	m.accountsOwner[key] = stored
	return nil
}

func (m *MemAccounts) SetProviderCustomer(_ context.Context, id, customerID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[id]
	if !ok {
		return ErrNotFound
	}
	cid := customerID
	a.ProviderCustomerID = &cid
	a.UpdatedAt = at
	return nil
}
