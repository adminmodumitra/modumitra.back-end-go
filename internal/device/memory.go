package device

import (
	"context"
	"sync"
)

type Memory struct {
	mu   sync.Mutex
	devs map[string]*Device
	ents map[string]*Entitlement
}

func NewMemory() *Memory {
	return &Memory{
		devs: map[string]*Device{},
		ents: map[string]*Entitlement{},
	}
}

func (m *Memory) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (m *Memory) Devices() *MemDevices           { return &MemDevices{Memory: m} }
func (m *Memory) Entitlements() *MemEntitlements { return &MemEntitlements{Memory: m} }

type MemDevices struct{ *Memory }
type MemEntitlements struct{ *Memory }

func cloneDevice(d *Device) *Device {
	if d == nil {
		return nil
	}
	cp := *d
	if d.DeactivatedAt != nil {
		t := *d.DeactivatedAt
		cp.DeactivatedAt = &t
	}
	return &cp
}

func cloneEntitlement(e *Entitlement) *Entitlement {
	if e == nil {
		return nil
	}
	cp := *e
	if e.OrganizationID != nil {
		s := *e.OrganizationID
		cp.OrganizationID = &s
	}
	if e.DeviceID != nil {
		s := *e.DeviceID
		cp.DeviceID = &s
	}
	if e.PlanID != nil {
		s := *e.PlanID
		cp.PlanID = &s
	}
	if e.GraceUntil != nil {
		t := *e.GraceUntil
		cp.GraceUntil = &t
	}
	return &cp
}

func (m *Memory) EntitlementCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.ents)
}
