package identity

import (
	"context"
	"sync"
	"time"
)

// Memory is a shared in-memory store. Use Users(), Sessions(), EmailTokens()
// for the repo interfaces (method names overlap across repos).

type Memory struct {
	mu           sync.Mutex
	users        map[string]*User
	usersByEmail map[string]*User
	sessions     map[string]*Session
	sessionsHash map[string]*Session
	tokens       map[string]*EmailToken
	tokensByHash map[string]*EmailToken
}

func NewMemory() *Memory {
	return &Memory{
		users:        map[string]*User{},
		usersByEmail: map[string]*User{},
		sessions:     map[string]*Session{},
		sessionsHash: map[string]*Session{},
		tokens:       map[string]*EmailToken{},
		tokensByHash: map[string]*EmailToken{},
	}
}

func (m *Memory) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (m *Memory) Users() *MemUsers        { return &MemUsers{Memory: m} }
func (m *Memory) Sessions() *MemSessions  { return &MemSessions{Memory: m} }
func (m *Memory) EmailTokens() *MemTokens { return &MemTokens{Memory: m} }

type MemUsers struct{ *Memory }
type MemSessions struct{ *Memory }
type MemTokens struct{ *Memory }

func cloneUser(u *User) *User {
	if u == nil {
		return nil
	}
	cp := *u
	if u.EmailVerifiedAt != nil {
		t := *u.EmailVerifiedAt
		cp.EmailVerifiedAt = &t
	}
	if u.LastLoginAt != nil {
		t := *u.LastLoginAt
		cp.LastLoginAt = &t
	}
	return &cp
}

func cloneSession(s *Session) *Session {
	if s == nil {
		return nil
	}
	cp := *s
	if s.DeviceID != nil {
		v := *s.DeviceID
		cp.DeviceID = &v
	}
	if s.RevokedAt != nil {
		t := *s.RevokedAt
		cp.RevokedAt = &t
	}
	if s.ReuseDetectedAt != nil {
		t := *s.ReuseDetectedAt
		cp.ReuseDetectedAt = &t
	}
	if s.ReplacedBy != nil {
		v := *s.ReplacedBy
		cp.ReplacedBy = &v
	}
	return &cp
}

func cloneToken(t *EmailToken) *EmailToken {
	if t == nil {
		return nil
	}
	cp := *t
	if t.ConsumedAt != nil {
		v := *t.ConsumedAt
		cp.ConsumedAt = &v
	}
	return &cp
}

func (m *MemUsers) UpdatePassword(_ context.Context, id, hash string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	u.PasswordHash = hash
	u.UpdatedAt = at
	return nil
}

var (
	_ UserRepo       = (*MemUsers)(nil)
	_ SessionRepo    = (*MemSessions)(nil)
	_ EmailTokenRepo = (*MemTokens)(nil)
	_ TxRunner       = (*Memory)(nil)
)
