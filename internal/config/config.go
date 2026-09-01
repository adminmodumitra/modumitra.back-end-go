package config

import (
	"encoding/json"
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
	BBR            BBR
}

// DurationAsString unmarshals JSON durations like "10s".
type DurationAsString time.Duration

func (d DurationAsString) Duration() time.Duration { return time.Duration(d) }

func (d DurationAsString) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *DurationAsString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = DurationAsString(parsed)
	return nil
}

// BBR is Kratos aegis BBR load-shedding config (JSON tags match the product shape).
type BBR struct {
	Enabled      bool             `json:"Enabled"`
	WindowSize   DurationAsString `json:"WindowSize"`
	Buckets      int              `json:"Buckets"`
	CpuThreshold int64            `json:"CpuThreshold"`
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
		BBR:            loadBBR(),
	}
}

func loadBBR() BBR {
	window := 10 * time.Second
	if v := strings.TrimSpace(os.Getenv("BBR_WINDOW_SIZE")); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			window = parsed
		}
	}
	return BBR{
		Enabled:      getenvBool("BBR_ENABLED", true),
		WindowSize:   DurationAsString(window),
		Buckets:      getenvInt("BBR_BUCKETS", 100),
		CpuThreshold: getenvInt64("BBR_CPU_THRESHOLD", 800),
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

func getenvInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
