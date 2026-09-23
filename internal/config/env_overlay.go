package config

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

func applyEnv(cfg *Config) {
	cfg.HTTPAddr = getenv("HTTP_ADDR", cfg.HTTPAddr)
	if v := strings.TrimSpace(os.Getenv("ALLOWED_IPS")); v != "" {
		var ips []string
		if err := json.Unmarshal([]byte(v), &ips); err == nil && ips != nil {
			cfg.AllowedIPs = ips
		}
	}
	cfg.OTLPEndpoint = getenv("OTEL_EXPORTER_OTLP_ENDPOINT", cfg.OTLPEndpoint)
	cfg.ServiceName = getenv("OTEL_SERVICE_NAME", cfg.ServiceName)
	cfg.TermsVersion = getenv("TERMS_VERSION", cfg.TermsVersion)
	cfg.PrivacyVersion = getenv("PRIVACY_VERSION", cfg.PrivacyVersion)
	cfg.MinAppVersion = getenv("MIN_APP_VERSION", cfg.MinAppVersion)
	cfg.TrialDays = getenvInt("TRIAL_DAYS", cfg.TrialDays)
	applyLogEnv(cfg)
	cfg.JWTSecret = getenv("JWT_SECRET", cfg.JWTSecret)
	cfg.AccessTTL = getenvDuration("JWT_ACCESS_TTL", cfg.AccessTTL)
	cfg.RefreshTTL = getenvDuration("JWT_REFRESH_TTL", cfg.RefreshTTL)
	applyBBREnv(cfg)
	applyDatastoreEnv(&cfg.Datastore)
	cfg.EntitlementSigningKey = getenv("ENTITLEMENT_SIGNING_KEY", cfg.EntitlementSigningKey)
	applyBillingEnv(cfg)
	applyKafkaEnv(cfg)
	applyOAuthEnv(cfg)
	cfg.Mail.Driver = getenv("MAIL_DRIVER", cfg.Mail.Driver)
}

func applyBillingEnv(cfg *Config) {
	cfg.Billing.Provider = getenv("BILLING_PROVIDER", cfg.Billing.Provider)
	cfg.Billing.WebhookSecret = getenv("BILLING_WEBHOOK_SECRET", cfg.Billing.WebhookSecret)
	cfg.Billing.Prices.StudioMonthly = getenv("BILLING_PRICE_STUDIO_MONTHLY", cfg.Billing.Prices.StudioMonthly)
	cfg.Billing.Prices.TeamMonthly = getenv("BILLING_PRICE_TEAM_MONTHLY", cfg.Billing.Prices.TeamMonthly)
}

func applyKafkaEnv(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("KAFKA_ENABLED")); v != "" {
		cfg.KafkaEnabled = getenvBool("KAFKA_ENABLED", cfg.KafkaEnabled)
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")); v != "" {
		cfg.KafkaBrokers = splitCSV(v)
	}
	cfg.KafkaClientID = getenv("KAFKA_CLIENT_ID", cfg.KafkaClientID)
}

func applyBBREnv(cfg *Config) {
	cfg.BBR.Enabled = getenvBool("BBR_ENABLED", cfg.BBR.Enabled)
	if v := strings.TrimSpace(os.Getenv("BBR_WINDOW_SIZE")); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			cfg.BBR.WindowSize = DurationAsString(parsed)
		}
	}
	cfg.BBR.Buckets = getenvInt("BBR_BUCKETS", cfg.BBR.Buckets)
	cfg.BBR.CpuThreshold = getenvInt64("BBR_CPU_THRESHOLD", cfg.BBR.CpuThreshold)
}

func applyOAuthEnv(cfg *Config) {
	cfg.OAuthGoogle.ClientID = getenv("OAUTH_GOOGLE_CLIENT_ID", cfg.OAuthGoogle.ClientID)
	cfg.OAuthGoogle.ClientSecret = getenv("OAUTH_GOOGLE_CLIENT_SECRET", cfg.OAuthGoogle.ClientSecret)
	cfg.OAuthFacebook.ClientID = getenv("OAUTH_FACEBOOK_CLIENT_ID", cfg.OAuthFacebook.ClientID)
	cfg.OAuthFacebook.ClientSecret = getenv("OAUTH_FACEBOOK_CLIENT_SECRET", cfg.OAuthFacebook.ClientSecret)
	cfg.OAuthX.ClientID = getenv("OAUTH_X_CLIENT_ID", cfg.OAuthX.ClientID)
	cfg.OAuthX.ClientSecret = getenv("OAUTH_X_CLIENT_SECRET", cfg.OAuthX.ClientSecret)
	cfg.OAuthRedirectURI = getenv("OAUTH_REDIRECT_URI", cfg.OAuthRedirectURI)
	cfg.OAuthDeviceVerifyURI = getenv("OAUTH_DEVICE_VERIFICATION_URI", cfg.OAuthDeviceVerifyURI)
}
