package cache

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

type memCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMem() *memCache { return &memCache{data: map[string][]byte{}} }

func (m *memCache) Get(_ context.Context, key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return nil, false
	}
	out := append([]byte(nil), v...)
	return out, true
}

func (m *memCache) Set(_ context.Context, key string, value []byte, _ time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = append([]byte(nil), value...)
}

func (m *memCache) Delete(_ context.Context, keys ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.data, k)
	}
}

func (m *memCache) DeletePrefix(_ context.Context, prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			delete(m.data, k)
		}
	}
}

func TestNoopMiss(t *testing.T) {
	var n Noop
	if _, ok := n.Get(context.Background(), "x"); ok {
		t.Fatal("noop should miss")
	}
}

func TestMemReadThroughJSON(t *testing.T) {
	ctx := context.Background()
	c := newMem()
	type payload struct {
		User string `json:"user"`
	}
	key := Me("usr_1")
	if GetJSON(ctx, c, key, &payload{}) {
		t.Fatal("expected miss")
	}
	SetJSON(ctx, c, key, payload{User: "ada"}, TTLMe)
	var got payload
	if !GetJSON(ctx, c, key, &got) {
		t.Fatal("expected hit")
	}
	if got.User != "ada" {
		t.Fatalf("got=%q", got.User)
	}
	c.Delete(ctx, key)
	if GetJSON(ctx, c, key, &got) {
		t.Fatal("expected miss after delete")
	}
}

func TestKeyScheme(t *testing.T) {
	if Me("u1") != "cabinet:me:u1" {
		t.Fatal(Me("u1"))
	}
	if Capabilities("u1", "") != "cabinet:me:capabilities:u1" {
		t.Fatal(Capabilities("u1", ""))
	}
	if Capabilities("u1", "org_1") != "cabinet:me:capabilities:u1:org:org_1" {
		t.Fatal(Capabilities("u1", "org_1"))
	}
	if OrgRoles("o1") != "cabinet:org:o1:roles" {
		t.Fatal(OrgRoles("o1"))
	}
	if BillingPlans() != "cabinet:billing:plans" {
		t.Fatal(BillingPlans())
	}
	if BillingSubUser("u1") != "cabinet:billing:sub:user:u1" {
		t.Fatal(BillingSubUser("u1"))
	}
	if BillingSubOrg("o1") != "cabinet:billing:sub:org:o1" {
		t.Fatal(BillingSubOrg("o1"))
	}
	if !strings.HasPrefix(Devices("u1", true, "c", 10), "cabinet:devices:user:u1:") {
		t.Fatal(Devices("u1", true, "c", 10))
	}
}

func TestDeletePrefix(t *testing.T) {
	ctx := context.Background()
	c := newMem()
	c.Set(ctx, DevicesPrefix("u1"), []byte("a"), TTLDevices)
	c.Set(ctx, Devices("u1", true, "x", 5), []byte("b"), TTLDevices)
	c.Set(ctx, Me("u1"), []byte("c"), TTLMe)
	c.DeletePrefix(ctx, DevicesPrefix("u1"))
	if _, ok := c.Get(ctx, Devices("u1", true, "x", 5)); ok {
		t.Fatal("devices key should be gone")
	}
	if _, ok := c.Get(ctx, Me("u1")); !ok {
		t.Fatal("me key should remain")
	}
}

func TestFromClientNil(t *testing.T) {
	c := FromClient(nil)
	if _, ok := c.Get(context.Background(), "k"); ok {
		t.Fatal("nil client should noop")
	}
}
