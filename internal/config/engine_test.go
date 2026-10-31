package config

import (
	"strings"
	"testing"
	"time"
)

func TestEngineEnv(t *testing.T) {
	t.Setenv("ENGINE_URL", "http://cabinet-engine:10000/")
	t.Setenv("ENGINE_SHARED_SECRET", strongSecret)
	t.Setenv("ENGINE_TIMEOUT", "5s")
	t.Setenv("ENGINE_QUOTA_USER_PER_MIN", "30")
	t.Setenv("ENGINE_QUOTA_ORG_PER_MIN", "0") // ignored: keeps the default
	t.Setenv("ENGINE_QUOTA_ORG_PER_DAY", "100")
	e := defaultEngine()
	applyEngineEnv(&e)
	if e.URL != "http://cabinet-engine:10000" || e.Timeout != 5*time.Second {
		t.Fatalf("%+v", e)
	}
	if e.UserPerMinute != 30 || e.OrgPerMinute != 300 || e.OrgPerDay != 100 {
		t.Fatalf("quotas %+v", e)
	}
}

func TestEngineDefaults(t *testing.T) {
	e := defaultEngine()
	if e.Enabled() || e.Timeout != 8*time.Second || e.UserPerMinute != 60 || e.OrgPerMinute != 300 || e.OrgPerDay != 5000 {
		t.Fatalf("%+v", e)
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("disabled engine must validate: %v", err)
	}
}

func TestEngineSecretChecked(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	cases := map[string]string{
		"":                                       "ENGINE_SHARED_SECRET is empty",
		"short":                                  "ENGINE_SHARED_SECRET is too short",
		"engine-local-dev-only-change-me-000000": "ENGINE_SHARED_SECRET is a dev placeholder",
	}
	for secret, want := range cases {
		cfg := defaultConfig()
		cfg.Engine.URL = "http://localhost:10000"
		cfg.Engine.SharedSecret = secret
		err := cfg.ValidateSecrets()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("secret %q: err=%v", secret, err)
		}
		if secret != "" && strings.Contains(err.Error(), secret) {
			t.Fatalf("error echoes the secret: %v", err)
		}
	}
	cfg := defaultConfig()
	cfg.Engine.URL = "http://localhost:10000"
	cfg.Engine.SharedSecret = strongSecret
	if err := cfg.ValidateSecrets(); err != nil {
		t.Fatalf("strong: %v", err)
	}
	cfg.Engine.URL = "cabinet-engine:10000"
	if err := cfg.ValidateSecrets(); err == nil || !strings.Contains(err.Error(), "ENGINE_URL") {
		t.Fatalf("bad url: %v", err)
	}
}
