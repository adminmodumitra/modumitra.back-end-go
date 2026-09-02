package billing

import (
	"context"
	"sort"
)

func (m *MemPlans) ListSellable(_ context.Context) ([]Plan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Plan
	for _, p := range m.plans {
		if p.Sellable && p.Active {
			out = append(out, *clonePlan(p))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	if out == nil {
		out = []Plan{}
	}
	return out, nil
}

func (m *MemPlans) GetByID(_ context.Context, id string) (*Plan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return clonePlan(p), nil
}
