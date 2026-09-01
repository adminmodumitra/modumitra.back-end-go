package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OTEL is logical middleware #3: otelgin (W3C extract/inject) plus
// X-Trace-Id / traceparent on the response. If the client sent
// traceparent, otelgin continues that trace_id.
func OTEL(serviceName string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		otelgin.Middleware(serviceName),
		func(c *gin.Context) {
			span := trace.SpanFromContext(c.Request.Context())
			tid := ""
			if sc := span.SpanContext(); sc.HasTraceID() {
				tid = sc.TraceID().String()
			}
			c.Set(ctxTraceID, tid)
			span.SetAttributes(attribute.String("request_id", RequestIDFrom(c)))
			InjectHeaders(c)
			c.Next()
		},
	}
}
