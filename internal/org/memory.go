package org

import (
	"context"
	"sync"
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

var (
	_ OrgRepo        = (*MemOrgs)(nil)
	_ MemberRepo     = (*MemMembers)(nil)
	_ InvitationRepo = (*MemInvites)(nil)
	_ TxRunner       = (*Memory)(nil)
)
