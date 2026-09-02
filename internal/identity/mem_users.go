package identity

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Memory is a shared in-memory store. Use Users(), Sessions(), EmailTokens()
// for the repo interfaces (method names overlap across repos).

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
