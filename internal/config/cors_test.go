package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCORSOrigins(t *testing.T) {
	clearDatastoreEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cabinet.json")
	body := `{"http":{"addr":":8080","cors_origins":["https://modumitrasoftware.com"]}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CABINET_CONFIG", path)
	cfg := Load()
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "https://modumitrasoftware.com" {
		t.Fatalf("CORSOrigins=%v", cfg.CORSOrigins)
	}
	t.Setenv("CORS_ORIGINS", `["http://localhost:1420","https://www.modumitrasoftware.com"]`)
	cfg = Load()
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[1] != "https://www.modumitrasoftware.com" {
		t.Fatalf("env CORSOrigins=%v", cfg.CORSOrigins)
	}
}

func TestProductionConfigListsSiteOrigin(t *testing.T) {
	clearDatastoreEnv(t)
	t.Setenv("CABINET_CONFIG", filepath.Join("..", "..", "configs", "cabinet.production.json"))
	t.Setenv("JWT_SECRET", "test-secret-for-production-config")
	cfg := Load()
	for _, o := range cfg.CORSOrigins {
		if o == "https://modumitrasoftware.com" {
			return
		}
	}
	t.Fatalf("CORSOrigins=%v", cfg.CORSOrigins)
}
