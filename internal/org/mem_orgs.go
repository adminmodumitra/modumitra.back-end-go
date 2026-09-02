package org

import (
	"context"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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

func (m *MemOrgs) SetProviderCustomer(_ context.Context, id, provider, customerID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orgs[id]
	if !ok {
		return ErrNotFound
	}
	o.Provider = provider
	o.ProviderCustomerID = customerID
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
