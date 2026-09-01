package postgres

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

func TestTableNames(t *testing.T) {
	if (PlanModel{}).TableName() != "plans" {
		t.Fatal("plans")
	}
	if (SubscriptionModel{}).TableName() != "subscriptions" {
		t.Fatal("subscriptions")
	}
	if (InvoiceModel{}).TableName() != "invoices" {
		t.Fatal("invoices")
	}
	if (EventModel{}).TableName() != "billing_events" {
		t.Fatal("billing_events")
	}
}

func TestSubModelRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("IST", 19800))
	s := &billing.Subscription{
		ID: "sub_1", OrganizationID: "org_1", PlanID: billing.PlanIDTeam, PlanCode: billing.PlanTeam,
		Status: billing.StatusActive, Provider: billing.ProviderFake, ProviderSubscriptionID: "sub_fake_1",
		SeatsIncluded: 5, CurrentPeriodStart: now, CurrentPeriodEnd: now.AddDate(0, 1, 0),
		CreatedAt: now, UpdatedAt: now,
	}
	got := subFromModel(ptr(subToModel(s)))
	if got.ID != s.ID || got.PlanCode != s.PlanCode || got.ProviderSubscriptionID != s.ProviderSubscriptionID {
		t.Fatalf("%+v", got)
	}
	if got.CurrentPeriodStart.Location() != time.UTC {
		t.Fatalf("tz=%v", got.CurrentPeriodStart.Location())
	}
}

func TestMapGormNotFoundAndDuplicate(t *testing.T) {
	if !errors.Is(mapGorm(gorm.ErrRecordNotFound), billing.ErrNotFound) {
		t.Fatal("not found")
	}
	if !errors.Is(mapGorm(gorm.ErrDuplicatedKey), billing.ErrDuplicate) {
		t.Fatal("duplicate")
	}
}

func ptr[T any](v T) *T { return &v }
