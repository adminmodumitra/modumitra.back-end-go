package billing

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")

type PlanRepo interface {
	ListSellable(ctx context.Context) ([]Plan, error)
	GetByID(ctx context.Context, id string) (*Plan, error)
}

type AccountRepo interface {
	GetByID(ctx context.Context, id string) (*BillingAccount, error)
	GetByOwner(ctx context.Context, ownerType, ownerID string) (*BillingAccount, error)
	Insert(ctx context.Context, a *BillingAccount) error
	SetProviderCustomer(ctx context.Context, id, customerID string, at time.Time) error
}

type SubscriptionRepo interface {
	GetByOrg(ctx context.Context, orgID string) (*Subscription, error)
	GetByAccountID(ctx context.Context, accountID string) (*Subscription, error)
	GetByID(ctx context.Context, id string) (*Subscription, error)
	Insert(ctx context.Context, s *Subscription) error
	Update(ctx context.Context, s *Subscription) error
}

type InvoiceRepo interface {
	Insert(ctx context.Context, inv *Invoice) error
	GetByID(ctx context.Context, id string) (*Invoice, error)
	GetByProviderInvoiceID(ctx context.Context, providerInvoiceID string) (*Invoice, error)
	ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]Invoice, string, error)
	ListByAccount(ctx context.Context, accountID, status, cursor string, limit int) ([]Invoice, string, error)
}

type EventRepo interface {
	Insert(ctx context.Context, e *Event) error
	GetByProviderEvent(ctx context.Context, provider, providerEventID string) (*Event, error)
	MarkProcessed(ctx context.Context, id string, orgID, subID, accountID *string, at time.Time) error
	MarkFailed(ctx context.Context, id, errMsg string, at time.Time) error
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type IDGen interface {
	Account() string
	Subscription() string
	Invoice() string
	Event() string
}

type Organizations interface {
	GetByID(ctx context.Context, id string) (*org.Organization, error)
	SetProviderCustomer(ctx context.Context, id, provider, customerID string, at time.Time) error
}

type Members interface {
	GetByOrgUser(ctx context.Context, orgID, userID string) (*org.Membership, error)
	ListByOrg(ctx context.Context, orgID, status, role, cursor string, limit int) ([]org.Membership, string, error)
}

type Invites interface {
	ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]org.Invitation, string, error)
}

type Users interface {
	GetByID(ctx context.Context, id string) (*identity.User, error)
}
