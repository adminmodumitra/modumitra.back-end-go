package config

import "strings"

const (
	BillingAuto   = "auto"
	BillingFake   = "fake"
	BillingStripe = "stripe"
)

type BillingPrices struct {
	StudioMonthly string `json:"studio_monthly"`
	TeamMonthly   string `json:"team_monthly"`
}

type Billing struct {
	Provider      string        `json:"provider"`
	WebhookSecret string        `json:"webhook_secret"`
	Prices        BillingPrices `json:"prices"`
}

func (b Billing) PriceStudio() string {
	return strings.TrimSpace(b.Prices.StudioMonthly)
}

func (b Billing) PriceTeam() string {
	return strings.TrimSpace(b.Prices.TeamMonthly)
}

// Live is true when Stripe should be selected. provider auto (default) is fake
// unless webhook secret is set and both price ids are real (non-empty, no
// "placeholder"). Explicit stripe/fake win.
func (b Billing) Live() bool {
	switch strings.ToLower(strings.TrimSpace(b.Provider)) {
	case BillingStripe:
		return true
	case BillingFake:
		return false
	default:
		return strings.TrimSpace(b.WebhookSecret) != "" && liveValue(b.PriceStudio()) && liveValue(b.PriceTeam())
	}
}

func (b Billing) resolvedProvider() string {
	if b.Live() {
		return BillingStripe
	}
	return BillingFake
}
