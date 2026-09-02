package billing

import (
	"errors"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Service struct {
	plans    PlanRepo
	accounts AccountRepo
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
	accounts AccountRepo,
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
		plans: plans, accounts: accounts, subs: subs, invoices: invoices, events: events,
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
	svc := NewService(mem.Plans(), mem.Accounts(), mem.Subs(), mem.Invoices(), mem.Events(), fp, orgs, members, invites, users, mem, SystemClock{}, NewIDGen(), cfg, outbox.NewMemoryWriter())
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
