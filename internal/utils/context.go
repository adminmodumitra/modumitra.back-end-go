package utils

import "context"

type requestIDCtxKey struct{}

// WithRequestID stores the HTTP request id on ctx so domain code (outbox
// envelopes, logs) can read the same value as X-Request-Id without Gin.
func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDCtxKey{}, id)
}

// RequestIDFromContext returns the request id previously stored with
// WithRequestID, or empty.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}
