package billing

import (
	"context"
	"errors"

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
