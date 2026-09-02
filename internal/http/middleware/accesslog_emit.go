package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
