package utils

import (
	"errors"
	"fmt"
)

// Code is a stable, machine-readable API error code (snake_case).
type Code string

func (c Code) String() string { return string(c) }

const (
	CodeInternal                Code = "internal_error"
	CodeUnauthenticated         Code = "unauthenticated"
	CodeForbidden               Code = "forbidden"
	CodeNotFound                Code = "not_found"
	CodeValidation              Code = "validation_error"
	CodeConflict                Code = "conflict"
	CodeRateLimited             Code = "rate_limited"
	CodeIdempotencyConflict     Code = "idempotency_conflict"
	CodeServiceUnavailable      Code = "service_unavailable"
	CodeTenantMismatch          Code = "tenant_mismatch"
	CodeEmailTaken              Code = "email_taken"
	CodeTokenReuse              Code = "token_reuse_detected"
	CodeDeviceLimit             Code = "device_limit_reached"
	CodeWeakPassword            Code = "weak_password"
	CodeUnknownDriver           Code = "unknown_datastore_driver"
	CodeInvalidCredentials      Code = "invalid_credentials"
	CodeInvalidToken            Code = "invalid_token"
	CodeTermsOutdated           Code = "terms_outdated"
	CodeAccountSuspended        Code = "account_suspended"
	CodeEmailUnverified         Code = "email_unverified"
	CodeLastOwner               Code = "last_owner"
	CodeAlreadyMember           Code = "already_member"
	CodeInvitationPending       Code = "invitation_pending"
	CodeInvitationNotPending    Code = "invitation_not_pending"
	CodeInvitationEmailMismatch Code = "invitation_email_mismatch"
	CodeInvitationExpired       Code = "invitation_expired"
	CodeInvitationRevoked       Code = "invitation_revoked"
	CodeInvitationAccepted      Code = "invitation_accepted"
	CodeRoleNotAssignable       Code = "role_not_assignable"
	CodeAlreadyOwner            Code = "already_owner"
	CodeCannotSuspendSelf       Code = "cannot_suspend_self"
	CodeCannotRemoveSelf        Code = "cannot_remove_self"
)

// Error is the non-generic application error. It always has a Code, a
// default user-facing Message, and an HTTP status. Wrap a cause with With
// when you have a real underlying error. Use Generic for unexpected failures.
type Error struct {
	Code    Code
	Name    string // same as Code; kept explicit for logs/metrics
	Message string
	HTTP    int
	Details map[string]any
	cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || e == nil || t == nil {
		return false
	}
	return e.Code == t.Code
}

func (e *Error) clone() *Error {
	if e == nil {
		return Generic(nil)
	}
	cp := *e
	if e.Details != nil {
		cp.Details = map[string]any{}
		for k, v := range e.Details {
			cp.Details[k] = v
		}
	}
	return &cp
}

// With attaches a cause without changing Code/Message/HTTP.
func (e *Error) With(cause error) *Error {
	cp := e.clone()
	cp.cause = cause
	return cp
}

// WithMessage overrides the user-facing message for this instance.
func (e *Error) WithMessage(msg string) *Error {
	cp := e.clone()
	cp.Message = msg
	return cp
}

// WithDetails merges extra fields into the API details object.
func (e *Error) WithDetails(details map[string]any) *Error {
	cp := e.clone()
	if cp.Details == nil {
		cp.Details = map[string]any{}
	}
	for k, v := range details {
		cp.Details[k] = v
	}
	return cp
}

type def struct {
	http    int
	message string
}

var catalog = map[Code]def{
	CodeInternal:                {StatusInternal, "Internal server error"},
	CodeUnauthenticated:         {StatusUnauthorized, "Authentication required"},
	CodeForbidden:               {StatusForbidden, "Forbidden"},
	CodeNotFound:                {StatusNotFound, "Not found"},
	CodeValidation:              {StatusUnprocessableEntity, "Invalid request"},
	CodeConflict:                {StatusConflict, "Conflict"},
	CodeRateLimited:             {StatusTooManyRequests, "Too many requests"},
	CodeIdempotencyConflict:     {StatusConflict, "Idempotency key reused with a different body"},
	CodeServiceUnavailable:      {StatusServiceUnavailable, "Service is not ready"},
	CodeTenantMismatch:          {StatusNotFound, "Not found"},
	CodeEmailTaken:              {StatusConflict, "Email already registered"},
	CodeTokenReuse:              {StatusUnauthorized, "Refresh token reuse detected"},
	CodeDeviceLimit:             {StatusConflict, "Device limit reached"},
	CodeWeakPassword:            {StatusUnprocessableEntity, "Password does not meet requirements"},
	CodeUnknownDriver:           {StatusInternal, "Unknown datastore driver"},
	CodeInvalidCredentials:      {StatusUnauthorized, "Invalid email or password"},
	CodeInvalidToken:            {StatusBadRequest, "Invalid or expired token"},
	CodeTermsOutdated:           {StatusUnprocessableEntity, "Terms or privacy version is not current"},
	CodeAccountSuspended:        {StatusForbidden, "Account is suspended"},
	CodeEmailUnverified:         {StatusForbidden, "Email is not verified"},
	CodeLastOwner:               {StatusConflict, "Last owner cannot leave or be removed"},
	CodeAlreadyMember:           {StatusConflict, "Already a member of this organization"},
	CodeInvitationPending:       {StatusConflict, "A pending invitation for this email already exists"},
	CodeInvitationNotPending:    {StatusConflict, "Invitation is not pending"},
	CodeInvitationEmailMismatch: {StatusForbidden, "Authenticated email does not match invitation"},
	CodeInvitationExpired:       {StatusConflict, "Invitation has expired"},
	CodeInvitationRevoked:       {StatusConflict, "Invitation has been revoked"},
	CodeInvitationAccepted:      {StatusConflict, "Invitation has already been accepted"},
	CodeRoleNotAssignable:       {StatusUnprocessableEntity, "Role is not assignable on this route"},
	CodeAlreadyOwner:            {StatusConflict, "User is already the owner"},
	CodeCannotSuspendSelf:       {StatusConflict, "Cannot suspend yourself"},
	CodeCannotRemoveSelf:        {StatusConflict, "Use leave to remove yourself"},
}

// E returns a coded error from the catalog. Unknown codes become Generic.
func E(code Code) *Error {
	d, ok := catalog[code]
	if !ok {
		return Generic(fmt.Errorf("unknown error code %q", code))
	}
	return &Error{
		Code:    code,
		Name:    code.String(),
		Message: d.message,
		HTTP:    d.http,
		Details: map[string]any{},
	}
}

// Generic wraps an unexpected failure as internal_error.
// The cause is kept for logs; the client still sees the generic message.
func Generic(err error) *Error {
	d := catalog[CodeInternal]
	return &Error{
		Code:    CodeInternal,
		Name:    CodeInternal.String(),
		Message: d.message,
		HTTP:    d.http,
		Details: map[string]any{},
		cause:   err,
	}
}

// From returns err if it is already *Error, otherwise Generic(err).
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return Generic(err)
}

// New is an explicit coded error with a custom message (still must be a known Code).
func New(code Code, message string) *Error {
	e := E(code)
	if message != "" {
		e.Message = message
	}
	return e
}
