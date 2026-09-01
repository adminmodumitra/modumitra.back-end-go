package device

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

var ErrNotFound = errors.New("not found")

type DeviceRepo interface {
	Insert(ctx context.Context, d *Device) error
	GetByID(ctx context.Context, id string) (*Device, error)
	ListByUser(ctx context.Context, userID string, includeDeactivated bool, cursor string, limit int) ([]Device, string, error)
	CountActive(ctx context.Context, userID string) (int, error)
	UpdateName(ctx context.Context, id, name string, at time.Time) error
	Deactivate(ctx context.Context, id, reason string, at time.Time) error
}

type EntitlementRepo interface {
	Insert(ctx context.Context, e *Entitlement) error
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type IDGen interface {
	Device() string
	Entitlement() string
}

type Signer interface {
	KeyID() string
	Sign(canonical []byte) string
	Verify(canonical []byte, signature string) bool
}

// OrgLookup is the org-side surface devices need (primary org + membership).
type OrgLookup interface {
	PrimaryForUser(ctx context.Context, userID string) (*org.Organization, error)
	MembershipOf(ctx context.Context, orgID, userID string) (*org.Organization, *org.Membership, error)
}
