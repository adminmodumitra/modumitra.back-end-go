package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func skipJWT(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/v1/public/config",
		"/v1/auth/register",
		"/v1/auth/verify-email",
		"/v1/auth/resend-verification",
		"/v1/auth/login",
		"/v1/auth/refresh",
		"/v1/auth/logout",
		"/v1/auth/password-reset",
		"/v1/auth/password-reset/confirm",
		"/v1/auth/otp/request",
		"/v1/auth/otp/verify",
		"/v1/auth/otp/resend":
		return true
	default:
		return false
	}
}

// AccessVerifier validates a Bearer access JWT and returns user and session public ids.
type AccessVerifier interface {
	ValidateAccess(ctx context.Context, token string) (userID, sessionID string, err error)
}

// JWTAuth skips health, public config, and public auth routes. All other
// routes require a valid Bearer JWT whose session is not revoked or expired.
func JWTAuth(v AccessVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipJWT(c.Request.URL.Path) {
			c.Next()
			return
		}
		c.Header(utils.HeaderWWWAuth, "Bearer")
		if v == nil {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		hdr := c.GetHeader(utils.HeaderAuthorization)
		const prefix = "Bearer "
		if hdr == "" || !strings.HasPrefix(hdr, prefix) {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(hdr, prefix))
		if raw == "" {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		uid, sid, err := v.ValidateAccess(c.Request.Context(), raw)
		if err != nil {
			Abort(c, err)
			return
		}
		c.Set(utils.CtxUserID, uid)
		c.Set(utils.CtxSessionID, sid)
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) string {
	if v, ok := c.Get(utils.CtxUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func SessionIDFrom(c *gin.Context) string {
	if v, ok := c.Get(utils.CtxSessionID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
