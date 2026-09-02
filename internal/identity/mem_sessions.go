package identity

import (
	"context"
	"time"
)

// Memory is a shared in-memory store. Use Users(), Sessions(), EmailTokens()
// for the repo interfaces (method names overlap across repos).

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
