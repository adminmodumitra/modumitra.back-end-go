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

// JWTAuth always requires a valid Bearer JWT whose session is not revoked or expired.
// Mount only on auth-required groups; public routes never see it.
func JWTAuth(v AccessVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
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
