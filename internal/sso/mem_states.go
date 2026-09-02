package sso

import (
	"context"
	"time"
)

func (m *MemStates) Insert(_ context.Context, st *State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[st.State] = cloneState(st)
	return nil
}

func (m *MemStates) Consume(_ context.Context, state, provider string, now time.Time) (*State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.states[state]
	if !ok || st.Provider != provider {
		return nil, ErrInvalidState
	}
	if st.ConsumedAt != nil || !st.ExpiresAt.After(now) {
		return nil, ErrInvalidState
	}
	t := now
	st.ConsumedAt = &t
	return cloneState(st), nil
}
