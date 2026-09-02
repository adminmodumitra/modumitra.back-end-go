package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
)

func planView(p *billing.Plan) dto.PlanView {
	if p == nil {
		return dto.PlanView{}
	}
	return dto.PlanView{
		ID: p.ID, Plan: p.Code, Interval: p.Interval, Currency: p.Currency,
		UnitAmountMinor: p.UnitAmountMinor, IncludedSeats: p.SeatsIncluded,
		MaxDevicesPerUser: p.MaxDevicesPerUser,
		Capabilities: dto.PlanCapabilitiesView{
			CloudProjects:      p.Capabilities.CloudProjects,
			OrganizationAssets: p.Capabilities.OrganizationAssets,
			ProposalBranding:   p.Capabilities.ProposalBranding,
			EngineeringHandoff: p.Capabilities.EngineeringHandoff,
		},
		TrialDays: p.TrialDays,
	}
}

func subscriptionView(s *billing.Subscription) dto.SubscriptionView {
	if s == nil {
		return dto.SubscriptionView{}
	}
	return dto.SubscriptionView{
		ID: s.ID, OrganizationID: s.OrganizationID, PlanID: s.PlanID, Plan: s.PlanCode,
		Status: s.Status, Seats: s.SeatsIncluded,
		CurrentPeriodStart: rfcTime(s.CurrentPeriodStart), CurrentPeriodEnd: rfcTime(s.CurrentPeriodEnd),
		TrialEndsAt: rfcTimePtr(s.TrialEndsAt), CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		CanceledAt: rfcTimePtr(s.CanceledAt), GraceUntil: rfcTimePtr(s.GraceUntil),
		Provider: s.Provider, ProviderSubscriptionID: emptyPtr(s.ProviderSubscriptionID),
	}
}

func invoiceView(inv *billing.Invoice) dto.InvoiceView {
	if inv == nil {
		return dto.InvoiceView{}
	}
	return dto.InvoiceView{
		ID: inv.ID, OrganizationID: inv.OrganizationID, Number: inv.Number, Status: inv.Status,
		Currency: inv.Currency, TotalMinor: inv.TotalMinor, AmountDueMinor: inv.AmountDueMinor,
		HostedInvoiceURL: inv.HostedInvoiceURL, InvoicePDFURL: inv.InvoicePDFURL,
		PeriodStart: rfcTimePtr(inv.PeriodStart), PeriodEnd: rfcTimePtr(inv.PeriodEnd),
		CreatedAt: rfcTime(inv.CreatedAt), PaidAt: rfcTimePtr(inv.PaidAt),
	}
}
