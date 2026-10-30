package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// AccessVerifier validates a Bearer access JWT and returns user and session public ids.
type AccessVerifier interface {
	ValidateAccess(ctx context.Context, token string) (userID, sessionID string, err error)
}

// JWTAuth always requires a valid access JWT whose session is not revoked or
// expired: the Bearer header, or else the ms_access cookie (then SessionGuard
// must follow). Mount only on auth-required groups; public routes never see it.
func JWTAuth(v AccessVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(utils.HeaderWWWAuth, "Bearer")
		if v == nil {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		raw, viaCookie := accessTokenFrom(c)
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
		c.Set(ctxViaCookie, viaCookie)
		c.Next()
	}
}

// accessTokenFrom prefers the Authorization header, so the desktop app and
// API clients never depend on cookies.
func accessTokenFrom(c *gin.Context) (string, bool) {
	const prefix = "Bearer "
	if hdr := c.GetHeader(utils.HeaderAuthorization); hdr != "" {
		if !strings.HasPrefix(hdr, prefix) {
			return "", false
		}
		return strings.TrimSpace(strings.TrimPrefix(hdr, prefix)), false
	}
	if v, err := c.Cookie(AccessCookie); err == nil {
		return strings.TrimSpace(v), true
	}
	return "", false
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
