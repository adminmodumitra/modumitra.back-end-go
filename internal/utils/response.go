package utils

// Meta is the top-level request/trace identity on success bodies.
type Meta struct {
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
}

// ErrorResponse is the v1 error envelope (no wrapping data object).
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	TraceID   string         `json:"trace_id"`
	Details   map[string]any `json:"details"`
}
