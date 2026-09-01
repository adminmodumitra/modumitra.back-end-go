package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	OTLPEndpoint   string
	ServiceName    string
	TermsVersion   string
	PrivacyVersion string
	MinAppVersion  string
	TrialDays      int
	LogLevel       string
	RequestTimeout time.Duration
	BodyLimitBytes int64
	OAuthProviders []string
}

func Load() Config {
	return Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:    getenv("DATABASE_URL", "postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable"),
		OTLPEndpoint:   getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:    getenv("OTEL_SERVICE_NAME", "cabinet-studio-api"),
		TermsVersion:   getenv("TERMS_VERSION", "2026-06-01"),
		PrivacyVersion: getenv("PRIVACY_VERSION", "2026-06-01"),
		MinAppVersion:  getenv("MIN_APP_VERSION", "1.0.0"),
		TrialDays:      getenvInt("TRIAL_DAYS", 14),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		RequestTimeout: 3 * time.Second,
		BodyLimitBytes: 1 << 20,
		OAuthProviders: []string{"google", "facebook", "x"},
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
