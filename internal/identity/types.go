package identity

import "time"

const (
	PurposeEmailVerify   = "email_verify"
	PurposePasswordReset = "password_reset"
	PurposeOTPLogin      = "otp_login"

	PrefixUser    = "usr_"
	PrefixSession = "ses_"
	PrefixToken   = "tok_"
	PrefixFamily  = "fam_"
	PrefixRefresh = "rt_"
	PrefixVerify  = "ev_"
	PrefixReset   = "pr_"

	Issuer = "cabinet-studio"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

// User is the identity record. PasswordHash is never serialized to clients.
type User struct {
	ID              string
	Email           string
	EmailNormalized string
	EmailVerifiedAt *time.Time
	PasswordHash    string
	DisplayName     string
	TermsVersion    string
	PrivacyVersion  string
	Status          string
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (u *User) EmailVerified() bool {
	return u != nil && u.EmailVerifiedAt != nil
}

func (u *User) Suspended() bool {
	return u != nil && u.Status != "" && u.Status != StatusActive
}

// Session is a rotating refresh-token family member. Access JWTs are not stored.
type Session struct {
	ID              string
	UserID          string
	DeviceID        *string
	RefreshHash     string
	FamilyID        string
	ExpiresAt       time.Time
	RevokedAt       *time.Time
	ReuseDetectedAt *time.Time
	ReplacedBy      *string
	DeviceName      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastSeenAt      time.Time
}

func (s *Session) RevokedOrReplaced() bool {
	return s != nil && (s.RevokedAt != nil || s.ReplacedBy != nil)
}

// EmailToken is a single-use hashed email verify / password-reset / OTP token.
type EmailToken struct {
	ID         string
	UserID     string
	Purpose    string
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

// TokenPair is the public access+refresh response (no user).
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
}

type RegisterInput struct {
	Email          string
	Password       string
	DisplayName    string
	TermsVersion   string
	PrivacyVersion string
}

type LoginInput struct {
	Email      string
	Password   string
	DeviceName string
}

type Settings struct {
	TermsVersion   string
	PrivacyVersion string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	EmailTokenTTL  time.Duration
	OTPTTL         time.Duration
	JWTSecret      string
	Issuer         string
}
