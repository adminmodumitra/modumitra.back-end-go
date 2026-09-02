package config

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func defaultConfig() Config {
	return Config{
		HTTPAddr:       ":8080",
		OTLPEndpoint:   "localhost:4317",
		ServiceName:    "cabinet-studio-api",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		MinAppVersion:  "1.0.0",
		TrialDays:      14,
		Log: Log{
			Level:  "info",
			Format: "json",
			Sinks:  []string{utils.SinkStderr},
		},
		RequestTimeout:        3 * time.Second,
		BodyLimitBytes:        1 << 20,
		OAuthProviders:        []string{"google", "facebook", "x"},
		JWTSecret:             "cabinet-studio-local-dev-only-change-me",
		AccessTTL:             15 * time.Minute,
		RefreshTTL:            30 * 24 * time.Hour,
		BBR:                   defaultBBR(),
		Datastore:             defaultDatastore(),
		EntitlementSigningKey: "cabinet-studio-entitlement-local-dev-only-change-me",
		Billing: Billing{
			Provider:      BillingAuto,
			WebhookSecret: "cabinet-studio-billing-webhook-local-dev-only",
		},
		KafkaEnabled:         false,
		KafkaClientID:        "cabinet-studio-api",
		OAuthRedirectURI:     "cabinetstudio://oauth/callback",
		OAuthDeviceVerifyURI: "https://app.cabinetstudio.example/device",
		Mail:                 Mail{Driver: "log"},
	}
}

func defaultBBR() BBR {
	return BBR{
		Enabled:      true,
		WindowSize:   DurationAsString(10 * time.Second),
		Buckets:      100,
		CpuThreshold: 800,
	}
}

func Load() Config {
	cfg := defaultConfig()
	cpath := resolveCabinetPath()
	if applyCabinetFile(&cfg, cpath) {
		cfg.DatastorePath = cpath
	} else {
		applyDatastoreJSON(&cfg, resolveDatastorePath())
	}
	overlayDatastoreFile(&cfg)
	applyEnv(&cfg)
	finalize(&cfg)
	return cfg
}

func applyCabinetFile(cfg *Config, path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var f File
	if err := json.Unmarshal(raw, &f); err != nil {
		return false
	}
	applyFile(cfg, f)
	return true
}

func applyDatastoreJSON(cfg *Config, path string) {
	ds := defaultDatastore()
	if raw, err := os.ReadFile(path); err == nil {
		var loaded Datastore
		if err := json.Unmarshal(raw, &loaded); err == nil {
			ds = loaded
		}
	}
	cfg.Datastore = ds
	cfg.DatastorePath = path
	cfg.DatabaseURL = ds.SQL.DSN
}

func overlayDatastoreFile(cfg *Config) {
	path := strings.TrimSpace(os.Getenv("DATASTORE_CONFIG"))
	if path == "" {
		return
	}
	applyDatastoreJSON(cfg, path)
}

func finalize(cfg *Config) {
	cfg.DatabaseURL = cfg.Datastore.SQL.DSN
	cfg.BillingWebhookSecret = cfg.Billing.WebhookSecret
	cfg.BillingProvider = cfg.Billing.resolvedProvider()
	if len(cfg.Log.Sinks) == 0 {
		cfg.Log.Sinks = []string{utils.SinkStderr}
	}
}
