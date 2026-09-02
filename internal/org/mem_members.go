package org

import (
	"context"
	"sort"
)

func (m *MemMembers) GetByOrgUser(_ context.Context, orgID, userID string) (*Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mem, ok := m.membersByOrgUser[orgUserKey(orgID, userID)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneMember(mem), nil
}

func (m *MemMembers) ListByUser(_ context.Context, userID, cursor string, limit int) ([]MembershipItem, string, error) {
	ct, cid, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var items []MembershipItem
	for _, mem := range m.members {
		if mem.UserID != userID {
			continue
		}
		if mem.Status != MemberActive && mem.Status != MemberSuspended {
			continue
		}
		o, ok := m.orgs[mem.OrganizationID]
		if !ok || o.Deleted() {
			continue
		}
		if !BeforeCursor(mem.CreatedAt, mem.ID, ct, cid) {
			continue
		}
		items = append(items, MembershipItem{Organization: *cloneOrg(o), Membership: *cloneMember(mem)})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Membership.CreatedAt.Equal(items[j].Membership.CreatedAt) {
			return items[i].Membership.ID > items[j].Membership.ID
		}
		return items[i].Membership.CreatedAt.After(items[j].Membership.CreatedAt)
	})
	next := ""
	if limit > 0 && len(items) > limit {
		last := items[limit-1]
		next = EncodeCursor(last.Membership.CreatedAt, last.Membership.ID)
		items = items[:limit]
	}
	return items, next, nil
}

func (m *MemMembers) Update(_ context.Context, mem *Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored, ok := m.members[mem.ID]
	if !ok {
		return ErrNotFound
	}
	cp := cloneMember(mem)
	*stored = *cp
	m.membersByOrgUser[orgUserKey(stored.OrganizationID, stored.UserID)] = stored
	return nil
}

func (m *MemMembers) CountActiveOwners(_ context.Context, orgID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, mem := range m.members {
		if mem.OrganizationID == orgID && mem.Status == MemberActive && mem.Role == RoleOwner {
			n++
		}
	}
	return n, nil
}

func (m *MemInvites) Update(_ context.Context, inv *Invitation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored, ok := m.invites[inv.ID]
	if !ok {
		return ErrNotFound
	}
	oldHash := stored.TokenHash
	cp := cloneInvite(inv)
	*stored = *cp
	if oldHash != stored.TokenHash {
		delete(m.invitesByHash, oldHash)
		m.invitesByHash[stored.TokenHash] = stored
	}
	return nil
}
