package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxLogOp   = "log_op"
	ctxLogDone = "log_done"
)

// Enter records the handler op and writes the enter line.
func Enter(c *gin.Context, op string) {
	if c == nil {
		return
	}
	if op != "" {
		c.Set(ctxLogOp, op)
	}
	emit(c, "enter", op, 0, "")
}

// Success writes the success line after a 2xx response.
func Success(c *gin.Context, op string, status int) {
	if c == nil || logged(c) {
		return
	}
	if op == "" {
		op = OpFrom(c)
	}
	if op == "" {
		return
	}
	markLogged(c)
	emit(c, "success", op, status, "")
}

// Exit writes the exit line for a coded failure (no stack, no error string).
func Exit(c *gin.Context, op string, status int, errCode string) {
	if c == nil || logged(c) {
		return
	}
	if op == "" {
		op = OpFrom(c)
	}
	if op == "" {
		return
	}
	markLogged(c)
	emit(c, "exit", op, status, errCode)
}

// OpFrom returns the op stored by Enter.
func OpFrom(c *gin.Context) string {
	return ginString(c, ctxLogOp)
}

// Fields returns request-scoped log fields present on the gin context.
func Fields(c *gin.Context) map[string]string {
	out := map[string]string{}
	if c == nil {
		return out
	}
	put := func(k, v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			out[k] = v
		}
	}
	put("op", OpFrom(c))
	put("request_id", RequestIDFrom(c))
	put("trace_id", TraceIDFrom(c))
	put("user_id", UserIDFrom(c))
	put("session_id", SessionIDFrom(c))
	put("org_id", orgIDForLog(c))
	put("role", RoleFrom(c))
	if c.Request != nil {
		put("method", c.Request.Method)
		put("path", requestPath(c))
	}
	put("device_id", deviceIDForLog(c))
	put("invoice_id", strings.TrimSpace(c.Param("invoiceId")))
	put("provider", strings.TrimSpace(c.Param("provider")))
	put("event_id", ginString(c, "event_id"))
	put("event_type", ginString(c, "event_type"))
	return out
}
