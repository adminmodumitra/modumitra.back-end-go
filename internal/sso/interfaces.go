package sso

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidState = errors.New("invalid state")
	ErrDuplicate    = errors.New("duplicate")
)

// Provider is an authorization-code adapter. Implementations must not persist tokens.
type Provider interface {
	ID() string
	Label() string
	DeviceCodeSupported() bool
	AuthCodeURL(state, redirectURI, codeChallenge string) string
	Exchange(ctx context.Context, code, redirectURI, codeVerifier string) (Profile, error)
}

type IdentityRepo interface {
	Insert(ctx context.Context, id *Identity) error
	GetByProviderSubject(ctx context.Context, provider, subject string) (*Identity, error)
	GetByUserProvider(ctx context.Context, userID, provider string) (*Identity, error)
	ListByUser(ctx context.Context, userID string) ([]Identity, error)
	DeleteByUserProvider(ctx context.Context, userID, provider string) error
}

type StateRepo interface {
	Insert(ctx context.Context, st *State) error
	Consume(ctx context.Context, state, provider string, now time.Time) (*State, error)
}

type DeviceCodeRepo interface {
	Insert(ctx context.Context, d *DeviceCode) error
	GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCode, error)
	GetByUserCode(ctx context.Context, userCode string) (*DeviceCode, error)
	Update(ctx context.Context, d *DeviceCode) error
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Accounts is the identity surface SSO needs. *identity.Service implements it.
type Accounts interface {
	UserByID(ctx context.Context, id string) (*identity.User, error)
	UserByEmail(ctx context.Context, email string) (*identity.User, error)
	CreateUser(ctx context.Context, u *identity.User) error
	LoginAs(ctx context.Context, userID, deviceName string) (*identity.User, *identity.TokenPair, error)
	NewUserID() string
	TermsVersion() string
	PrivacyVersion() string
}

var _ Accounts = (*identity.Service)(nil)
