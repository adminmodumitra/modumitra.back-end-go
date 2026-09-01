package billing

import (
	"context"
	"sort"
	"sync"
	"time"
)

type Memory struct {
	mu       sync.Mutex
	plans    map[string]*Plan
	subs     map[string]*Subscription
	subsOrg  map[string]*Subscription
	invoices map[string]*Invoice
	invProv  map[string]*Invoice
	events   map[string]*Event
	evtProv  map[string]*Event
}

func NewMemory() *Memory {
	m := &Memory{
		plans:    map[string]*Plan{},
		subs:     map[string]*Subscription{},
		subsOrg:  map[string]*Subscription{},
		invoices: map[string]*Invoice{},
		invProv:  map[string]*Invoice{},
		events:   map[string]*Event{},
		evtProv:  map[string]*Event{},
	}
	for i := range SeedPlans() {
		p := SeedPlans()[i]
		m.plans[p.ID] = clonePlan(&p)
	}
	return m
}

func (m *Memory) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (m *Memory) Plans() *MemPlans       { return &MemPlans{Memory: m} }
func (m *Memory) Subs() *MemSubs         { return &MemSubs{Memory: m} }
func (m *Memory) Invoices() *MemInvoices { return &MemInvoices{Memory: m} }
func (m *Memory) Events() *MemEvents     { return &MemEvents{Memory: m} }

type MemPlans struct{ *Memory }
type MemSubs struct{ *Memory }
type MemInvoices struct{ *Memory }
type MemEvents struct{ *Memory }

func (m *MemPlans) ListSellable(_ context.Context) ([]Plan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Plan
	for _, p := range m.plans {
		if p.Sellable && p.Active {
			out = append(out, *clonePlan(p))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	if out == nil {
		out = []Plan{}
	}
	return out, nil
}

func (m *MemPlans) GetByID(_ context.Context, id string) (*Plan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return clonePlan(p), nil
}

func cloneSub(s *Subscription) *Subscription {
	if s == nil {
		return nil
	}
	cp := *s
	if s.TrialEndsAt != nil {
		t := *s.TrialEndsAt
		cp.TrialEndsAt = &t
	}
	if s.CanceledAt != nil {
		t := *s.CanceledAt
		cp.CanceledAt = &t
	}
	if s.GraceUntil != nil {
		t := *s.GraceUntil
		cp.GraceUntil = &t
	}
	return &cp
}

func (m *MemSubs) GetByOrg(_ context.Context, orgID string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.subsOrg[orgID]
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
	if _, ok := m.subsOrg[s.OrganizationID]; ok {
		return ErrDuplicate
	}
	stored := cloneSub(s)
	m.subs[s.ID] = stored
	m.subsOrg[s.OrganizationID] = stored
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
	m.subsOrg[s.OrganizationID] = stored
	return nil
}

func cloneInvoice(inv *Invoice) *Invoice {
	if inv == nil {
		return nil
	}
	cp := *inv
	if inv.SubscriptionID != nil {
		s := *inv.SubscriptionID
		cp.SubscriptionID = &s
	}
	if inv.Number != nil {
		s := *inv.Number
		cp.Number = &s
	}
	if inv.HostedInvoiceURL != nil {
		s := *inv.HostedInvoiceURL
		cp.HostedInvoiceURL = &s
	}
	if inv.InvoicePDFURL != nil {
		s := *inv.InvoicePDFURL
		cp.InvoicePDFURL = &s
	}
	if inv.PeriodStart != nil {
		t := *inv.PeriodStart
		cp.PeriodStart = &t
	}
	if inv.PeriodEnd != nil {
		t := *inv.PeriodEnd
		cp.PeriodEnd = &t
	}
	if inv.PaidAt != nil {
		t := *inv.PaidAt
		cp.PaidAt = &t
	}
	return &cp
}

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

func (m *MemInvoices) ListByOrg(_ context.Context, orgID, status, cursor string, limit int) ([]Invoice, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Invoice
	for _, inv := range m.invoices {
		if inv.OrganizationID != orgID {
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

func cloneEvent(e *Event) *Event {
	if e == nil {
		return nil
	}
	cp := *e
	if e.OrganizationID != nil {
		s := *e.OrganizationID
		cp.OrganizationID = &s
	}
	if e.SubscriptionID != nil {
		s := *e.SubscriptionID
		cp.SubscriptionID = &s
	}
	if e.ProcessedAt != nil {
		t := *e.ProcessedAt
		cp.ProcessedAt = &t
	}
	if e.Summary != nil {
		cp.Summary = map[string]any{}
		for k, v := range e.Summary {
			cp.Summary[k] = v
		}
	}
	return &cp
}

func eventKey(provider, id string) string { return provider + "\x00" + id }

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

func (m *MemEvents) MarkProcessed(_ context.Context, id string, orgID, subID *string, at time.Time) error {
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
