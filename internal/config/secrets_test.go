package config

import (
	"strings"
	"testing"
)

const strongSecret = "k3Q9vZp1Lr8xT2mWc7YbN4sHd6FgJ0aE"

func TestValidateSecrets(t *testing.T) {
	neon := "postgresql://u:p@ep-x-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require"

	t.Run("local dev keeps the shipped defaults", func(t *testing.T) {
		t.Setenv("APP_ENV", "local")
		cfg := defaultConfig()
		cfg.DatabaseURL = "postgres://cabinet:cabinet@localhost:5432/cabinet_studio"
		if err := cfg.ValidateSecrets(); err != nil {
			t.Fatalf("local: %v", err)
		}
	})

	t.Run("hosted database rejects the dev placeholder", func(t *testing.T) {
		t.Setenv("APP_ENV", "local")
		cfg := defaultConfig()
		cfg.DatabaseURL = neon
		err := cfg.ValidateSecrets()
		if err == nil || !strings.Contains(err.Error(), "JWT_SECRET is a dev placeholder") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("production rejects an empty secret", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		cfg := defaultConfig()
		cfg.JWTSecret = ""
		err := cfg.ValidateSecrets()
		if err == nil || !strings.Contains(err.Error(), "JWT_SECRET is empty") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("short entitlement key is rejected", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		cfg := defaultConfig()
		cfg.JWTSecret = strongSecret
		cfg.EntitlementSigningKey = "short-key"
		err := cfg.ValidateSecrets()
		if err == nil || !strings.Contains(err.Error(), "ENTITLEMENT_SIGNING_KEY is too short") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("strong secrets pass on a hosted database", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		cfg := defaultConfig()
		cfg.DatabaseURL = neon
		cfg.JWTSecret = strongSecret
		cfg.EntitlementSigningKey = strongSecret + "-entitlement"
		if err := cfg.ValidateSecrets(); err != nil {
			t.Fatalf("strong: %v", err)
		}
	})

	t.Run("error never echoes the secret", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		cfg := defaultConfig()
		cfg.JWTSecret = "my-short-secret"
		if err := cfg.ValidateSecrets(); err == nil || strings.Contains(err.Error(), "my-short-secret") {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestIsLocalDSN(t *testing.T) {
	cases := map[string]bool{
		"":                                       true,
		"host=localhost user=cabinet":            true,
		"postgres://cabinet:x@postgres:5432/db":  true,
		"postgres://cabinet:x@127.0.0.1:5432/db": true,
		"postgres://u:p@db.example.com:5432/db":  false,
	}
	for dsn, want := range cases {
		if got := isLocalDSN(dsn); got != want {
			t.Errorf("isLocalDSN(%q)=%v want %v", dsn, got, want)
		}
	}
}

func TestValidateSecretsGateKey(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	cfg := defaultConfig()
	cfg.JWTSecret = strongSecret
	cfg.EntitlementSigningKey = strongSecret + "-entitlement"
	cfg.Session = Session{CookieDomain: ".modumitrasoftware.com", Origins: []string{"https://app.modumitrasoftware.com"}}

	if err := cfg.ValidateSecrets(); err == nil || !strings.Contains(err.Error(), "SESSION_GATE_KEY is empty") {
		t.Fatalf("missing gate key: %v", err)
	}
	cfg.Session.GateKey = strongSecret
	if err := cfg.ValidateSecrets(); err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("gate key = jwt secret: %v", err)
	}
	cfg.Session.GateKey = strongSecret + "-gate"
	if err := cfg.ValidateSecrets(); err != nil {
		t.Fatalf("good gate key: %v", err)
	}
}
