package org

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

var ErrNotFound = errors.New("not found")

type OrgRepo interface {
	Insert(ctx context.Context, o *Organization) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	UpdateName(ctx context.Context, id, name string, at time.Time) error
	SetProviderCustomer(ctx context.Context, id, provider, customerID string, at time.Time) error
	SoftDelete(ctx context.Context, id string, at time.Time) error
	SetOwner(ctx context.Context, id, ownerUserID string, at time.Time) error
}

type MemberRepo interface {
	Insert(ctx context.Context, m *Membership) error
	GetByOrgUser(ctx context.Context, orgID, userID string) (*Membership, error)
	ListByUser(ctx context.Context, userID, cursor string, limit int) ([]MembershipItem, string, error)
	ListByOrg(ctx context.Context, orgID, status, role, cursor string, limit int) ([]Membership, string, error)
	Update(ctx context.Context, m *Membership) error
	CountActiveOwners(ctx context.Context, orgID string) (int, error)
}

type InvitationRepo interface {
	Insert(ctx context.Context, inv *Invitation) error
	GetByID(ctx context.Context, id string) (*Invitation, error)
	GetByTokenHash(ctx context.Context, hash string) (*Invitation, error)
	GetPendingByOrgEmail(ctx context.Context, orgID, email string) (*Invitation, error)
	ListByOrg(ctx context.Context, orgID, status, cursor string, limit int) ([]Invitation, string, error)
	Update(ctx context.Context, inv *Invitation) error
}

type Mailer interface {
	SendInvite(ctx context.Context, email, rawToken, orgName string) error
}

type Clock interface {
	Now() time.Time
}

type IDGen interface {
	Org() string
	Member() string
	Invitation() string
}

type Tokens interface {
	NewInvite() (raw, hash string)
	Hash(raw string) string
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// UserLookup is identity.UserRepo; named here so org does not force every
// caller to think in identity-repo terms.
type UserLookup = identity.UserRepo

// OrgProvisioner creates an organization-owned billing account + trial. Implemented by billing.
type OrgProvisioner interface {
	EnsureOrganization(ctx context.Context, orgID string) error
}
