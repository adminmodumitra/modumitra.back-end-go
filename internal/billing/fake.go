package billing

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type storedCheckout struct {
	Session CheckoutSession
	Input   CheckoutInput
	SubID   string
}

// FakeProvider is an in-memory billing adapter. Completing checkout is done
// by applying a webhook event (CompletedPayload / Service.CompleteCheckout).
type FakeProvider struct {
	mu        sync.Mutex
	checkouts map[string]*storedCheckout
	portals   map[string]PortalSession
	canceled  map[string]bool
}

func NewFakeProvider() *FakeProvider {
	return &FakeProvider{
		checkouts: map[string]*storedCheckout{},
		portals:   map[string]PortalSession{},
		canceled:  map[string]bool{},
	}
}

func (p *FakeProvider) Name() string { return ProviderFake }

func (p *FakeProvider) CreateCheckout(_ context.Context, in CheckoutInput) (CheckoutSession, error) {
	now := time.Now().UTC()
	id := PrefixCheckout + randHex(16)
	customer := in.CustomerID
	if customer == "" {
		customer = PrefixCustomer + randHex(12)
	}
	sess := CheckoutSession{
		ID:         id,
		URL:        "https://billing.fake.local/checkout/" + id,
		ExpiresAt:  now.Add(SessionTTL),
		CustomerID: customer,
	}
	p.mu.Lock()
	p.checkouts[id] = &storedCheckout{Session: sess, Input: in}
	p.mu.Unlock()
	return sess, nil
}

func (p *FakeProvider) CreatePortal(_ context.Context, in PortalInput) (PortalSession, error) {
	now := time.Now().UTC()
	id := PrefixPortal + randHex(16)
	sess := PortalSession{
		ID:        id,
		URL:       "https://billing.fake.local/portal/" + id,
		ExpiresAt: now.Add(SessionTTL),
	}
	p.mu.Lock()
	p.portals[id] = sess
	p.mu.Unlock()
	_ = in
	return sess, nil
}

func (p *FakeProvider) Cancel(_ context.Context, providerSubID string, atPeriodEnd bool) error {
	if providerSubID == "" {
		return ErrNotFound
	}
	p.mu.Lock()
	p.canceled[providerSubID] = atPeriodEnd
	p.mu.Unlock()
	return nil
}

func (p *FakeProvider) Resume(_ context.Context, providerSubID string) error {
	if providerSubID == "" {
		return ErrNotFound
	}
	p.mu.Lock()
	delete(p.canceled, providerSubID)
	p.mu.Unlock()
	return nil
}

func (p *FakeProvider) LookupCheckout(id string) (CheckoutSession, CheckoutInput, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	rec, ok := p.checkouts[id]
	if !ok {
		return CheckoutSession{}, CheckoutInput{}, false
	}
	return rec.Session, rec.Input, true
}

// CompletedPayload builds a checkout.session.completed webhook body for tests
// and the webhook handler's CompleteCheckout helper.
func (p *FakeProvider) CompletedPayload(sessionID string) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	rec, ok := p.checkouts[sessionID]
	if !ok {
		return nil, ErrNotFound
	}
	if rec.SubID == "" {
		rec.SubID = PrefixProvSub + randHex(12)
	}
	evtID := PrefixProvEvt + randHex(12)
	body := map[string]any{
		"id":   evtID,
		"type": EventCheckoutCompleted,
		"data": map[string]any{
			"object": map[string]any{
				"id":                  rec.Session.ID,
				"organization_id":     rec.Input.OrganizationID,
				"plan_id":             rec.Input.PlanID,
				"customer":            rec.Session.CustomerID,
				"subscription":        rec.SubID,
				"amount_total":        rec.Input.AmountMinor,
				"currency":            rec.Input.Currency,
				"seats":               rec.Input.Seats,
				"hosted_invoice_url":  "https://billing.fake.local/invoice/" + PrefixInvoice + rec.Session.ID,
				"provider_invoice_id": PrefixProvInv + rec.Session.ID,
			},
		},
	}
	return json.Marshal(body)
}

var _ Provider = (*FakeProvider)(nil)
