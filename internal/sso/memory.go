package sso

import (
	"context"
	"sync"
)

type Memory struct {
	mu          sync.Mutex
	identities  map[string]*Identity
	bySubject   map[string]*Identity // provider+"\x00"+subject
	byUserProv  map[string]*Identity // userID+"\x00"+provider
	states      map[string]*State    // state token
	devices     map[string]*DeviceCode
	devicesUser map[string]*DeviceCode
}

func NewMemory() *Memory {
	return &Memory{
		identities:  map[string]*Identity{},
		bySubject:   map[string]*Identity{},
		byUserProv:  map[string]*Identity{},
		states:      map[string]*State{},
		devices:     map[string]*DeviceCode{},
		devicesUser: map[string]*DeviceCode{},
	}
}

func (m *Memory) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (m *Memory) Identities() *MemIdentities   { return &MemIdentities{Memory: m} }
func (m *Memory) States() *MemStates           { return &MemStates{Memory: m} }
func (m *Memory) DeviceCodes() *MemDeviceCodes { return &MemDeviceCodes{Memory: m} }

type MemIdentities struct{ *Memory }
type MemStates struct{ *Memory }
type MemDeviceCodes struct{ *Memory }

func cloneIdentity(i *Identity) *Identity {
	if i == nil {
		return nil
	}
	cp := *i
	return &cp
}

func cloneState(s *State) *State {
	if s == nil {
		return nil
	}
	cp := *s
	if s.ConsumedAt != nil {
		t := *s.ConsumedAt
		cp.ConsumedAt = &t
	}
	return &cp
}

func cloneDevice(d *DeviceCode) *DeviceCode {
	if d == nil {
		return nil
	}
	cp := *d
	if d.LastPollAt != nil {
		t := *d.LastPollAt
		cp.LastPollAt = &t
	}
	if d.AuthorizedAt != nil {
		t := *d.AuthorizedAt
		cp.AuthorizedAt = &t
	}
	if d.ConsumedAt != nil {
		t := *d.ConsumedAt
		cp.ConsumedAt = &t
	}
	return &cp
}

func subKey(provider, subject string) string { return provider + "\x00" + subject }
func upKey(userID, provider string) string   { return userID + "\x00" + provider }

var (
	_ IdentityRepo   = (*MemIdentities)(nil)
	_ StateRepo      = (*MemStates)(nil)
	_ DeviceCodeRepo = (*MemDeviceCodes)(nil)
	_ TxRunner       = (*Memory)(nil)
)
