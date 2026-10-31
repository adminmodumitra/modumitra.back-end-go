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

func TestEngineManufacturingQuotaEnv(t *testing.T) {
	t.Setenv("ENGINE_MFG_QUOTA_USER_PER_MIN", "10")
	t.Setenv("ENGINE_MFG_QUOTA_ORG_PER_MIN", "-1") // ignored: keeps the default
	t.Setenv("ENGINE_MFG_QUOTA_ORG_PER_DAY", "900")
	e := defaultEngine()
	if e.MfgUserPerMinute != 120 || e.MfgOrgPerMinute != 600 || e.MfgOrgPerDay != 20000 {
		t.Fatalf("defaults %+v", e)
	}
	applyEngineEnv(&e)
	if e.MfgUserPerMinute != 10 || e.MfgOrgPerMinute != 600 || e.MfgOrgPerDay != 900 {
		t.Fatalf("manufacturing quotas %+v", e)
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

func TestEngineDocumentEnv(t *testing.T) {
	e := defaultEngine()
	if e.DocUserPerMinute != 10 || e.DocOrgPerMinute != 60 || e.DocOrgPerDay != 2000 || e.DocTimeout != 30*time.Second {
		t.Fatalf("defaults %+v", e)
	}
	t.Setenv("ENGINE_DOC_QUOTA_USER_PER_MIN", "4")
	t.Setenv("ENGINE_DOC_QUOTA_ORG_PER_MIN", "0") // ignored: keeps the default
	t.Setenv("ENGINE_DOC_QUOTA_ORG_PER_DAY", "50")
	t.Setenv("ENGINE_DOC_TIMEOUT", "45s")
	applyEngineEnv(&e)
	if e.DocUserPerMinute != 4 || e.DocOrgPerMinute != 60 || e.DocOrgPerDay != 50 || e.DocTimeout != 45*time.Second {
		t.Fatalf("env %+v", e)
	}
}

func TestEngineViewQuotaEnv(t *testing.T) {
	t.Setenv("ENGINE_VIEW_QUOTA_USER_PER_MIN", "50")
	t.Setenv("ENGINE_VIEW_QUOTA_ORG_PER_MIN", "0") // ignored: keeps the default
	t.Setenv("ENGINE_VIEW_QUOTA_ORG_PER_DAY", "700")
	e := defaultEngine()
	if e.ViewUserPerMinute != 300 || e.ViewOrgPerMinute != 1500 || e.ViewOrgPerDay != 60000 {
		t.Fatalf("defaults %+v", e)
	}
	applyEngineEnv(&e)
	if e.ViewUserPerMinute != 50 || e.ViewOrgPerMinute != 1500 || e.ViewOrgPerDay != 700 {
		t.Fatalf("view quotas %+v", e)
	}
}
