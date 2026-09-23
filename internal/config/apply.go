package config

import "time"

func applyFile(cfg *Config, f File) {
	if f.HTTP.Addr != "" {
		cfg.HTTPAddr = f.HTTP.Addr
	}
	if f.HTTP.AllowedIPs != nil {
		cfg.AllowedIPs = append([]string{}, f.HTTP.AllowedIPs...)
	}
	applyFileLog(cfg, f.Log)
	if f.Datastore.Driver != "" || f.Datastore.SQL.DSN != "" {
		cfg.Datastore = f.Datastore
		cfg.DatabaseURL = f.Datastore.SQL.DSN
	}
	applyFileJWT(cfg, f.JWT)
	applyFileBilling(cfg, f.Billing)
	applyFileOAuth(cfg, f.OAuth)
	applyFileKafka(cfg, f.Kafka)
	if f.OTEL.Endpoint != "" {
		cfg.OTLPEndpoint = f.OTEL.Endpoint
	}
	if f.OTEL.ServiceName != "" {
		cfg.ServiceName = f.OTEL.ServiceName
	}
	if f.BBR.Buckets != 0 || f.BBR.WindowSize != 0 {
		cfg.BBR = f.BBR
	}
	cfg.EntitlementSigningKey = f.EntitlementSigningKey
	if f.TermsVersion != "" {
		cfg.TermsVersion = f.TermsVersion
	}
	if f.PrivacyVersion != "" {
		cfg.PrivacyVersion = f.PrivacyVersion
	}
	if f.MinAppVersion != "" {
		cfg.MinAppVersion = f.MinAppVersion
	}
	if f.TrialDays > 0 {
		cfg.TrialDays = f.TrialDays
	}
	if f.Mail.Driver != "" {
		cfg.Mail.Driver = f.Mail.Driver
	}
}

func applyFileJWT(cfg *Config, j FileJWT) {
	cfg.JWTSecret = j.Secret
	if d, err := time.ParseDuration(j.AccessTTL); err == nil && d > 0 {
		cfg.AccessTTL = d
	}
	if d, err := time.ParseDuration(j.RefreshTTL); err == nil && d > 0 {
		cfg.RefreshTTL = d
	}
}

func applyFileBilling(cfg *Config, b Billing) {
	if b.Provider != "" {
		cfg.Billing.Provider = b.Provider
	}
	cfg.Billing.WebhookSecret = b.WebhookSecret
	cfg.Billing.Prices.StudioMonthly = b.Prices.StudioMonthly
	cfg.Billing.Prices.TeamMonthly = b.Prices.TeamMonthly
}

func applyFileKafka(cfg *Config, k FileKafka) {
	cfg.KafkaEnabled = k.Enabled
	if k.Brokers != nil {
		cfg.KafkaBrokers = append([]string{}, k.Brokers...)
	}
	if k.ClientID != "" {
		cfg.KafkaClientID = k.ClientID
	}
}

func applyFileOAuth(cfg *Config, o FileOAuth) {
	cfg.OAuthGoogle = overlayOAuth(cfg.OAuthGoogle, o.Google)
	cfg.OAuthFacebook = overlayOAuth(cfg.OAuthFacebook, o.Facebook)
	cfg.OAuthX = overlayOAuth(cfg.OAuthX, o.X)
	if o.RedirectURI != "" {
		cfg.OAuthRedirectURI = o.RedirectURI
	}
	if o.DeviceVerificationURI != "" {
		cfg.OAuthDeviceVerifyURI = o.DeviceVerificationURI
	}
}

func overlayOAuth(dst, src OAuthApp) OAuthApp {
	if src.ClientID != "" {
		dst.ClientID = src.ClientID
	}
	if src.ClientSecret != "" {
		dst.ClientSecret = src.ClientSecret
	}
	return dst
}
