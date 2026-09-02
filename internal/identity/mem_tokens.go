package identity

import (
	"context"
	"time"
)

// Memory is a shared in-memory store. Use Users(), Sessions(), EmailTokens()
// for the repo interfaces (method names overlap across repos).

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
