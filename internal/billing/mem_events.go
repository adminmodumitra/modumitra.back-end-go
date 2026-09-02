package billing

import (
	"context"
	"time"
)

func (m *MemEvents) Insert(_ context.Context, e *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := eventKey(e.Provider, e.ProviderEventID)
	if _, ok := m.evtProv[key]; ok {
		return ErrDuplicate
	}
	stored := cloneEvent(e)
	m.events[e.ID] = stored
	m.evtProv[key] = stored
	return nil
}

func (m *MemEvents) GetByProviderEvent(_ context.Context, provider, providerEventID string) (*Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.evtProv[eventKey(provider, providerEventID)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneEvent(e), nil
}

func (m *MemEvents) MarkProcessed(_ context.Context, id string, orgID, subID, accountID *string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.events[id]
	if !ok {
		return ErrNotFound
	}
	e.Status = EventProcessed
	t := at
	e.ProcessedAt = &t
	if orgID != nil {
		s := *orgID
		e.OrganizationID = &s
	}
	if subID != nil {
		s := *subID
		e.SubscriptionID = &s
	}
	if accountID != nil {
		s := *accountID
		e.BillingAccountID = &s
	}
	return nil
}

func (m *MemEvents) MarkFailed(_ context.Context, id, errMsg string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.events[id]
	if !ok {
		return ErrNotFound
	}
	e.Status = EventFailed
	e.Error = errMsg
	t := at
	e.ProcessedAt = &t
	return nil
}
