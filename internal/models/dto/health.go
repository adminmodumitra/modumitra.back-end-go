package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type HealthOK struct {
	utils.Meta
	Status string `json:"status"`
}

type ReadyOK struct {
	utils.Meta
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

type PublicConfigOK struct {
	utils.Meta
	TermsVersion   string   `json:"terms_version"`
	PrivacyVersion string   `json:"privacy_version"`
	OAuthProviders []string `json:"oauth_providers"`
	MinAppVersion  string   `json:"min_app_version"`
	TrialDays      int      `json:"trial_days"`
}

type StatusOK struct {
	utils.Meta
	Status string `json:"status"`
}

type WebhookOK struct {
	utils.Meta
	Received bool `json:"received"`
}
