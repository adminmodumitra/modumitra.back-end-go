package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// Browser session cookies (subdomain setup). The access cookie is shared on
// the parent domain so the app. edge gate can read it; the refresh cookie
// stays on api. under /v1/auth.
const (
	AccessCookie  = "ms_access"
	RefreshCookie = "ms_refresh"
	HeaderCSRF    = "X-CSRF-Token"
	ctxViaCookie  = "auth_via_cookie"
)

// CSRFKey derives the CSRF HMAC key from the JWT secret, so no new secret is
// needed and rotating JWT_SECRET rotates CSRF tokens with the sessions.
func CSRFKey(jwtSecret string) []byte {
	sum := sha256.Sum256([]byte("csrf-v1\x00" + jwtSecret))
	return sum[:]
}

// CSRFToken is bound to the session id. Refresh rotates the session, so every
// refresh answer carries the new token; a stolen old one dies with its session.
func CSRFToken(key []byte, sessionID string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(sessionID))
	return hex.EncodeToString(m.Sum(nil))
}

// OriginAllowed reports whether the request's Origin is exactly one of origins.
func OriginAllowed(c *gin.Context, origins []string) bool {
	origin := strings.TrimRight(c.GetHeader("Origin"), "/")
	if origin == "" {
		return false
	}
	for _, o := range origins {
		if strings.TrimRight(strings.TrimSpace(o), "/") == origin {
			return true
		}
	}
	return false
}

func safeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

// SessionGuard runs after JWTAuth. Bearer requests pass untouched. A request
// authenticated by the access cookie must, when it changes state, come from
// an allowed Origin and carry the session's CSRF token. With cookie sessions
// off (enabled false), cookie authentication is refused outright.
func SessionGuard(enabled bool, origins []string, key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool(ctxViaCookie) {
			c.Next()
			return
		}
		if !enabled {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		if safeMethod(c.Request.Method) {
			c.Next()
			return
		}
		want := CSRFToken(key, SessionIDFrom(c))
		got := c.GetHeader(HeaderCSRF)
		if !OriginAllowed(c, origins) || !hmac.Equal([]byte(got), []byte(want)) {
			Abort(c, utils.E(utils.CodeForbidden))
			return
		}
		c.Next()
	}
}
