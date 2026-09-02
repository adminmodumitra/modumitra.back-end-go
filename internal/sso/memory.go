package sso

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
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

func (m *MemStates) Insert(_ context.Context, st *State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[st.State] = cloneState(st)
	return nil
}

func (m *MemStates) Consume(_ context.Context, state, provider string, now time.Time) (*State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.states[state]
	if !ok || st.Provider != provider {
		return nil, ErrInvalidState
	}
	if st.ConsumedAt != nil || !st.ExpiresAt.After(now) {
		return nil, ErrInvalidState
	}
	t := now
	st.ConsumedAt = &t
	return cloneState(st), nil
}

func (m *MemDeviceCodes) Insert(_ context.Context, d *DeviceCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := cloneDevice(d)
	m.devices[d.DeviceCode] = stored
	m.devicesUser[strings.ToUpper(d.UserCode)] = stored
	return nil
}

func (m *MemDeviceCodes) GetByDeviceCode(_ context.Context, deviceCode string) (*DeviceCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.devices[deviceCode]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneDevice(d), nil
}

func (m *MemDeviceCodes) GetByUserCode(_ context.Context, userCode string) (*DeviceCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.devicesUser[strings.ToUpper(userCode)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneDevice(d), nil
}

func (m *MemDeviceCodes) Update(_ context.Context, d *DeviceCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.devices[d.DeviceCode]; !ok {
		return ErrNotFound
	}
	stored := cloneDevice(d)
	m.devices[d.DeviceCode] = stored
	m.devicesUser[strings.ToUpper(d.UserCode)] = stored
	return nil
}

var (
	_ IdentityRepo   = (*MemIdentities)(nil)
	_ StateRepo      = (*MemStates)(nil)
	_ DeviceCodeRepo = (*MemDeviceCodes)(nil)
	_ TxRunner       = (*Memory)(nil)
)
