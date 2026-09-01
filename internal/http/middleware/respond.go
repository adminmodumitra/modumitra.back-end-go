package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

const (
	ctxRequestID = "request_id"
	ctxTraceID   = "trace_id"
)

func RequestIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return c.GetHeader(utils.HeaderRequestID)
}

func TraceIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxTraceID); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	sc := trace.SpanFromContext(c.Request.Context()).SpanContext()
	if sc.HasTraceID() {
		return sc.TraceID().String()
	}
	return ""
}

func InjectHeaders(c *gin.Context) {
	rid := RequestIDFrom(c)
	tid := TraceIDFrom(c)
	if tid == "" {
		sc := trace.SpanFromContext(c.Request.Context()).SpanContext()
		if sc.HasTraceID() {
			tid = sc.TraceID().String()
			c.Set(ctxTraceID, tid)
		}
	}
	if rid != "" {
		c.Header(utils.HeaderRequestID, rid)
	}
	if tid != "" {
		c.Header(utils.HeaderTraceID, tid)
	}
	otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.HeaderCarrier(c.Writer.Header()))
}

func WriteJSON(c *gin.Context, status int, body any) {
	InjectHeaders(c)
	c.JSON(status, body)
}

func WriteError(c *gin.Context, status int, code, message string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	InjectHeaders(c)
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":       code,
			"message":    message,
			"request_id": RequestIDFrom(c),
			"trace_id":   TraceIDFrom(c),
			"details":    details,
		},
	})
}

// Abort writes a catalogued *utils.Error (or Generic of a plain error).
func Abort(c *gin.Context, err error) {
	e := utils.From(err)
	if e == nil {
		e = utils.Generic(nil)
	}
	WriteError(c, e.HTTP, e.Code.String(), e.Message, e.Details)
}
