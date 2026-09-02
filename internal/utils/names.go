package utils

// Canonical string names used across HTTP, logs, datastore, and tracing.
// Import this package instead of scattering magic strings.

const (
	HeaderRequestID        = "X-Request-Id"
	HeaderTraceID          = "X-Trace-Id"
	HeaderTraceparent      = "traceparent"
	HeaderTracestate       = "tracestate"
	HeaderAuthorization    = "Authorization"
	HeaderIdempotency      = "Idempotency-Key"
	HeaderRetryAfter       = "Retry-After"
	HeaderWWWAuth          = "WWW-Authenticate"
	HeaderOrganization     = "X-Organization-Id"
	HeaderStripeSignature  = "Stripe-Signature"
	HeaderBillingSignature = "X-Billing-Signature"

	CtxRequestID = "request_id"
	CtxTraceID   = "trace_id"
	CtxUserID    = "user_id"
	CtxSessionID = "session_id"
	CtxOrgID     = "organization_id"
	CtxRole      = "role"

	DriverPostgres = "postgres"
	DriverSQL      = "sql"
	DriverRedis    = "redis"
	DriverMongo    = "mongodb"
	DriverNoSQL    = "nosql"

	SinkStderr     = "stderr"
	SinkStdout     = "stdout"
	SinkLoki       = "loki"
	SinkCloudWatch = "cloudwatch"

	JSONCode      = "code"
	JSONMessage   = "message"
	JSONRequestID = "request_id"
	JSONTraceID   = "trace_id"
	JSONDetails   = "details"
	JSONError     = "error"
)
