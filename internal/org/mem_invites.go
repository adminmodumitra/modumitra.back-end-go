package org

import (
	"context"
	"sort"
)

func (m *MemMembers) ListByOrg(_ context.Context, orgID, status, role, cursor string, limit int) ([]Membership, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*Membership
	for _, mem := range m.members {
		if mem.OrganizationID != orgID || mem.Status == MemberRemoved {
			continue
		}
		if status != "" && mem.Status != status {
			continue
		}
		if role != "" && mem.Role != role {
			continue
		}
		if !BeforeCursor(mem.CreatedAt, mem.ID, ct, cid) {
			continue
		}
		out = append(out, mem)
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
	res := make([]Membership, len(out))
	for i, mem := range out {
		res[i] = *cloneMember(mem)
	}
	return res, next, nil
}

func (m *MemInvites) GetByTokenHash(_ context.Context, hash string) (*Invitation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invitesByHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneInvite(inv), nil
}

func (m *MemInvites) GetPendingByOrgEmail(_ context.Context, orgID, email string) (*Invitation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, inv := range m.invites {
		if inv.OrganizationID == orgID && inv.Email == email && inv.Status == InvitePending {
			return cloneInvite(inv), nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemInvites) ListByOrg(_ context.Context, orgID, status, cursor string, limit int) ([]Invitation, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*Invitation
	for _, inv := range m.invites {
		if inv.OrganizationID != orgID {
			continue
		}
		if status != "" && inv.Status != status {
			continue
		}
		if !BeforeCursor(inv.CreatedAt, inv.ID, ct, cid) {
			continue
		}
		out = append(out, inv)
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
	res := make([]Invitation, len(out))
	for i, inv := range out {
		res[i] = *cloneInvite(inv)
	}
	return res, next, nil
}
