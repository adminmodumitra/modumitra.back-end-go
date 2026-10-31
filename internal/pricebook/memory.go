package pricebook

import (
	"context"
	"sync"
)

// Memory is the in-process Repo for tests and local runs without Postgres.
type Memory struct {
	mu    sync.Mutex
	books map[string]Book
}

func NewMemory() *Memory { return &Memory{books: map[string]Book{}} }

func (m *Memory) Get(_ context.Context, orgID string) (*Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.books[orgID]
	if !ok {
		return nil, nil
	}
	b.Book = append([]byte(nil), b.Book...)
	return &b, nil
}

func (m *Memory) Put(_ context.Context, b Book) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b.Book = append([]byte(nil), b.Book...)
	m.books[b.OrgID] = b
	return nil
}

var _ Repo = (*Memory)(nil)
