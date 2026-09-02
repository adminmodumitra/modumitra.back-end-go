package postgres

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
)

func planFromModel(m *tables.Plan) *billing.Plan {
	return &billing.Plan{
		ID: m.PublicID, Code: m.Code, Name: m.Name, Interval: m.Interval,
		Currency: m.Currency, UnitAmountMinor: m.UnitAmountMinor, SeatsIncluded: m.SeatsIncluded,
		MaxDevicesPerUser: m.MaxDevicesPerUser, TrialDays: m.TrialDays, Sellable: m.Sellable,
		Active: m.Active, Capabilities: planCapsFrom(m.Capabilities), Provider: deref(m.Provider),
		ProviderPriceID: deref(m.ProviderPriceID), SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

func subToModel(s *billing.Subscription) tables.Subscription {
	return tables.Subscription{
		PublicID: s.ID, BillingAccountID: s.BillingAccountID, OrganizationID: emptyToNil(s.OrganizationID), PlanID: s.PlanID, PlanCode: s.PlanCode,
		Status: s.Status, Provider: s.Provider, ProviderSubscriptionID: emptyToNil(s.ProviderSubscriptionID),
		SeatsIncluded: s.SeatsIncluded, TrialEndsAt: utcPtr(s.TrialEndsAt),
		CurrentPeriodStart: s.CurrentPeriodStart.UTC(), CurrentPeriodEnd: s.CurrentPeriodEnd.UTC(),
		CancelAtPeriodEnd: s.CancelAtPeriodEnd, CanceledAt: utcPtr(s.CanceledAt), GraceUntil: utcPtr(s.GraceUntil),
		CreatedAt: s.CreatedAt.UTC(), UpdatedAt: s.UpdatedAt.UTC(),
	}
}

func subFromModel(m *tables.Subscription) *billing.Subscription {
	return &billing.Subscription{
		ID: m.PublicID, BillingAccountID: m.BillingAccountID, OrganizationID: deref(m.OrganizationID), PlanID: m.PlanID, PlanCode: m.PlanCode,
		Status: m.Status, Provider: m.Provider, ProviderSubscriptionID: deref(m.ProviderSubscriptionID),
		SeatsIncluded: m.SeatsIncluded, TrialEndsAt: utcPtr(m.TrialEndsAt),
		CurrentPeriodStart: m.CurrentPeriodStart.UTC(), CurrentPeriodEnd: m.CurrentPeriodEnd.UTC(),
		CancelAtPeriodEnd: m.CancelAtPeriodEnd, CanceledAt: utcPtr(m.CanceledAt), GraceUntil: utcPtr(m.GraceUntil),
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

func invToModel(inv *billing.Invoice) tables.Invoice {
	return tables.Invoice{
		PublicID: inv.ID, BillingAccountID: inv.BillingAccountID, OrganizationID: emptyToNil(inv.OrganizationID), SubscriptionID: inv.SubscriptionID,
		Provider: inv.Provider, ProviderInvoiceID: inv.ProviderInvoiceID, Number: inv.Number,
		Status: inv.Status, Currency: inv.Currency, TotalMinor: inv.TotalMinor, AmountDueMinor: inv.AmountDueMinor,
		HostedInvoiceURL: inv.HostedInvoiceURL, InvoicePDFURL: inv.InvoicePDFURL,
		PeriodStart: utcPtr(inv.PeriodStart), PeriodEnd: utcPtr(inv.PeriodEnd), PaidAt: utcPtr(inv.PaidAt),
		CreatedAt: inv.CreatedAt.UTC(), UpdatedAt: inv.UpdatedAt.UTC(),
	}
}

func invFromModel(m *tables.Invoice) *billing.Invoice {
	return &billing.Invoice{
		ID: m.PublicID, BillingAccountID: m.BillingAccountID, OrganizationID: deref(m.OrganizationID), SubscriptionID: m.SubscriptionID,
		Provider: m.Provider, ProviderInvoiceID: m.ProviderInvoiceID, Number: m.Number,
		Status: m.Status, Currency: m.Currency, TotalMinor: m.TotalMinor, AmountDueMinor: m.AmountDueMinor,
		HostedInvoiceURL: m.HostedInvoiceURL, InvoicePDFURL: m.InvoicePDFURL,
		PeriodStart: utcPtr(m.PeriodStart), PeriodEnd: utcPtr(m.PeriodEnd), PaidAt: utcPtr(m.PaidAt),
		CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(),
	}
}

func eventToModel(e *billing.Event) tables.BillingEvent {
	sum := e.Summary
	if sum == nil {
		sum = map[string]any{}
	}
	return tables.BillingEvent{
		PublicID: e.ID, Provider: e.Provider, ProviderEventID: e.ProviderEventID, Type: e.Type,
		BillingAccountID: e.BillingAccountID, OrganizationID: e.OrganizationID, SubscriptionID: e.SubscriptionID, Status: e.Status,
		PayloadDigest: e.PayloadDigest, Summary: sum, Error: emptyToNil(e.Error),
		ProcessedAt: utcPtr(e.ProcessedAt), CreatedAt: e.CreatedAt.UTC(),
	}
}

func eventFromModel(m *tables.BillingEvent) *billing.Event {
	sum := m.Summary
	if sum == nil {
		sum = map[string]any{}
	}
	return &billing.Event{
		ID: m.PublicID, Provider: m.Provider, ProviderEventID: m.ProviderEventID, Type: m.Type,
		BillingAccountID: m.BillingAccountID, OrganizationID: m.OrganizationID, SubscriptionID: m.SubscriptionID, Status: m.Status,
		PayloadDigest: m.PayloadDigest, Summary: sum, Error: deref(m.Error),
		ProcessedAt: utcPtr(m.ProcessedAt), CreatedAt: m.CreatedAt.UTC(),
	}
}

func planCapsFrom(c tables.PlanCapabilities) billing.Capabilities {
	return billing.Capabilities{
		CloudProjects: c.CloudProjects, OrganizationAssets: c.OrganizationAssets,
		ProposalBranding: c.ProposalBranding, EngineeringHandoff: c.EngineeringHandoff,
		MaxSeats: c.MaxSeats, MaxDevices: c.MaxDevices, MaxCloudProjects: c.MaxCloudProjects,
	}
}
