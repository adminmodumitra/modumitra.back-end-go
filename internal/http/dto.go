package httpapi

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func metaFrom(c *gin.Context) utils.Meta {
	return utils.Meta{
		RequestID: middleware.RequestIDFrom(c),
		TraceID:   middleware.TraceIDFrom(c),
	}
}

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

type UserView struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"display_name"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type UserOK struct {
	utils.Meta
	User UserView `json:"user"`
}

// StatusOK is a success body with a status string (accepted, logged_out, ...).
type StatusOK struct {
	utils.Meta
	Status string `json:"status"`
}

type LogoutAllOK struct {
	utils.Meta
	Status          string `json:"status"`
	RevokedSessions int    `json:"revoked_sessions"`
}

type TokenOK struct {
	utils.Meta
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// OrganizationView is filled in a later slice. Slice 1 always sends null.
type OrganizationView struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type LoginOK struct {
	TokenOK
	User         UserView          `json:"user"`
	Organization *OrganizationView `json:"organization"`
}

func userView(u *identity.User) UserView {
	if u == nil {
		return UserView{}
	}
	return UserView{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified(),
		DisplayName:   u.DisplayName,
		CreatedAt:     u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
