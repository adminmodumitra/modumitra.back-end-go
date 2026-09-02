package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
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

func emit(c *gin.Context, phase, op string, status int, errCode string) {
	l := obs.Logger()
	ev := l.Info()
	if phase == "exit" {
		ev = l.Error()
	}
	applyFields(ev, c, phase, op, status, errCode)
	ev.Msg("")
}

func applyFields(ev *zerolog.Event, c *gin.Context, phase, op string, status int, errCode string) {
	str := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			ev.Str(k, v)
		}
	}
	str("op", op)
	str("phase", phase)
	if c != nil {
		f := Fields(c)
		for _, k := range []string{
			"request_id", "trace_id", "user_id", "session_id", "org_id", "role",
			"method", "path", "device_id", "invoice_id", "provider", "event_id", "event_type",
		} {
			str(k, f[k])
		}
	}
	if phase == "success" || phase == "exit" {
		if status > 0 {
			ev.Int("status", status)
		}
	}
	if phase == "exit" {
		str("error_code", errCode)
	}
}

func requestPath(c *gin.Context) string {
	if fp := c.FullPath(); fp != "" {
		return fp
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return ""
}

func orgIDForLog(c *gin.Context) string {
	if id := OrgIDFrom(c); id != "" {
		// OrgIDFrom falls back to c.Param("id"), which is a device id on /devices/:id.
		fp := c.FullPath()
		path := requestPath(c)
		if strings.Contains(fp, "/organizations/") || strings.Contains(path, "/organizations/") {
			return id
		}
		if v, ok := c.Get(utils.CtxOrgID); ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
		if q := strings.TrimSpace(c.Query("organization_id")); q != "" {
			return q
		}
		return ""
	}
	return strings.TrimSpace(c.Query("organization_id"))
}

func deviceIDForLog(c *gin.Context) string {
	if q := strings.TrimSpace(c.Query("device_id")); q != "" {
		return q
	}
	fp := c.FullPath()
	path := requestPath(c)
	if strings.Contains(fp, "/devices/") || strings.Contains(path, "/devices/") {
		return strings.TrimSpace(c.Param("id"))
	}
	return ""
}

func ginString(c *gin.Context, key string) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func logged(c *gin.Context) bool {
	if v, ok := c.Get(ctxLogDone); ok {
		if b, ok := v.(bool); ok && b {
			return true
		}
	}
	return false
}

func markLogged(c *gin.Context) {
	c.Set(ctxLogDone, true)
}
