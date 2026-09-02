package sso

import (
	"context"
	"strings"
)

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
