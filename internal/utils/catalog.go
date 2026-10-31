package utils

// Code is a stable, machine-readable API error code (snake_case).

const (
	CodeInternal                  Code = "internal_error"
	CodeUnauthenticated           Code = "unauthenticated"
	CodeForbidden                 Code = "forbidden"
	CodeNotFound                  Code = "not_found"
	CodeValidation                Code = "validation_error"
	CodeConflict                  Code = "conflict"
	CodeRateLimited               Code = "rate_limited"
	CodeIdempotencyConflict       Code = "idempotency_conflict"
	CodeServiceUnavailable        Code = "service_unavailable"
	CodeTenantMismatch            Code = "tenant_mismatch"
	CodeEmailTaken                Code = "email_taken"
	CodeTokenReuse                Code = "token_reuse_detected"
	CodeDeviceLimit               Code = "device_limit_reached"
	CodeWeakPassword              Code = "weak_password"
	CodeUnknownDriver             Code = "unknown_datastore_driver"
	CodeInvalidCredentials        Code = "invalid_credentials"
	CodeInvalidToken              Code = "invalid_token"
	CodeTermsOutdated             Code = "terms_outdated"
	CodeAccountSuspended          Code = "account_suspended"
	CodeEmailUnverified           Code = "email_unverified"
	CodeLastOwner                 Code = "last_owner"
	CodeAlreadyMember             Code = "already_member"
	CodeInvitationPending         Code = "invitation_pending"
	CodeInvitationNotPending      Code = "invitation_not_pending"
	CodeInvitationEmailMismatch   Code = "invitation_email_mismatch"
	CodeInvitationExpired         Code = "invitation_expired"
	CodeInvitationRevoked         Code = "invitation_revoked"
	CodeInvitationAccepted        Code = "invitation_accepted"
	CodeRoleNotAssignable         Code = "role_not_assignable"
	CodeAlreadyOwner              Code = "already_owner"
	CodeCannotSuspendSelf         Code = "cannot_suspend_self"
	CodeCannotRemoveSelf          Code = "cannot_remove_self"
	CodeDeviceNotOwned            Code = "device_not_owned"
	CodeInvalidSignature          Code = "invalid_signature"
	CodeInvalidPayload            Code = "invalid_payload"
	CodePlanNotFound              Code = "plan_not_found"
	CodeSubscriptionActive        Code = "subscription_active"
	CodeRedirectURINotAllowed     Code = "redirect_uri_not_allowed"
	CodeBillingCustomerMissing    Code = "billing_customer_missing"
	CodeAlreadyCanceled           Code = "already_canceled"
	CodeSubscriptionNotCancelable Code = "subscription_not_cancelable"
	CodeNotCanceling              Code = "not_canceling"
	CodeSubscriptionNotResumable  Code = "subscription_not_resumable"
	CodeSeatLimit                 Code = "seat_limit"
	CodeInvalidState              Code = "invalid_state"
	CodeIdentityAlreadyLinked     Code = "identity_already_linked"
	CodeProviderAlreadyLinked     Code = "provider_already_linked"
	CodeLastLoginMethod           Code = "last_login_method"
	CodeAuthorizationPending      Code = "authorization_pending"
	CodeUnsupportedProvider       Code = "unsupported_provider"
	CodeOAuthDenied               Code = "oauth_denied"
	CodeInvalidClient             Code = "invalid_client"
	CodeSlowDown                  Code = "slow_down"
	CodeExpiredToken              Code = "expired_token"
	CodeAccessDenied              Code = "access_denied"
	CodeInvalidGrant              Code = "invalid_grant"
	CodePayloadTooLarge           Code = "payload_too_large"
	CodeQuotaExceeded             Code = "quota_exceeded"
	CodeEngineUnavailable         Code = "engine_unavailable"
	CodeEngineError               Code = "engine_error"
	CodeEngineTimeout             Code = "engine_timeout"
)

type def struct {
	http    int
	message string
}

var catalog = map[Code]def{
	CodeInternal:                  {StatusInternal, "Internal server error"},
	CodeUnauthenticated:           {StatusUnauthorized, "Authentication required"},
	CodeForbidden:                 {StatusForbidden, "Forbidden"},
	CodeNotFound:                  {StatusNotFound, "Not found"},
	CodeValidation:                {StatusUnprocessableEntity, "Invalid request"},
	CodeConflict:                  {StatusConflict, "Conflict"},
	CodeRateLimited:               {StatusTooManyRequests, "Too many requests"},
	CodeIdempotencyConflict:       {StatusConflict, "Idempotency key reused with a different body"},
	CodeServiceUnavailable:        {StatusServiceUnavailable, "Service is not ready"},
	CodeTenantMismatch:            {StatusNotFound, "Not found"},
	CodeEmailTaken:                {StatusConflict, "Email already registered"},
	CodeTokenReuse:                {StatusUnauthorized, "Refresh token reuse detected"},
	CodeDeviceLimit:               {StatusConflict, "Device limit reached"},
	CodeWeakPassword:              {StatusUnprocessableEntity, "Password does not meet requirements"},
	CodeUnknownDriver:             {StatusInternal, "Unknown datastore driver"},
	CodeInvalidCredentials:        {StatusUnauthorized, "Invalid email or password"},
	CodeInvalidToken:              {StatusBadRequest, "Invalid or expired token"},
	CodeTermsOutdated:             {StatusUnprocessableEntity, "Terms or privacy version is not current"},
	CodeAccountSuspended:          {StatusForbidden, "Account is suspended"},
	CodeEmailUnverified:           {StatusForbidden, "Email is not verified"},
	CodeLastOwner:                 {StatusConflict, "Last owner cannot leave or be removed"},
	CodeAlreadyMember:             {StatusConflict, "Already a member of this organization"},
	CodeInvitationPending:         {StatusConflict, "A pending invitation for this email already exists"},
	CodeInvitationNotPending:      {StatusConflict, "Invitation is not pending"},
	CodeInvitationEmailMismatch:   {StatusForbidden, "Authenticated email does not match invitation"},
	CodeInvitationExpired:         {StatusConflict, "Invitation has expired"},
	CodeInvitationRevoked:         {StatusConflict, "Invitation has been revoked"},
	CodeInvitationAccepted:        {StatusConflict, "Invitation has already been accepted"},
	CodeRoleNotAssignable:         {StatusUnprocessableEntity, "Role is not assignable on this route"},
	CodeAlreadyOwner:              {StatusConflict, "User is already the owner"},
	CodeCannotSuspendSelf:         {StatusConflict, "Cannot suspend yourself"},
	CodeCannotRemoveSelf:          {StatusConflict, "Use leave to remove yourself"},
	CodeDeviceNotOwned:            {StatusBadRequest, "Device is not an active device of the caller"},
	CodeInvalidSignature:          {StatusUnauthorized, "Invalid webhook signature"},
	CodeInvalidPayload:            {StatusBadRequest, "Invalid webhook payload"},
	CodePlanNotFound:              {StatusNotFound, "Plan not found"},
	CodeSubscriptionActive:        {StatusConflict, "Organization already has an active subscription"},
	CodeRedirectURINotAllowed:     {StatusUnprocessableEntity, "Redirect URL is not allowlisted"},
	CodeBillingCustomerMissing:    {StatusConflict, "No billing customer; use checkout-session"},
	CodeAlreadyCanceled:           {StatusConflict, "Subscription is already canceled"},
	CodeSubscriptionNotCancelable: {StatusConflict, "Subscription cannot be canceled"},
	CodeNotCanceling:              {StatusConflict, "Subscription is not scheduled to cancel"},
	CodeSubscriptionNotResumable:  {StatusConflict, "Subscription cannot be resumed"},
	CodeSeatLimit:                 {StatusConflict, "Active members exceed target seats"},
	CodeInvalidState:              {StatusBadRequest, "OAuth state is invalid, expired, or already used"},
	CodeIdentityAlreadyLinked:     {StatusConflict, "This identity is already linked to another user"},
	CodeProviderAlreadyLinked:     {StatusConflict, "This provider is already linked"},
	CodeLastLoginMethod:           {StatusConflict, "Cannot remove the last login method"},
	CodeAuthorizationPending:      {StatusPreconditionRequired, "User has not yet approved this device."},
	CodeUnsupportedProvider:       {StatusBadRequest, "Provider is not enabled"},
	CodeOAuthDenied:               {StatusBadRequest, "The user denied the authorization request"},
	CodeInvalidClient:             {StatusUnauthorized, "Unknown client_id"},
	CodeSlowDown:                  {StatusForbidden, "Polled faster than the required interval"},
	CodeExpiredToken:              {StatusForbidden, "Device code has expired"},
	CodeAccessDenied:              {StatusForbidden, "The user denied the device authorization request"},
	CodeInvalidGrant:              {StatusBadRequest, "Device code is unknown or already consumed"},
	CodePayloadTooLarge:           {StatusPayloadTooLarge, "Request body is too large"},
	CodeQuotaExceeded:             {StatusTooManyRequests, "Pricing quota reached; try again later"},
	CodeEngineUnavailable:         {StatusServiceUnavailable, "Pricing engine is not configured"},
	CodeEngineError:               {StatusBadGateway, "Pricing engine is unavailable"},
	CodeEngineTimeout:             {StatusGatewayTimeout, "Pricing engine timed out"},
}
