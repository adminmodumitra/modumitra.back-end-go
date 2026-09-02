package billing

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (s *Service) ensurePaidInvoice(ctx context.Context, obj map[string]any, acc *BillingAccount, sub *Subscription, plan *Plan, now, periodEnd time.Time) error {
	provInv := asString(obj, "provider_invoice_id")
	if provInv == "" {
		provInv = PrefixProvInv + asString(obj, "id")
	}
	if _, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); err == nil || !errors.Is(err, ErrNotFound) {
		return mapRepoErr(err)
	}
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
		ID: s.ids.Invoice(), BillingAccountID: acc.ID, OrganizationID: accOrg(acc),
		SubscriptionID: strPtr(sub.ID), Provider: s.provider.Name(), ProviderInvoiceID: provInv,
		Number: strPtr(num), Status: InvoicePaid, Currency: currency, TotalMinor: amount, AmountDueMinor: 0,
		HostedInvoiceURL: strPtr(hosted), PeriodStart: timePtr(now), PeriodEnd: timePtr(periodEnd),
		PaidAt: timePtr(now), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Service) applyInvoicePaid(ctx context.Context, obj map[string]any, now time.Time) (string, string, error) {
	provInv := asString(obj, "id", "provider_invoice_id")
	if provInv == "" {
		return "", "", utils.E(utils.CodeInvalidPayload)
	}
	if existing, err := s.invoices.GetByProviderInvoiceID(ctx, provInv); err == nil {
		return existing.OrganizationID, deref(existing.SubscriptionID), nil
	} else if !errors.Is(err, ErrNotFound) {
		return "", "", mapRepoErr(err)
	}
	acc, err := s.accountFromEvent(ctx, obj)
	if err != nil {
		return "", "", err
	}
	var localSubID *string
	if sub, err := s.subs.GetByAccountID(ctx, acc.ID); err == nil {
		localSubID = strPtr(sub.ID)
	}
	amount := asInt(obj, "amount_paid", "amount_total", "total_minor")
	currency := asString(obj, "currency")
	if currency == "" {
		currency = "usd"
	}
	inv := &Invoice{
		ID: s.ids.Invoice(), BillingAccountID: acc.ID, OrganizationID: accOrg(acc), SubscriptionID: localSubID,
		Provider: s.provider.Name(), ProviderInvoiceID: provInv, Status: InvoicePaid, Currency: currency,
		TotalMinor: amount, AmountDueMinor: 0, HostedInvoiceURL: strPtr(asString(obj, "hosted_invoice_url")),
		InvoicePDFURL: strPtr(asString(obj, "invoice_pdf", "invoice_pdf_url")),
		PaidAt:        timePtr(now), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.invoices.Insert(ctx, inv); err != nil && !errors.Is(err, ErrDuplicate) {
		return "", "", mapRepoErr(err)
	}
	return accOrg(acc), deref(localSubID), nil
}
