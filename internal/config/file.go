package config

// File is the on-disk cabinet JSON. Nested datastore keeps today's PascalCase tags.
type File struct {
	HTTP                  FileHTTP  `json:"http"`
	Log                   Log       `json:"log"`
	Datastore             Datastore `json:"datastore"`
	JWT                   FileJWT   `json:"jwt"`
	Billing               Billing   `json:"billing"`
	OAuth                 FileOAuth `json:"oauth"`
	Kafka                 FileKafka `json:"kafka"`
	OTEL                  FileOTEL  `json:"otel"`
	BBR                   BBR       `json:"bbr"`
	EntitlementSigningKey string    `json:"entitlement_signing_key"`
	TermsVersion          string    `json:"terms_version"`
	PrivacyVersion        string    `json:"privacy_version"`
	MinAppVersion         string    `json:"min_app_version"`
	TrialDays             int       `json:"trial_days"`
	Mail                  Mail      `json:"mail"`
}

type FileHTTP struct {
	Addr       string   `json:"addr"`
	AllowedIPs []string `json:"allowed_ips"`
}

type FileJWT struct {
	Secret     string `json:"secret"`
	AccessTTL  string `json:"access_ttl"`
	RefreshTTL string `json:"refresh_ttl"`
}

type FileOAuth struct {
	Google                OAuthApp `json:"google"`
	Facebook              OAuthApp `json:"facebook"`
	X                     OAuthApp `json:"x"`
	RedirectURI           string   `json:"redirect_uri"`
	DeviceVerificationURI string   `json:"device_verification_uri"`
}

type FileKafka struct {
	Enabled  bool     `json:"enabled"`
	Brokers  []string `json:"brokers"`
	ClientID string   `json:"client_id"`
}

type FileOTEL struct {
	Endpoint    string `json:"endpoint"`
	ServiceName string `json:"service_name"`
}
