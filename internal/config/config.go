package config

import (
	"encoding/json"
	"time"
)

type Config struct {
	HTTPAddr              string
	AllowedIPs            []string
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
	Billing               Billing
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
	Mail                  Mail
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
