package device

import (
	"context"
)

func (m *MemEntitlements) Insert(_ context.Context, e *Entitlement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ents[e.ID] = cloneEntitlement(e)
	return nil
}
