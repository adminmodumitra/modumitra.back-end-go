package postgres

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
)

type PlanModel struct {
	PublicID          string               `gorm:"column:public_id;primaryKey"`
	Code              string               `gorm:"column:code"`
	Name              string               `gorm:"column:name"`
	Interval          string               `gorm:"column:interval"`
	Currency          string               `gorm:"column:currency"`
	UnitAmountMinor   int                  `gorm:"column:unit_amount_minor"`
	SeatsIncluded     int                  `gorm:"column:seats_included"`
	MaxDevicesPerUser int                  `gorm:"column:max_devices_per_user"`
	TrialDays         int                  `gorm:"column:trial_days"`
	Sellable          bool                 `gorm:"column:sellable"`
	Active            bool                 `gorm:"column:active"`
	Capabilities      billing.Capabilities `gorm:"column:capabilities;serializer:json"`
	Provider          *string              `gorm:"column:provider"`
	ProviderPriceID   *string              `gorm:"column:provider_price_id"`
	SortOrder         int                  `gorm:"column:sort_order"`
	CreatedAt         time.Time            `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time            `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (PlanModel) TableName() string { return "plans" }

func planFromModel(m *PlanModel) *billing.Plan {
	return &billing.Plan{
		ID: m.PublicID, Code: m.Code, Name: m.Name, Interval: m.Interval,
		Currency: m.Currency, UnitAmountMinor: m.UnitAmountMinor, SeatsIncluded: m.SeatsIncluded,
		MaxDevicesPerUser: m.MaxDevicesPerUser, TrialDays: m.TrialDays, Sellable: m.Sellable,
		Active: m.Active, Capabilities: m.Capabilities, Provider: deref(m.Provider),
		ProviderPriceID: deref(m.ProviderPriceID), SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

type SubscriptionModel struct {
	PublicID               string     `gorm:"column:public_id;primaryKey"`
	OrganizationID         string     `gorm:"column:organization_id"`
	PlanID                 string     `gorm:"column:plan_id"`
	PlanCode               string     `gorm:"column:plan_code"`
	Status                 string     `gorm:"column:status"`
	Provider               string     `gorm:"column:provider"`
	ProviderSubscriptionID *string    `gorm:"column:provider_subscription_id"`
	SeatsIncluded          int        `gorm:"column:seats_included"`
	TrialEndsAt            *time.Time `gorm:"column:trial_ends_at"`
	CurrentPeriodStart     time.Time  `gorm:"column:current_period_start"`
	CurrentPeriodEnd       time.Time  `gorm:"column:current_period_end"`
	CancelAtPeriodEnd      bool       `gorm:"column:cancel_at_period_end"`
	CanceledAt             *time.Time `gorm:"column:canceled_at"`
	GraceUntil             *time.Time `gorm:"column:grace_until"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (SubscriptionModel) TableName() string { return "subscriptions" }

func subToModel(s *billing.Subscription) SubscriptionModel {
	return SubscriptionModel{
		PublicID: s.ID, OrganizationID: s.OrganizationID, PlanID: s.PlanID, PlanCode: s.PlanCode,
		Status: s.Status, Provider: s.Provider, ProviderSubscriptionID: emptyToNil(s.ProviderSubscriptionID),
		SeatsIncluded: s.SeatsIncluded, TrialEndsAt: utcPtr(s.TrialEndsAt),
		CurrentPeriodStart: s.CurrentPeriodStart.UTC(), CurrentPeriodEnd: s.CurrentPeriodEnd.UTC(),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd, CanceledAt: utcPtr(s.CanceledAt), GraceUntil: utcPtr(s.GraceUntil),
		CreatedAt: s.CreatedAt.UTC(), UpdatedAt: s.UpdatedAt.UTC(),
	}
}

func subFromModel(m *SubscriptionModel) *billing.Subscription {
	return &billing.Subscription{
		ID: m.PublicID, OrganizationID: m.OrganizationID, PlanID: m.PlanID, PlanCode: m.PlanCode,
		Status: m.Status, Provider: m.Provider, ProviderSubscriptionID: deref(m.ProviderSubscriptionID),
		SeatsIncluded: m.SeatsIncluded, TrialEndsAt: utcPtr(m.TrialEndsAt),
		CurrentPeriodStart: m.CurrentPeriodStart.UTC(), CurrentPeriodEnd: m.CurrentPeriodEnd.UTC(),
		CancelAtPeriodEnd: m.CancelAtPeriodEnd, CanceledAt: utcPtr(m.CanceledAt), GraceUntil: utcPtr(m.GraceUntil),
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

type InvoiceModel struct {
	PublicID          string     `gorm:"column:public_id;primaryKey"`
	OrganizationID    string     `gorm:"column:organization_id"`
	SubscriptionID    *string    `gorm:"column:subscription_id"`
	Provider          string     `gorm:"column:provider"`
	ProviderInvoiceID string     `gorm:"column:provider_invoice_id"`
	Number            *string    `gorm:"column:number"`
	Status            string     `gorm:"column:status"`
	Currency          string     `gorm:"column:currency"`
	TotalMinor        int        `gorm:"column:total_minor"`
	AmountDueMinor    int        `gorm:"column:amount_due_minor"`
	HostedInvoiceURL  *string    `gorm:"column:hosted_invoice_url"`
	InvoicePDFURL     *string    `gorm:"column:invoice_pdf_url"`
	PeriodStart       *time.Time `gorm:"column:period_start"`
	PeriodEnd         *time.Time `gorm:"column:period_end"`
	PaidAt            *time.Time `gorm:"column:paid_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (InvoiceModel) TableName() string { return "invoices" }

func invToModel(inv *billing.Invoice) InvoiceModel {
	return InvoiceModel{
		PublicID: inv.ID, OrganizationID: inv.OrganizationID, SubscriptionID: inv.SubscriptionID,
		Provider: inv.Provider, ProviderInvoiceID: inv.ProviderInvoiceID, Number: inv.Number,
		Status: inv.Status, Currency: inv.Currency, TotalMinor: inv.TotalMinor, AmountDueMinor: inv.AmountDueMinor,
		HostedInvoiceURL: inv.HostedInvoiceURL, InvoicePDFURL: inv.InvoicePDFURL,
		PeriodStart: utcPtr(inv.PeriodStart), PeriodEnd: utcPtr(inv.PeriodEnd), PaidAt: utcPtr(inv.PaidAt),
		CreatedAt: inv.CreatedAt.UTC(), UpdatedAt: inv.UpdatedAt.UTC(),
	}
}

func invFromModel(m *InvoiceModel) *billing.Invoice {
	return &billing.Invoice{
		ID: m.PublicID, OrganizationID: m.OrganizationID, SubscriptionID: m.SubscriptionID,
		Provider: m.Provider, ProviderInvoiceID: m.ProviderInvoiceID, Number: m.Number,
		Status: m.Status, Currency: m.Currency, TotalMinor: m.TotalMinor, AmountDueMinor: m.AmountDueMinor,
		HostedInvoiceURL: m.HostedInvoiceURL, InvoicePDFURL: m.InvoicePDFURL,
		PeriodStart: utcPtr(m.PeriodStart), PeriodEnd: utcPtr(m.PeriodEnd), PaidAt: utcPtr(m.PaidAt),
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

type EventModel struct {
	PublicID        string         `gorm:"column:public_id;primaryKey"`
	Provider        string         `gorm:"column:provider"`
	ProviderEventID string         `gorm:"column:provider_event_id"`
	Type            string         `gorm:"column:type"`
	OrganizationID  *string        `gorm:"column:organization_id"`
	SubscriptionID  *string        `gorm:"column:subscription_id"`
	Status          string         `gorm:"column:status"`
	PayloadDigest   string         `gorm:"column:payload_digest"`
	Summary         map[string]any `gorm:"column:summary;serializer:json"`
	Error           *string        `gorm:"column:error"`
	ProcessedAt     *time.Time     `gorm:"column:processed_at"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime:false"`
}

func (EventModel) TableName() string { return "billing_events" }

func eventToModel(e *billing.Event) EventModel {
	sum := e.Summary
	if sum == nil {
		sum = map[string]any{}
	}
	return EventModel{
		PublicID: e.ID, Provider: e.Provider, ProviderEventID: e.ProviderEventID, Type: e.Type,
		OrganizationID: e.OrganizationID, SubscriptionID: e.SubscriptionID, Status: e.Status,
		PayloadDigest: e.PayloadDigest, Summary: sum, Error: emptyToNil(e.Error),
		ProcessedAt: utcPtr(e.ProcessedAt), CreatedAt: e.CreatedAt.UTC(),
	}
}

func eventFromModel(m *EventModel) *billing.Event {
	sum := m.Summary
	if sum == nil {
		sum = map[string]any{}
	}
	return &billing.Event{
		ID: m.PublicID, Provider: m.Provider, ProviderEventID: m.ProviderEventID, Type: m.Type,
		OrganizationID: m.OrganizationID, SubscriptionID: m.SubscriptionID, Status: m.Status,
		PayloadDigest: m.PayloadDigest, Summary: sum, Error: deref(m.Error),
		ProcessedAt: utcPtr(m.ProcessedAt), CreatedAt: m.CreatedAt.UTC(),
	}
}
