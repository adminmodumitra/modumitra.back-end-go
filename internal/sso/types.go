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
