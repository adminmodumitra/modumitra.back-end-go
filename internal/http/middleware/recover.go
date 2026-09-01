package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"request_id", RequestIDFrom(c),
					"trace_id", TraceIDFrom(c),
					"stack", string(debug.Stack()),
				)
				if !c.Writer.Written() {
					WriteError(c, http.StatusInternalServerError, "internal_error", "Internal server error", nil)
				} else {
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}
