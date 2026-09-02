package billing

import (
	"context"
)

func (m *MemSubs) GetByOrg(_ context.Context, orgID string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if orgID == "" {
		return nil, ErrNotFound
	}
	s, ok := m.subsOrg[orgID]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSub(s), nil
}

func (m *MemSubs) GetByAccountID(_ context.Context, accountID string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.subsAccount[accountID]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSub(s), nil
}

func (m *MemSubs) GetByID(_ context.Context, id string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.subs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSub(s), nil
}

func (m *MemSubs) Insert(_ context.Context, s *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.BillingAccountID != "" {
		if _, ok := m.subsAccount[s.BillingAccountID]; ok {
			return ErrDuplicate
		}
	}
	stored := cloneSub(s)
	m.subs[s.ID] = stored
	if s.BillingAccountID != "" {
		m.subsAccount[s.BillingAccountID] = stored
	}
	if s.OrganizationID != "" {
		m.subsOrg[s.OrganizationID] = stored
	}
	return nil
}

func (m *MemSubs) Update(_ context.Context, s *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.subs[s.ID]; !ok {
		return ErrNotFound
	}
	stored := cloneSub(s)
	m.subs[s.ID] = stored
	if s.BillingAccountID != "" {
		m.subsAccount[s.BillingAccountID] = stored
	}
	if s.OrganizationID != "" {
		m.subsOrg[s.OrganizationID] = stored
	}
	return nil
}
