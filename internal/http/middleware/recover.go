package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				// Omit panic value; it may include request secrets.
				l := obs.FromGin(c)
				ev := l.Error()
				if rid := RequestIDFrom(c); rid != "" {
					ev = ev.Str("request_id", rid)
				}
				if tid := TraceIDFrom(c); tid != "" {
					ev = ev.Str("trace_id", tid)
				}
				ev.Bytes("stack", debug.Stack()).Msg("panic recovered")
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
