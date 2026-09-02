package sso

import "time"

const (
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
	ProviderX        = "x"

	IntentLogin = "login"
	IntentLink  = "link"

	StatusPending    = "pending"
	StatusAuthorized = "authorized"
	StatusDenied     = "denied"
	StatusConsumed   = "consumed"
	StatusExpired    = "expired"

	PrefixIdentity = "oid_"
	PrefixState    = "st_"
	PrefixDevice   = "dc_"

	DesktopClientID = "cabinet-studio-desktop"

	DefaultStateTTL       = 10 * time.Minute
	DefaultDeviceTTL      = 10 * time.Minute
	DefaultDeviceInterval = 5
)

// Profile is the provider userinfo after code exchange. Never includes tokens.
type Profile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	DisplayName    string
}

type ProviderInfo struct {
	ID                  string
	Label               string
	DeviceCodeSupported bool
}

type Identity struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	Email          string
	EmailVerified  bool
	LinkedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type State struct {
	ID               string
	State            string
	Provider         string
	Intent           string
	RedirectURI      string
	CodeChallenge    string
	CodeVerifierHash string
	UserID           string
	ConsumedAt       *time.Time
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type DeviceCode struct {
	ID              string
	DeviceCode      string
	UserCode        string
	ClientID        string
	Provider        string
	Status          string
	UserID          string
	IntervalSeconds int
	LastPollAt      *time.Time
	AuthorizedAt    *time.Time
	ConsumedAt      *time.Time
	ExpiresAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StartInput struct {
	Provider    string
	RedirectURI string
	Intent      string
	ActorID     string
}

type StartResult struct {
	AuthorizationURL string
	State            string
}

type LoginResult struct {
	User           *UserView
	Pair           TokenPair
	LinkedProvider string
}

// UserView is the identity.User fields SSO returns without importing HTTP DTOs.
type UserView struct {
	ID            string
	Email         string
	EmailVerified bool
	DisplayName   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PasswordHash  string // never serialized; used internally for last-method checks
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
}

type DeviceCodeResult struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
}

type Settings struct {
	TermsVersion    string
	PrivacyVersion  string
	StateTTL        time.Duration
	DeviceTTL       time.Duration
	DeviceInterval  int
	DeviceClientID  string
	VerificationURI string
	PKCESecret      string
}

func KnownProvider(id string) bool {
	switch id {
	case ProviderGoogle, ProviderFacebook, ProviderX:
		return true
	default:
		return false
	}
}

func ProviderLabel(id string) string {
	switch id {
	case ProviderGoogle:
		return "Google"
	case ProviderFacebook:
		return "Facebook"
	case ProviderX:
		return "X"
	default:
		return id
	}
}

func DeviceCodeSupported(id string) bool {
	return id == ProviderGoogle
}
