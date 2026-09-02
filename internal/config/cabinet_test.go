package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cabinetPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, "configs", "cabinet.dummy.json")
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("configs/cabinet.dummy.json not found")
	return ""
}

func TestLoadCabinetDummyFile(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", cabinetPath(t))

	cfg := Load()
	if cfg.Datastore.Driver != "postgres" {
		t.Fatalf("Driver=%q", cfg.Datastore.Driver)
	}
	if !strings.Contains(cfg.Datastore.SQL.DSN, "localhost:5432") {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
	if cfg.Billing.Provider != BillingAuto {
		t.Fatalf("provider=%q", cfg.Billing.Provider)
	}
	if cfg.Billing.PriceStudio() != "price_studio_monthly_placeholder" {
		t.Fatalf("studio=%q", cfg.Billing.PriceStudio())
	}
	if cfg.Billing.PriceTeam() != "price_team_monthly_placeholder" {
		t.Fatalf("team=%q", cfg.Billing.PriceTeam())
	}
	if cfg.Billing.Live() {
		t.Fatal("placeholder prices must not enable live billing")
	}
	if cfg.BillingProvider != BillingFake {
		t.Fatalf("resolved provider=%q", cfg.BillingProvider)
	}
	if cfg.KafkaEnabled {
		t.Fatal("kafka should be disabled in dummy")
	}
	if len(cfg.KafkaBrokers) != 0 {
		t.Fatalf("brokers=%v", cfg.KafkaBrokers)
	}
	if cfg.Mail.Driver != "log" {
		t.Fatalf("mail=%q", cfg.Mail.Driver)
	}
	if cfg.OAuthGoogle.Enabled() || cfg.OAuthFacebook.Enabled() || cfg.OAuthX.Enabled() {
		t.Fatal("oauth must be disabled with empty clients")
	}
	if err := cfg.Mail.Validate(); err != nil {
		t.Fatalf("mail validate: %v", err)
	}
}

func TestBillingLivePlaceholderPrices(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", cabinetPath(t))
	cfg := Load()
	if cfg.Billing.Live() {
		t.Fatal("dummy placeholder prices: Live() want false")
	}
}

func TestBillingLiveRealPrices(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", cabinetPath(t))
	t.Setenv("BILLING_WEBHOOK_SECRET", "whsec_live_test")
	t.Setenv("BILLING_PRICE_STUDIO_MONTHLY", "price_1Abc")
	t.Setenv("BILLING_PRICE_TEAM_MONTHLY", "price_1Def")
	cfg := Load()
	if !cfg.Billing.Live() {
		t.Fatal("real-looking prices + webhook: Live() want true")
	}
	if cfg.Billing.PriceStudio() != "price_1Abc" || cfg.Billing.PriceTeam() != "price_1Def" {
		t.Fatalf("prices studio=%q team=%q", cfg.Billing.PriceStudio(), cfg.Billing.PriceTeam())
	}
	if cfg.BillingProvider != BillingStripe {
		t.Fatalf("resolved=%q", cfg.BillingProvider)
	}
}

func TestBillingLiveTypeRules(t *testing.T) {
	placeholder := Billing{
		Provider:      BillingAuto,
		WebhookSecret: "whsec_x",
		Prices:        BillingPrices{StudioMonthly: "price_studio_monthly_placeholder", TeamMonthly: "price_team_monthly_placeholder"},
	}
	if placeholder.Live() {
		t.Fatal("placeholder Live() true")
	}
	live := Billing{
		Provider:      BillingAuto,
		WebhookSecret: "whsec_x",
		Prices:        BillingPrices{StudioMonthly: "price_1Abc", TeamMonthly: "price_1Def"},
	}
	if !live.Live() {
		t.Fatal("real prices Live() false")
	}
	forcedFake := live
	forcedFake.Provider = BillingFake
	if forcedFake.Live() {
		t.Fatal("explicit fake must not be live")
	}
	forcedStripe := placeholder
	forcedStripe.Provider = BillingStripe
	if !forcedStripe.Live() {
		t.Fatal("explicit stripe must be live")
	}
	missingSecret := live
	missingSecret.WebhookSecret = ""
	if missingSecret.Live() {
		t.Fatal("auto without webhook must not be live")
	}
}

func TestOAuthPlaceholderDisabled(t *testing.T) {
	a := OAuthApp{ClientID: "google-client-placeholder", ClientSecret: "secret-placeholder"}
	if a.Enabled() {
		t.Fatal("placeholder oauth must be disabled")
	}
	a = OAuthApp{ClientID: "real-id", ClientSecret: "real-secret"}
	if !a.Enabled() {
		t.Fatal("real oauth must be enabled")
	}
}

func TestCabinetMissingUsesDefaults(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", filepath.Join(t.TempDir(), "missing.json"))
	cfg := Load()
	if cfg.Datastore.Driver != "postgres" {
		t.Fatalf("Driver=%q", cfg.Datastore.Driver)
	}
	if !strings.Contains(cfg.Datastore.SQL.DSN, "localhost:5432") {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
}

func TestMailSMTPReserved(t *testing.T) {
	err := (Mail{Driver: "smtp"}).Validate()
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("smtp validate=%v", err)
	}
}
