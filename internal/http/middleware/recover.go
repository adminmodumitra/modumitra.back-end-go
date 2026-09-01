package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
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
					Abort(c, utils.Generic(nil))
				} else {
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}
