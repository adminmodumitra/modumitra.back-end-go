package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
)

func TestTableName(t *testing.T) {
	if (tables.OrgPriceBook{}).TableName() != "org_price_books" {
		t.Fatal("org_price_books")
	}
}

func TestModelRoundTrip(t *testing.T) {
	ist := time.Date(2026, 10, 31, 12, 0, 0, 0, time.FixedZone("IST", 19800))
	b := pricebook.Book{OrgID: "org_1", Book: json.RawMessage(`{"a":1}`), UpdatedAt: ist, UpdatedBy: "usr_1"}
	m := toModel(b)
	if m.UpdatedAt.Location() != time.UTC || m.Book != `{"a":1}` {
		t.Fatalf("%+v", m)
	}
	got := fromModel(&m)
	if got.OrgID != "org_1" || string(got.Book) != `{"a":1}` || !got.UpdatedAt.Equal(ist) || got.UpdatedBy != "usr_1" {
		t.Fatalf("%+v", got)
	}
}
