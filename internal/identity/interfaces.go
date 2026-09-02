package identity

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by repos when a row is missing. Service maps it
// to a catalog error; it must never become a 500.
var ErrNotFound = errors.New("not found")

type UserRepo interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmailNormalized(ctx context.Context, email string) (*User, error)
	Insert(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id, hash string, at time.Time) error
	SetEmailVerified(ctx context.Context, id string, at time.Time) error
	SetLastLogin(ctx context.Context, id string, at time.Time) error
}

type SessionRepo interface {
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByRefreshHash(ctx context.Context, hash string) (*Session, error)
	GetByRefreshHashForUpdate(ctx context.Context, hash string) (*Session, error)
	Insert(ctx context.Context, s *Session) error
	RevokeAndReplace(ctx context.Context, oldID, newID string, at time.Time) error
	MarkReuseAndRevokeFamily(ctx context.Context, sessionID, familyID string, at time.Time) error
	Revoke(ctx context.Context, id string, at time.Time) error
	RevokeAllForUser(ctx context.Context, userID string, at time.Time) (int, error)
}

type EmailTokenRepo interface {
	Insert(ctx context.Context, t *EmailToken) error
	GetByHash(ctx context.Context, hash string) (*EmailToken, error)
	Consume(ctx context.Context, id string, at time.Time) error
	InvalidateOpen(ctx context.Context, userID, purpose string, at time.Time) error
}

type Hasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, hash, password string) (bool, error)
	DummyVerify(ctx context.Context)
}

type Tokens interface {
	IssueAccess(ctx context.Context, userID, sessionID string) (string, error)
	ParseAccess(ctx context.Context, raw string) (userID, sessionID string, err error)
	NewRefresh() (raw, hash string)
	NewEmailToken(prefix string) (raw, hash string)
	Hash(raw string) string
	HashOTP(userID, otp string) string
}

type Mailer interface {
	SendVerification(ctx context.Context, email, rawToken string) error
	SendPasswordReset(ctx context.Context, email, rawToken string) error
	SendOTP(ctx context.Context, email, otp string) error
}

type Clock interface {
	Now() time.Time
}

type IDGen interface {
	User() string
	Session() string
	Token() string
	Family() string
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PersonalProvisioner creates a user-owned billing account + trial. Implemented by billing.
type PersonalProvisioner interface {
	EnsurePersonal(ctx context.Context, userID string) error
}
