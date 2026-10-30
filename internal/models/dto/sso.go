package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type OAuthStartBody struct {
	RedirectURI string `json:"redirect_uri"`
	Intent      string `json:"intent"`
}

type OAuthLinkBody struct {
	AuthorizationCode string `json:"authorization_code"`
	State             string `json:"state"`
}

type DeviceCodeBody struct {
	ClientID string `json:"client_id"`
	Provider string `json:"provider"`
}

type DeviceApproveBody struct {
	UserCode string `json:"user_code"`
}

type DeviceTokenBody struct {
	DeviceCode string `json:"device_code"`
}

type OAuthProviderView struct {
	ID                  string `json:"id"`
	Label               string `json:"label"`
	DeviceCodeSupported bool   `json:"device_code_supported"`
}

type OAuthProvidersOK struct {
	utils.Meta
	Providers []OAuthProviderView `json:"providers"`
}

type OAuthStartOK struct {
	utils.Meta
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

type OAuthLoginOK struct {
	LoginOK
	LinkedProvider string `json:"linked_provider"`
}

type IdentityView struct {
	Provider       string  `json:"provider"`
	ProviderUserID string  `json:"provider_user_id"`
	Email          *string `json:"email"`
	LinkedAt       string  `json:"linked_at"`
}

type IdentityListOK struct {
	utils.Meta
	Items      []IdentityView `json:"items"`
	NextCursor *string        `json:"next_cursor"`
}

type IdentityOK struct {
	utils.Meta
	Identity IdentityView `json:"identity"`
}

type UnlinkOK struct {
	utils.Meta
	Status   string `json:"status"`
	Provider string `json:"provider"`
}

type DeviceCodeOK struct {
	utils.Meta
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}
