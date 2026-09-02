package sso

import (
	"context"
	"sort"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (m *MemIdentities) Insert(_ context.Context, id *Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.bySubject[subKey(id.Provider, id.ProviderUserID)]; ok {
		return utils.E(utils.CodeIdentityAlreadyLinked)
	}
	if _, ok := m.byUserProv[upKey(id.UserID, id.Provider)]; ok {
		return utils.E(utils.CodeProviderAlreadyLinked)
	}
	stored := cloneIdentity(id)
	m.identities[id.ID] = stored
	m.bySubject[subKey(id.Provider, id.ProviderUserID)] = stored
	m.byUserProv[upKey(id.UserID, id.Provider)] = stored
	return nil
}

func (m *MemIdentities) GetByProviderSubject(_ context.Context, provider, subject string) (*Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.bySubject[subKey(provider, subject)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneIdentity(id), nil
}

func (m *MemIdentities) GetByUserProvider(_ context.Context, userID, provider string) (*Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byUserProv[upKey(userID, provider)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneIdentity(id), nil
}

func (m *MemIdentities) ListByUser(_ context.Context, userID string) ([]Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Identity
	for _, id := range m.identities {
		if id.UserID == userID {
			out = append(out, *cloneIdentity(id))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LinkedAt.Before(out[j].LinkedAt) })
	if out == nil {
		out = []Identity{}
	}
	return out, nil
}

func (m *MemIdentities) DeleteByUserProvider(_ context.Context, userID, provider string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byUserProv[upKey(userID, provider)]
	if !ok {
		return ErrNotFound
	}
	delete(m.identities, id.ID)
	delete(m.bySubject, subKey(id.Provider, id.ProviderUserID))
	delete(m.byUserProv, upKey(userID, provider))
	return nil
}
