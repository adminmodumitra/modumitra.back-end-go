package billing

import (
	"context"
	"sync"
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
