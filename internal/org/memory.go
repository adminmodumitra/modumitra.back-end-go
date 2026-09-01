package org

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Memory struct {
	mu               sync.Mutex
	orgs             map[string]*Organization
	members          map[string]*Membership
	membersByOrgUser map[string]*Membership
	invites          map[string]*Invitation
	invitesByHash    map[string]*Invitation
}

func NewMemory() *Memory {
	return &Memory{
		orgs:             map[string]*Organization{},
		members:          map[string]*Membership{},
		membersByOrgUser: map[string]*Membership{},
		invites:          map[string]*Invitation{},
		invitesByHash:    map[string]*Invitation{},
	}
}

func (m *Memory) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (m *Memory) Orgs() *MemOrgs       { return &MemOrgs{Memory: m} }
func (m *Memory) Members() *MemMembers { return &MemMembers{Memory: m} }
func (m *Memory) Invites() *MemInvites { return &MemInvites{Memory: m} }

type MemOrgs struct{ *Memory }
type MemMembers struct{ *Memory }
type MemInvites struct{ *Memory }

func orgUserKey(orgID, userID string) string { return orgID + "\x00" + userID }

func cloneOrg(o *Organization) *Organization {
	if o == nil {
		return nil
	}
	cp := *o
	if o.TrialEndsAt != nil {
		t := *o.TrialEndsAt
		cp.TrialEndsAt = &t
	}
	if o.DeletedAt != nil {
		t := *o.DeletedAt
		cp.DeletedAt = &t
	}
	return &cp
}

func cloneMember(m *Membership) *Membership {
	if m == nil {
		return nil
	}
	cp := *m
	if m.InvitedAt != nil {
		t := *m.InvitedAt
		cp.InvitedAt = &t
	}
	if m.ActivatedAt != nil {
		t := *m.ActivatedAt
		cp.ActivatedAt = &t
	}
	if m.SuspendedAt != nil {
		t := *m.SuspendedAt
		cp.SuspendedAt = &t
	}
	if m.RemovedAt != nil {
		t := *m.RemovedAt
		cp.RemovedAt = &t
	}
	return &cp
}

func cloneInvite(inv *Invitation) *Invitation {
	if inv == nil {
		return nil
	}
	cp := *inv
	if inv.AcceptedAt != nil {
		t := *inv.AcceptedAt
		cp.AcceptedAt = &t
	}
	if inv.AcceptedByUserID != nil {
		s := *inv.AcceptedByUserID
		cp.AcceptedByUserID = &s
	}
	if inv.RevokedAt != nil {
		t := *inv.RevokedAt
		cp.RevokedAt = &t
	}
	if inv.RevokedByUserID != nil {
		s := *inv.RevokedByUserID
		cp.RevokedByUserID = &s
	}
	if inv.ResentAt != nil {
		t := *inv.ResentAt
		cp.ResentAt = &t
	}
	return &cp
}

func (m *MemOrgs) Insert(_ context.Context, o *Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orgs[o.ID] = cloneOrg(o)
	return nil
}

func (m *MemOrgs) GetByID(_ context.Context, id string) (*Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orgs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneOrg(o), nil
}

func (m *MemOrgs) UpdateName(_ context.Context, id, name string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orgs[id]
	if !ok {
		return ErrNotFound
	}
	o.Name = name
	o.UpdatedAt = at
	return nil
}

func (m *MemOrgs) SoftDelete(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orgs[id]
	if !ok {
		return ErrNotFound
	}
	t := at
	o.Status = OrgDeleted
	o.DeletedAt = &t
	o.UpdatedAt = at
	return nil
}

func (m *MemOrgs) SetOwner(_ context.Context, id, ownerUserID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orgs[id]
	if !ok {
		return ErrNotFound
	}
	o.OwnerUserID = ownerUserID
	o.UpdatedAt = at
	return nil
}

func (m *MemMembers) Insert(_ context.Context, mem *Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := orgUserKey(mem.OrganizationID, mem.UserID)
	if _, ok := m.membersByOrgUser[key]; ok {
		return utils.E(utils.CodeAlreadyMember)
	}
	stored := cloneMember(mem)
	m.members[mem.ID] = stored
	m.membersByOrgUser[key] = stored
	return nil
}

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

func (m *MemInvites) Insert(_ context.Context, inv *Invitation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.invitesByHash[inv.TokenHash]; ok {
		return utils.E(utils.CodeConflict)
	}
	if inv.Status == InvitePending {
		for _, existing := range m.invites {
			if existing.OrganizationID == inv.OrganizationID && existing.Email == inv.Email && existing.Status == InvitePending {
				return utils.E(utils.CodeInvitationPending)
			}
		}
	}
	stored := cloneInvite(inv)
	m.invites[inv.ID] = stored
	m.invitesByHash[inv.TokenHash] = stored
	return nil
}

func (m *MemInvites) GetByID(_ context.Context, id string) (*Invitation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invites[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneInvite(inv), nil
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

var (
	_ OrgRepo        = (*MemOrgs)(nil)
	_ MemberRepo     = (*MemMembers)(nil)
	_ InvitationRepo = (*MemInvites)(nil)
	_ TxRunner       = (*Memory)(nil)
)
