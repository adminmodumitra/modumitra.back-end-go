package config

import (
	"strings"
	"testing"
)

func TestCabinetFileForEnv(t *testing.T) {
	cases := map[string]string{
		"":            defaultCabinetPath,
		"local":       defaultCabinetPath,
		"LOCAL":       defaultCabinetPath,
		"dev":         "configs/cabinet.dev.json",
		"development": "configs/cabinet.dev.json",
		"production":  "configs/cabinet.production.json",
		"prod":        "configs/cabinet.production.json",
		"staging":     defaultCabinetPath,
	}
	for in, want := range cases {
		if got := cabinetFileForEnv(in); got != want {
			t.Fatalf("%q -> %q want %q", in, got, want)
		}
	}
}

func TestResolveCabinetPathCABINETConfigWins(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("CABINET_CONFIG", "/explicit/cabinet.json")
	if got := resolveCabinetPath(); got != "/explicit/cabinet.json" {
		t.Fatalf("path=%q", got)
	}
}

func TestLoadAppEnvLocal(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("APP_ENV", "local")
	cfg := Load()
	if !strings.Contains(cfg.Datastore.SQL.DSN, "localhost:5432") {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
	if cfg.KafkaEnabled {
		t.Fatal("kafka should be off")
	}
	if cfg.Billing.Live() {
		t.Fatal("Live() want false")
	}
}

func TestLoadAppEnvDev(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("APP_ENV", "dev")
	cfg := Load()
	if !cfg.KafkaEnabled {
		t.Fatal("kafka should be enabled")
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "kafka:9092" {
		t.Fatalf("brokers=%v", cfg.KafkaBrokers)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("level=%q", cfg.Log.Level)
	}
	if cfg.Billing.Live() {
		t.Fatal("Live() want false")
	}
}

func TestLoadAppEnvProduction(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("APP_ENV", "production")
	cfg := Load()
	if cfg.JWTSecret != "" {
		t.Fatalf("jwt=%q", cfg.JWTSecret)
	}
	if cfg.Billing.PriceStudio() != "" || cfg.Billing.PriceTeam() != "" {
		t.Fatalf("prices studio=%q team=%q", cfg.Billing.PriceStudio(), cfg.Billing.PriceTeam())
	}
	if cfg.Billing.Live() {
		t.Fatal("Live() want false")
	}
	if !cfg.KafkaEnabled {
		t.Fatal("kafka should be enabled")
	}
	if len(cfg.KafkaBrokers) != 0 {
		t.Fatalf("brokers=%v", cfg.KafkaBrokers)
	}
}

func TestCabinetConfigOverridesAppEnv(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("CABINET_CONFIG", cabinetPath(t))
	cfg := Load()
	if !strings.Contains(cfg.Datastore.SQL.DSN, "localhost:5432") {
		t.Fatalf("SQL.DSN=%q", cfg.Datastore.SQL.DSN)
	}
	if cfg.KafkaEnabled {
		t.Fatal("explicit CABINET_CONFIG dummy should keep kafka off")
	}
	if cfg.JWTSecret == "" {
		t.Fatal("dummy jwt secret should remain")
	}
}
