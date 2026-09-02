package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	plans    PlanRepo
	subs     SubscriptionRepo
	invoices InvoiceRepo
	events   EventRepo
	provider Provider
	orgs     Organizations
	members  Members
	invites  Invites
	users    Users
	tx       TxRunner
	clock    Clock
	ids      IDGen
	cfg      Settings
	outbox   outbox.Writer
}

func NewService(
	plans PlanRepo,
	subs SubscriptionRepo,
	invoices InvoiceRepo,
	events EventRepo,
	provider Provider,
	orgs Organizations,
	members Members,
	invites Invites,
	users Users,
	tx TxRunner,
	clock Clock,
	ids IDGen,
	cfg Settings,
	ob outbox.Writer,
) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	if ids == nil {
		ids = NewIDGen()
	}
	if provider == nil {
		provider = NewFakeProvider()
	}
	if cfg.Provider == "" {
		cfg.Provider = provider.Name()
	}
	return &Service{
		plans: plans, subs: subs, invoices: invoices, events: events,
		provider: provider, orgs: orgs, members: members, invites: invites,
		users: users, tx: tx, clock: clock, ids: ids, cfg: cfg, outbox: ob,
	}
}

func NewMemoryService(orgs Organizations, members Members, invites Invites, users Users, cfg Settings) (*Service, *Memory, *FakeProvider) {
	mem := NewMemory()
	fp := NewFakeProvider()
	if cfg.Provider == "" {
		cfg.Provider = ProviderFake
	}
	svc := NewService(mem.Plans(), mem.Subs(), mem.Invoices(), mem.Events(), fp, orgs, members, invites, users, mem, SystemClock{}, NewIDGen(), cfg, outbox.NewMemoryWriter())
	return svc, mem, fp
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, org.ErrNotFound) || errors.Is(err, identity.ErrNotFound) {
		return utils.E(utils.CodeNotFound)
	}
	return utils.From(err)
}

func (s *Service) loadVisible(ctx context.Context, actorID, orgID string) (*org.Organization, *org.Membership, error) {
	if actorID == "" {
		return nil, nil, utils.E(utils.CodeUnauthenticated)
	}
	o, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if o.Deleted() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	m, err := s.members.GetByOrgUser(ctx, orgID, actorID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if !m.Active() {
		return nil, nil, utils.E(utils.CodeNotFound)
	}
	return o, m, nil
}

func requireRead(m *org.Membership) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	switch m.Role {
	case org.RoleOwner, org.RoleAdmin, org.RoleBillingAdmin:
		return nil
	default:
		return utils.E(utils.CodeForbidden)
	}
}

func requireWrite(m *org.Membership) error {
	if m == nil {
		return utils.E(utils.CodeNotFound)
	}
	switch m.Role {
	case org.RoleOwner, org.RoleBillingAdmin:
		return nil
	default:
		return utils.E(utils.CodeForbidden)
	}
}

func allowedRedirect(u string) bool {
	u = strings.TrimSpace(u)
	return strings.HasPrefix(u, "cabinetstudio://") || strings.HasPrefix(u, "https://")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	u := t.UTC()
	return &u
}

func (s *Service) ListPlans(ctx context.Context) ([]Plan, error) {
	items, err := s.plans.ListSellable(ctx)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if items == nil {
		items = []Plan{}
	}
	return items, nil
}

func (s *Service) GetSubscription(ctx context.Context, actorID, orgID string) (*Subscription, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(m); err != nil {
		return nil, err
	}
	return s.getOrCreateTrial(ctx, o)
}

func (s *Service) getOrCreateTrial(ctx context.Context, o *org.Organization) (*Subscription, error) {
	sub, err := s.subs.GetByOrg(ctx, o.ID)
	if err == nil {
		return sub, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if o.TrialEndsAt == nil || !o.TrialEndsAt.After(now) {
		return nil, utils.E(utils.CodeNotFound)
	}
	plan, err := s.plans.GetByID(ctx, PlanIDTrial)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	sub = &Subscription{
		ID:                 s.ids.Subscription(),
		OrganizationID:     o.ID,
		PlanID:             plan.ID,
		PlanCode:           plan.Code,
		Status:             StatusTrialing,
		Provider:           s.provider.Name(),
		SeatsIncluded:      plan.SeatsIncluded,
		TrialEndsAt:        o.TrialEndsAt,
		CurrentPeriodStart: o.CreatedAt,
		CurrentPeriodEnd:   *o.TrialEndsAt,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.subs.Insert(ctx, sub); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return s.subs.GetByOrg(ctx, o.ID)
		}
		return nil, mapRepoErr(err)
	}
	return sub, nil
}

func (s *Service) Checkout(ctx context.Context, actorID, orgID, planID, successURL, cancelURL string) (*CheckoutSession, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	u, err := s.users.GetByID(ctx, actorID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !u.EmailVerified() {
		return nil, utils.E(utils.CodeEmailUnverified)
	}
	if !allowedRedirect(successURL) || !allowedRedirect(cancelURL) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, utils.E(utils.CodePlanNotFound)
	}
	if !plan.Sellable || !plan.Active {
		return nil, utils.E(utils.CodePlanNotFound)
	}
	existing, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, mapRepoErr(err)
	}
	if existing != nil && existing.PlanCode != PlanTrial && (existing.Status == StatusActive || existing.Status == StatusTrialing) {
		return nil, utils.E(utils.CodeSubscriptionActive)
	}
	sess, err := s.provider.CreateCheckout(ctx, CheckoutInput{
		OrganizationID: orgID,
		PlanID:         plan.ID,
		SuccessURL:     successURL,
		CancelURL:      cancelURL,
		CustomerID:     o.ProviderCustomerID,
		Email:          u.Email,
		Currency:       plan.Currency,
		AmountMinor:    plan.UnitAmountMinor,
		Seats:          plan.SeatsIncluded,
		PlanCode:       plan.Code,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}

func (s *Service) Portal(ctx context.Context, actorID, orgID, returnURL string) (*PortalSession, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	if o.ProviderCustomerID == "" {
		return nil, utils.E(utils.CodeBillingCustomerMissing)
	}
	if !allowedRedirect(returnURL) {
		return nil, utils.E(utils.CodeRedirectURINotAllowed)
	}
	sess, err := s.provider.CreatePortal(ctx, PortalInput{
		OrganizationID: orgID,
		CustomerID:     o.ProviderCustomerID,
		ReturnURL:      returnURL,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &sess, nil
}

func (s *Service) Cancel(ctx context.Context, actorID, orgID string, atPeriodEnd bool) (*Subscription, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	if !atPeriodEnd {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "at_period_end", "reason": "must be true"})
	}
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if sub.Status == StatusCanceled || sub.CancelAtPeriodEnd {
		return nil, utils.E(utils.CodeAlreadyCanceled)
	}
	if sub.Status != StatusActive && sub.Status != StatusTrialing {
		return nil, utils.E(utils.CodeSubscriptionNotCancelable)
	}
	if sub.ProviderSubscriptionID != "" {
		if err := s.provider.Cancel(ctx, sub.ProviderSubscriptionID, true); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	now := s.clock.Now()
	sub.CancelAtPeriodEnd = true
	sub.CanceledAt = timePtr(now)
	sub.UpdatedAt = now
	if err := s.subs.Update(ctx, sub); err != nil {
		return nil, mapRepoErr(err)
	}
	return sub, nil
}

func (s *Service) Resume(ctx context.Context, actorID, orgID string) (*Subscription, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	now := s.clock.Now()
	if sub.Status == StatusCanceled || !sub.CurrentPeriodEnd.After(now) {
		return nil, utils.E(utils.CodeSubscriptionNotResumable)
	}
	if !sub.CancelAtPeriodEnd {
		return nil, utils.E(utils.CodeNotCanceling)
	}
	if sub.ProviderSubscriptionID != "" {
		if err := s.provider.Resume(ctx, sub.ProviderSubscriptionID); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = now
	if err := s.subs.Update(ctx, sub); err != nil {
		return nil, mapRepoErr(err)
	}
	return sub, nil
}

func (s *Service) ListInvoices(ctx context.Context, actorID, orgID, status, cursor string, limit int) ([]Invoice, string, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, "", err
	}
	if err := requireRead(m); err != nil {
		return nil, "", err
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	items, next, err := s.invoices.ListByOrg(ctx, orgID, status, cursor, limit)
	if err != nil {
		return nil, "", mapRepoErr(err)
	}
	if items == nil {
		items = []Invoice{}
	}
	return items, next, nil
}

func (s *Service) GetInvoice(ctx context.Context, actorID, orgID, invoiceID string) (*Invoice, error) {
	_, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(m); err != nil {
		return nil, err
	}
	inv, err := s.invoices.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if inv.OrganizationID != orgID {
		return nil, utils.E(utils.CodeNotFound)
	}
	return inv, nil
}

func (s *Service) Seats(ctx context.Context, actorID, orgID string) (*Seats, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireRead(m); err != nil {
		return nil, err
	}
	sub, err := s.getOrCreateTrial(ctx, o)
	if err != nil {
		return nil, err
	}
	return s.seatSnapshot(ctx, orgID, sub.SeatsIncluded)
}

func (s *Service) seatSnapshot(ctx context.Context, orgID string, purchased int) (*Seats, error) {
	members, _, err := s.members.ListByOrg(ctx, orgID, "", "", "", 500)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	pending := 0
	if s.invites != nil {
		invs, _, err := s.invites.ListByOrg(ctx, orgID, org.InvitePending, "", 500)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		pending = len(invs)
	}
	out := &Seats{Purchased: purchased, PendingInvitations: pending, Members: []SeatMember{}}
	for i := range members {
		mem := members[i]
		if mem.Status == org.MemberRemoved || mem.Status == org.MemberInvited {
			continue
		}
		consumes := mem.Status == org.MemberActive
		if consumes {
			out.Used++
		}
		out.Members = append(out.Members, SeatMember{
			UserID: mem.UserID, Role: mem.Role, Status: mem.Status, ConsumesSeat: consumes,
		})
	}
	out.Available = purchased - out.Used
	return out, nil
}

func (s *Service) PreviewDowngrade(ctx context.Context, actorID, orgID, planID string, seats *int) (*Preview, error) {
	o, m, err := s.loadVisible(ctx, actorID, orgID)
	if err != nil {
		return nil, err
	}
	if err := requireWrite(m); err != nil {
		return nil, err
	}
	sub, err := s.getOrCreateTrial(ctx, o)
	if err != nil {
		return nil, err
	}
	targetSeats := 0
	planCode := ""
	var targetCaps Capabilities
	if strings.TrimSpace(planID) != "" {
		plan, err := s.plans.GetByID(ctx, planID)
		if err != nil {
			return nil, utils.E(utils.CodePlanNotFound)
		}
		targetSeats = plan.SeatsIncluded
		planCode = plan.Code
		targetCaps = plan.Capabilities
	} else if seats != nil {
		if *seats < 0 {
			return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "seats"})
		}
		targetSeats = *seats
		planCode = sub.PlanCode
		planID = sub.PlanID
	} else {
		return nil, utils.E(utils.CodeValidation).WithDetails(map[string]any{"field": "plan_id"})
	}
	snap, err := s.seatSnapshot(ctx, orgID, sub.SeatsIncluded)
	if err != nil {
		return nil, err
	}
	shortfall := snap.Used - targetSeats
	if shortfall < 0 {
		shortfall = 0
	}
	over := []SeatMember{}
	active := []SeatMember{}
	for _, mem := range snap.Members {
		if mem.ConsumesSeat {
			active = append(active, mem)
		}
	}
	if targetSeats < len(active) {
		over = append([]SeatMember{}, active[targetSeats:]...)
	}
	loss := []string{}
	cur, _ := s.plans.GetByID(ctx, sub.PlanID)
	if cur != nil {
		if cur.Capabilities.EngineeringHandoff && !targetCaps.EngineeringHandoff {
			loss = append(loss, "engineering_handoff")
		}
		if cur.Capabilities.ProposalBranding && !targetCaps.ProposalBranding {
			loss = append(loss, "proposal_branding")
		}
		if cur.Capabilities.OrganizationAssets && !targetCaps.OrganizationAssets {
			loss = append(loss, "organization_assets")
		}
		if cur.Capabilities.CloudProjects && !targetCaps.CloudProjects {
			loss = append(loss, "cloud_projects")
		}
	}
	if loss == nil {
		loss = []string{}
	}
	blocked := shortfall > 0
	msg := fmt.Sprintf("Switching to %d seats is allowed.", targetSeats)
	if blocked {
		msg = fmt.Sprintf("Reduce active members to %d before switching.", targetSeats)
	}
	preview := &Preview{
		PlanID: planID, Plan: planCode, PurchasedSeats: sub.SeatsIncluded,
		TargetSeats: targetSeats, SeatShortfall: shortfall,
		MembersOverTarget: over, CapabilityLoss: loss, Blocked: blocked, Message: msg,
	}
	if blocked {
		return nil, utils.E(utils.CodeSeatLimit).WithDetails(map[string]any{
			"purchased_seats":     preview.PurchasedSeats,
			"target_seats":        preview.TargetSeats,
			"used_seats":          snap.Used,
			"seat_shortfall":      preview.SeatShortfall,
			"members_over_target": preview.MembersOverTarget,
			"plan_id":             preview.PlanID,
			"plan":                preview.Plan,
			"capability_loss":     preview.CapabilityLoss,
			"message":             preview.Message,
		})
	}
	return preview, nil
}

func (s *Service) CompleteCheckout(ctx context.Context, sessionID string) error {
	fp, ok := s.provider.(*FakeProvider)
	if !ok {
		return utils.E(utils.CodeServiceUnavailable)
	}
	raw, err := fp.CompletedPayload(sessionID)
	if err != nil {
		return mapRepoErr(err)
	}
	return s.HandleWebhook(ctx, raw)
}

func (s *Service) HandleWebhook(ctx context.Context, raw []byte) error {
	var env struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object map[string]any `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || strings.TrimSpace(env.ID) == "" || strings.TrimSpace(env.Type) == "" {
		return utils.E(utils.CodeInvalidPayload)
	}
	provider := s.provider.Name()
	if existing, err := s.events.GetByProviderEvent(ctx, provider, env.ID); err == nil && existing != nil {
		return nil
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return mapRepoErr(err)
	}
	now := s.clock.Now()
	summary := redactSummary(env.Type, env.Data.Object)
	evt := &Event{
		ID: s.ids.Event(), Provider: provider, ProviderEventID: env.ID,
		Type: env.Type, Status: EventReceived, PayloadDigest: PayloadDigest(raw),
		Summary: summary, CreatedAt: now,
	}
	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.events.Insert(ctx, evt); err != nil {
			return err
		}
		orgID, subID, applyErr := s.applyEvent(ctx, env.Type, env.Data.Object, now)
		if applyErr != nil {
			_ = s.events.MarkFailed(ctx, evt.ID, applyErr.Error(), now)
			return applyErr
		}
		if env.Type == EventCheckoutCompleted && orgID != "" {
			planID := asString(env.Data.Object, "plan_id")
			status := StatusActive
			if subID != "" {
				if sub, gerr := s.subs.GetByID(ctx, subID); gerr == nil && sub != nil {
					planID = sub.PlanID
					status = sub.Status
				}
			}
			if err := outbox.Enqueue(ctx, s.outbox, outbox.BillingSubscriptionUpdated(ctx, orgID, subID, planID, status, now)); err != nil {
				return err
			}
		}
		return s.events.MarkProcessed(ctx, evt.ID, strPtr(orgID), strPtr(subID), now)
	})
	if errors.Is(err, ErrDuplicate) {
		return nil
	}
	return mapRepoErr(err)
}

func redactSummary(typ string, obj map[string]any) map[string]any {
	out := map[string]any{"type": typ}
	for _, k := range []string{"id", "organization_id", "plan_id", "customer", "subscription", "amount_total", "currency", "status"} {
		if v, ok := obj[k]; ok {
			out[k] = v
		}
	}
	return out
}

func asString(obj map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func asInt(obj map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case json.Number:
				i, _ := n.Int64()
				return int(i)
			}
		}
	}
	return 0
}

func (s *Service) applyEvent(ctx context.Context, typ string, obj map[string]any, now time.Time) (orgID, subID string, err error) {
	switch typ {
	case EventCheckoutCompleted:
		return s.applyCheckoutCompleted(ctx, obj, now)
	case EventInvoicePaid:
		return s.applyInvoicePaid(ctx, obj, now)
	default:
		return "", "", nil
	}
}

func (s *Service) applyCheckoutCompleted(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	orgID := asString(obj, "organization_id", "client_reference_id")
	planID := asString(obj, "plan_id")
	customer := asString(obj, "customer")
	provSub := asString(obj, "subscription")
	if orgID == "" || planID == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return "", "", mapRepoErr(err)
	}
	if err := s.orgs.SetProviderCustomer(ctx, orgID, s.provider.Name(), customer, now); err != nil {
		return "", "", mapRepoErr(err)
	}
	periodEnd := now.AddDate(0, 1, 0)
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	if sub == nil {
		sub = &Subscription{
			ID: s.ids.Subscription(), OrganizationID: orgID, CreatedAt: now,
		}
	}
	sub.PlanID = plan.ID
	sub.PlanCode = plan.Code
	sub.Status = StatusActive
	sub.Provider = s.provider.Name()
	sub.ProviderSubscriptionID = provSub
	sub.SeatsIncluded = plan.SeatsIncluded
	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = periodEnd
	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = now
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	if _, err := s.subs.GetByID(ctx, sub.ID); err != nil {
		if err := s.subs.Insert(ctx, sub); err != nil && !errors.Is(err, ErrDuplicate) {
			return "", "", mapRepoErr(err)
		}
	} else {
		if err := s.subs.Update(ctx, sub); err != nil {
			return "", "", mapRepoErr(err)
		}
	}
	provInv := asString(obj, "provider_invoice_id")
	if provInv == "" {
		provInv = PrefixProvInv + asString(obj, "id")
	}
	if _, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); errors.Is(err, ErrNotFound) {
		hosted := asString(obj, "hosted_invoice_url")
		if hosted == "" {
			hosted = "https://billing.fake.local/invoice/" + provInv
		}
		amount := asInt(obj, "amount_total")
		if amount == 0 {
			amount = plan.UnitAmountMinor
		}
		currency := asString(obj, "currency")
		if currency == "" {
			currency = plan.Currency
		}
		num := "INV-" + now.UTC().Format("20060102") + "-" + sub.ID[len(sub.ID)-4:]
		inv := &Invoice{
			ID: s.ids.Invoice(), OrganizationID: orgID, SubscriptionID: strPtr(sub.ID),
			Provider: s.provider.Name(), ProviderInvoiceID: provInv, Number: strPtr(num),
			Status: InvoicePaid, Currency: currency, TotalMinor: amount, AmountDueMinor: 0,
			HostedInvoiceURL: strPtr(hosted), PeriodStart: timePtr(now), PeriodEnd: timePtr(periodEnd),
			PaidAt: timePtr(now), CreatedAt: now, UpdatedAt: now,
		}
		if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
			return "", "", mapRepoErr(err)
		}
	} else if err != nil {
		return "", "", mapRepoErr(err)
	}
	return orgID, sub.ID, nil
}

func (s *Service) applyInvoicePaid(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	orgID := asString(obj, "organization_id")
	provInv := asString(obj, "id", "provider_invoice_id")
	if provInv == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	if existing, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); err == nil {
		return existing.OrganizationID, deref(existing.SubscriptionID), nil
	} else if !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	subID := asString(obj, "subscription_id", "subscription")
	var localSubID *string
	if orgID == "" && subID != "" {
		if sub, err := s.subs.GetByOrg(ctx, orgID); err == nil {
			orgID = sub.OrganizationID
			localSubID = strPtr(sub.ID)
		}
	}
	if orgID == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	if sub, err := s.subs.GetByOrg(ctx, orgID); err == nil {
		localSubID = strPtr(sub.ID)
	}
	amount := asInt(obj, "amount_paid", "amount_total", "total_minor")
	currency := asString(obj, "currency")
	if currency == "" {
		currency = "usd"
	}
	hosted := asString(obj, "hosted_invoice_url")
	pdf := asString(obj, "invoice_pdf", "invoice_pdf_url")
	inv := &Invoice{
		ID: s.ids.Invoice(), OrganizationID: orgID, SubscriptionID: localSubID,
		Provider: s.provider.Name(), ProviderInvoiceID: provInv,
		Status: InvoicePaid, Currency: currency, TotalMinor: amount, AmountDueMinor: 0,
		HostedInvoiceURL: strPtr(hosted), InvoicePDFURL: strPtr(pdf),
		PaidAt: timePtr(now), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
		return "", "", mapRepoErr(err)
	}
	return orgID, deref(localSubID), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
