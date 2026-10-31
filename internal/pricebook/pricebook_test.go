package pricebook

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestValidateBook(t *testing.T) {
	ok := []string{`{}`, ` {"carcassPerSqft": 450} `}
	bad := []string{``, `null`, `[]`, `"rates"`, `42`, `{"a":`}
	for _, s := range ok {
		if err := ValidateBook(json.RawMessage(s)); err != nil {
			t.Errorf("%q: %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ValidateBook(json.RawMessage(s)); err == nil {
			t.Errorf("%q: want error", s)
		}
	}
}

func TestMemoryRepo(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	got, err := m.Get(ctx, "org_1")
	if err != nil || got != nil {
		t.Fatalf("empty: %v %v", got, err)
	}
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	if err := m.Put(ctx, Book{OrgID: "org_1", Book: json.RawMessage(`{"a":1}`), UpdatedAt: now, UpdatedBy: "usr_1"}); err != nil {
		t.Fatal(err)
	}
	got, err = m.Get(ctx, "org_1")
	if err != nil || got == nil || string(got.Book) != `{"a":1}` || got.UpdatedBy != "usr_1" {
		t.Fatalf("got %+v %v", got, err)
	}
	got.Book[1] = 'X' // callers cannot mutate the stored copy
	again, _ := m.Get(ctx, "org_1")
	if string(again.Book) != `{"a":1}` {
		t.Fatalf("aliasing: %s", again.Book)
	}
}
