package dto

import "github.com/Danish-13-Khan/backend-cabinet/internal/utils"

type RegisterBody struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	DisplayName    string `json:"display_name"`
	TermsVersion   string `json:"terms_version"`
	PrivacyVersion string `json:"privacy_version"`
}

type TokenBody struct {
	Token string `json:"token"`
}

type EmailBody struct {
	Email string `json:"email"`
}

type LoginBody struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceName string `json:"device_name"`
}

type RefreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

type ResetConfirmBody struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type OTPVerifyBody struct {
	Email      string `json:"email"`
	OTP        string `json:"otp"`
	DeviceName string `json:"device_name"`
}

type ChangePasswordBody struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
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
	// CSRFToken is set only for browser cookie sessions; send it back as
	// X-CSRF-Token on every write. refresh_token is then empty (it is a cookie).
	CSRFToken string `json:"csrf_token,omitempty"`
}

type LoginOK struct {
	TokenOK
	User         UserView          `json:"user"`
	Organization *OrganizationView `json:"organization"`
}
