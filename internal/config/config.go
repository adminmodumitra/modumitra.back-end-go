package config

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr              string
	DatabaseURL           string
	OTLPEndpoint          string
	ServiceName           string
	TermsVersion          string
	PrivacyVersion        string
	MinAppVersion         string
	TrialDays             int
	Log                   Log
	RequestTimeout        time.Duration
	BodyLimitBytes        int64
	OAuthProviders        []string
	JWTSecret             string
	AccessTTL             time.Duration
	RefreshTTL            time.Duration
	BBR                   BBR
	Datastore             Datastore
	DatastorePath         string
	EntitlementSigningKey string
	BillingProvider       string
	BillingWebhookSecret  string
	KafkaEnabled          bool
	KafkaBrokers          []string
	KafkaClientID         string
	OAuthGoogle           OAuthApp
	OAuthFacebook         OAuthApp
	OAuthX                OAuthApp
	OAuthRedirectURI      string
	OAuthDeviceVerifyURI  string
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
	path := resolveDatastorePath()
	ds := defaultDatastore()
	if raw, err := os.ReadFile(path); err == nil {
		var loaded Datastore
		if err := json.Unmarshal(raw, &loaded); err == nil {
			ds = loaded
		}
	}
	applyDatastoreEnv(&ds)

	return Config{
		HTTPAddr:              getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:           ds.SQL.DSN,
		OTLPEndpoint:          getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:           getenv("OTEL_SERVICE_NAME", "cabinet-studio-api"),
		TermsVersion:          getenv("TERMS_VERSION", "2026-06-01"),
		PrivacyVersion:        getenv("PRIVACY_VERSION", "2026-06-01"),
		MinAppVersion:         getenv("MIN_APP_VERSION", "1.0.0"),
		TrialDays:             getenvInt("TRIAL_DAYS", 14),
		Log:                   loadLog(),
		RequestTimeout:        3 * time.Second,
		BodyLimitBytes:        1 << 20,
		OAuthProviders:        []string{"google", "facebook", "x"},
		JWTSecret:             getenv("JWT_SECRET", "cabinet-studio-local-dev-only-change-me"),
		AccessTTL:             getenvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:            getenvDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
		BBR:                   loadBBR(),
		Datastore:             ds,
		DatastorePath:         path,
		EntitlementSigningKey: getenv("ENTITLEMENT_SIGNING_KEY", "cabinet-studio-entitlement-local-dev-only-change-me"),
		BillingProvider:       getenv("BILLING_PROVIDER", "fake"),
		BillingWebhookSecret:  getenv("BILLING_WEBHOOK_SECRET", "cabinet-studio-billing-webhook-local-dev-only"),
		KafkaEnabled:          getenvBool("KAFKA_ENABLED", false),
		KafkaBrokers:          splitCSV(os.Getenv("KAFKA_BROKERS")),
		KafkaClientID:         getenv("KAFKA_CLIENT_ID", "cabinet-studio-api"),
		OAuthGoogle: OAuthApp{
			ClientID:     getenv("OAUTH_GOOGLE_CLIENT_ID", ""),
			ClientSecret: getenv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
		},
		OAuthFacebook: OAuthApp{
			ClientID:     getenv("OAUTH_FACEBOOK_CLIENT_ID", ""),
			ClientSecret: getenv("OAUTH_FACEBOOK_CLIENT_SECRET", ""),
		},
		OAuthX: OAuthApp{
			ClientID:     getenv("OAUTH_X_CLIENT_ID", ""),
			ClientSecret: getenv("OAUTH_X_CLIENT_SECRET", ""),
		},
		OAuthRedirectURI:     getenv("OAUTH_REDIRECT_URI", "cabinetstudio://oauth/callback"),
		OAuthDeviceVerifyURI: getenv("OAUTH_DEVICE_VERIFICATION_URI", "https://app.cabinetstudio.example/device"),
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
