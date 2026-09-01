package identity

import (
	"context"
	"sync"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
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

func (m *MemUsers) GetByID(_ context.Context, id string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneUser(u), nil
}

func (m *MemUsers) GetByEmailNormalized(_ context.Context, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneUser(u), nil
}

func (m *MemUsers) Insert(_ context.Context, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.usersByEmail[u.EmailNormalized]; ok {
		return utils.E(utils.CodeEmailTaken)
	}
	stored := cloneUser(u)
	m.users[u.ID] = stored
	m.usersByEmail[u.EmailNormalized] = stored
	return nil
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

func (m *MemUsers) SetEmailVerified(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	t := at
	u.EmailVerifiedAt = &t
	u.UpdatedAt = at
	return nil
}

func (m *MemUsers) SetLastLogin(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	t := at
	u.LastLoginAt = &t
	u.UpdatedAt = at
	return nil
}

func (m *MemSessions) GetByID(_ context.Context, id string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSession(s), nil
}

func (m *MemSessions) GetByRefreshHash(_ context.Context, hash string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessionsHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSession(s), nil
}

func (m *MemSessions) GetByRefreshHashForUpdate(ctx context.Context, hash string) (*Session, error) {
	return m.GetByRefreshHash(ctx, hash)
}

func (m *MemSessions) Insert(_ context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := cloneSession(s)
	m.sessions[s.ID] = stored
	m.sessionsHash[s.RefreshHash] = stored
	return nil
}

func (m *MemSessions) RevokeAndReplace(_ context.Context, oldID, newID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[oldID]
	if !ok {
		return ErrNotFound
	}
	t := at
	id := newID
	s.RevokedAt = &t
	s.ReplacedBy = &id
	s.UpdatedAt = at
	return nil
}

func (m *MemSessions) MarkReuseAndRevokeFamily(_ context.Context, sessionID, familyID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := at
	if s, ok := m.sessions[sessionID]; ok {
		s.ReuseDetectedAt = &t
		s.UpdatedAt = at
		if s.RevokedAt == nil {
			s.RevokedAt = &t
		}
	}
	for _, s := range m.sessions {
		if s.FamilyID == familyID && s.RevokedAt == nil {
			s.RevokedAt = &t
			s.UpdatedAt = at
		}
	}
	return nil
}

func (m *MemSessions) Revoke(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return ErrNotFound
	}
	if s.RevokedAt == nil {
		t := at
		s.RevokedAt = &t
		s.UpdatedAt = at
	}
	return nil
}

func (m *MemSessions) RevokeAllForUser(_ context.Context, userID string, at time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	t := at
	for _, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			s.RevokedAt = &t
			s.UpdatedAt = at
			n++
		}
	}
	return n, nil
}

func (m *MemTokens) Insert(_ context.Context, t *EmailToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := cloneToken(t)
	m.tokens[t.ID] = stored
	m.tokensByHash[t.TokenHash] = stored
	return nil
}

func (m *MemTokens) GetByHash(_ context.Context, hash string) (*EmailToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokensByHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneToken(t), nil
}

func (m *MemTokens) Consume(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[id]
	if !ok || t.ConsumedAt != nil {
		return ErrNotFound
	}
	v := at
	t.ConsumedAt = &v
	return nil
}

func (m *MemTokens) InvalidateOpen(_ context.Context, userID, purpose string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, tok := range m.tokens {
		if tok.UserID == userID && tok.Purpose == purpose && tok.ConsumedAt == nil {
			consumed := at
			tok.ConsumedAt = &consumed
		}
	}
	return nil
}

var (
	_ UserRepo       = (*MemUsers)(nil)
	_ SessionRepo    = (*MemSessions)(nil)
	_ EmailTokenRepo = (*MemTokens)(nil)
	_ TxRunner       = (*Memory)(nil)
)
