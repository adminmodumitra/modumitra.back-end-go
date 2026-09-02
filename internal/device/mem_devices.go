package device

import (
	"context"
	"sort"
	"time"
)

func (m *MemDevices) Insert(_ context.Context, d *Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devs[d.ID] = cloneDevice(d)
	return nil
}

func (m *MemDevices) GetByID(_ context.Context, id string) (*Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.devs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneDevice(d), nil
}

func (m *MemDevices) ListByUser(_ context.Context, userID string, includeDeactivated bool, cursor string, limit int) ([]Device, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Device
	for _, d := range m.devs {
		if d.UserID != userID {
			continue
		}
		if !includeDeactivated && d.Status != StatusActive {
			continue
		}
		if !BeforeCursor(d.CreatedAt, d.ID, ct, cid) {
			continue
		}
		out = append(out, *cloneDevice(d))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	next := ""
	if limit > 0 && len(out) > limit {
		last := out[limit-1]
		next = EncodeCursor(last.CreatedAt, last.ID)
		out = out[:limit]
	}
	if out == nil {
		out = []Device{}
	}
	return out, next, nil
}

func (m *MemDevices) CountActive(_ context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, d := range m.devs {
		if d.UserID == userID && d.Status == StatusActive {
			n++
		}
	}
	return n, nil
}

func (m *MemDevices) UpdateName(_ context.Context, id, name string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.devs[id]
	if !ok {
		return ErrNotFound
	}
	d.Name = name
	d.UpdatedAt = at
	return nil
}

func (m *MemDevices) Deactivate(_ context.Context, id, reason string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.devs[id]
	if !ok {
		return ErrNotFound
	}
	t := at
	d.Status = StatusDeactivated
	d.DeactivatedAt = &t
	d.DeactivatedReason = reason
	d.UpdatedAt = at
	return nil
}
