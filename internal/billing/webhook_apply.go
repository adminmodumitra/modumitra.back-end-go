package billing

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) applyCheckoutCompleted(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	orgID := asString(obj, "organization_id", "client_reference_id")
	planID := asString(obj, "plan_id")
	customer := asString(obj, "customer")
	provSub := asString(obj, "subscription")
	if orgID == "" || planID == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return "", "", mapRepoErr(err)
	}
	if err := s.orgs.SetProviderCustomer(ctx, orgID, s.provider.Name(), customer, now); err != nil {
		return "", "", mapRepoErr(err)
	}
	periodEnd := now.AddDate(0, 1, 0)
	sub, err := s.subs.GetByOrg(ctx, orgID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	if sub == nil {
		sub = &Subscription{
			ID: s.ids.Subscription(), OrganizationID: orgID, CreatedAt: now,
		}
	}
	sub.PlanID = plan.ID
	sub.PlanCode = plan.Code
	sub.Status = StatusActive
	sub.Provider = s.provider.Name()
	sub.ProviderSubscriptionID = provSub
	sub.SeatsIncluded = plan.SeatsIncluded
	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = periodEnd
	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = now
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	if _, err := s.subs.GetByID(ctx, sub.ID); err != nil {
		if err := s.subs.Insert(ctx, sub); err != nil && !errors.Is(err, ErrDuplicate) {
			return "", "", mapRepoErr(err)
		}
	} else {
		if err := s.subs.Update(ctx, sub); err != nil {
			return "", "", mapRepoErr(err)
		}
	}
	provInv := asString(obj, "provider_invoice_id")
	if provInv == "" {
		provInv = PrefixProvInv + asString(obj, "id")
	}
	if _, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); errors.Is(err, ErrNotFound) {
		hosted := asString(obj, "hosted_invoice_url")
		if hosted == "" {
			hosted = "https://billing.fake.local/invoice/" + provInv
		}
		amount := asInt(obj, "amount_total")
		if amount == 0 {
			amount = plan.UnitAmountMinor
		}
		currency := asString(obj, "currency")
		if currency == "" {
			currency = plan.Currency
		}
		num := "INV-" + now.UTC().Format("20060102") + "-" + sub.ID[len(sub.ID)-4:]
		inv := &Invoice{
			ID: s.ids.Invoice(), OrganizationID: orgID, SubscriptionID: strPtr(sub.ID),
			Provider: s.provider.Name(), ProviderInvoiceID: provInv, Number: strPtr(num),
			Status: InvoicePaid, Currency: currency, TotalMinor: amount, AmountDueMinor: 0,
			HostedInvoiceURL: strPtr(hosted), PeriodStart: timePtr(now), PeriodEnd: timePtr(periodEnd),
			PaidAt: timePtr(now), CreatedAt: now, UpdatedAt: now,
		}
		if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
			return "", "", mapRepoErr(err)
		}
	} else if err != nil {
		return "", "", mapRepoErr(err)
	}
	return orgID, sub.ID, nil
}

func (s *Service) applyInvoicePaid(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	orgID := asString(obj, "organization_id")
	provInv := asString(obj, "id", "provider_invoice_id")
	if provInv == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	if existing, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); err == nil {
		return existing.OrganizationID, deref(existing.SubscriptionID), nil
	} else if !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	subID := asString(obj, "subscription_id", "subscription")
	var localSubID *string
	if orgID == "" && subID != "" {
		if sub, err := s.subs.GetByOrg(ctx, orgID); err == nil {
			orgID = sub.OrganizationID
			localSubID = strPtr(sub.ID)
		}
	}
	if orgID == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	if sub, err := s.subs.GetByOrg(ctx, orgID); err == nil {
		localSubID = strPtr(sub.ID)
	}
	amount := asInt(obj, "amount_paid", "amount_total", "total_minor")
	currency := asString(obj, "currency")
	if currency == "" {
		currency = "usd"
	}
	hosted := asString(obj, "hosted_invoice_url")
	pdf := asString(obj, "invoice_pdf", "invoice_pdf_url")
	inv := &Invoice{
		ID: s.ids.Invoice(), OrganizationID: orgID, SubscriptionID: localSubID,
		Provider: s.provider.Name(), ProviderInvoiceID: provInv,
		Status: InvoicePaid, Currency: currency, TotalMinor: amount, AmountDueMinor: 0,
		HostedInvoiceURL: strPtr(hosted), InvoicePDFURL: strPtr(pdf),
		PaidAt: timePtr(now), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
		return "", "", mapRepoErr(err)
	}
	return orgID, deref(localSubID), nil
}
