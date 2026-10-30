package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

const refreshCookiePath = "/v1/auth"

// browserSession reports whether this request gets cookies: sessions are on
// and it comes from one of the session origins (app., login., register.).
func (h *handlers) browserSession(c *gin.Context) bool {
	return h.cfg.Session.Enabled() && middleware.OriginAllowed(c, h.cfg.Session.Origins)
}

// tokenView builds the token body. For a browser session it also sets the
// cookies, hides the refresh token from JavaScript and adds the CSRF token.
func (h *handlers) tokenView(c *gin.Context, pair *identity.TokenPair) dto.TokenOK {
	out := dto.TokenOK{
		Meta:         metaFrom(c),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}
	if pair.AccessToken == "" || !h.browserSession(c) {
		return out
	}
	_, sid, err := h.ident.ValidateAccess(c.Request.Context(), pair.AccessToken)
	if err != nil {
		return out
	}
	h.setSessionCookies(c, pair)
	out.RefreshToken = ""
	out.CSRFToken = middleware.CSRFToken(middleware.CSRFKey(h.cfg.JWTSecret), sid)
	return out
}

func (h *handlers) setSessionCookies(c *gin.Context, pair *identity.TokenPair) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: middleware.AccessCookie, Value: pair.AccessToken,
		Domain: h.cfg.Session.CookieDomain, Path: "/", MaxAge: pair.ExpiresIn,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	if pair.RefreshToken != "" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name: middleware.RefreshCookie, Value: pair.RefreshToken,
			Path: refreshCookiePath, MaxAge: int(h.cfg.RefreshTTL.Seconds()),
			HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
		})
	}
}

// clearSessionCookies runs on logout whenever sessions are on; expiring a
// cookie the browser does not have is harmless.
func (h *handlers) clearSessionCookies(c *gin.Context) {
	if !h.cfg.Session.Enabled() {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name: middleware.AccessCookie, Domain: h.cfg.Session.CookieDomain, Path: "/",
		MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: middleware.RefreshCookie, Path: refreshCookiePath,
		MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
}

// refreshTokenFrom takes the body token (desktop, API clients) or else the
// refresh cookie. The cookie is honoured only from a session origin, which is
// the CSRF check for these public endpoints.
func (h *handlers) refreshTokenFrom(c *gin.Context, body string) (string, error) {
	if t := strings.TrimSpace(body); t != "" {
		return t, nil
	}
	v, err := c.Cookie(middleware.RefreshCookie)
	if err != nil || strings.TrimSpace(v) == "" {
		return "", utils.E(utils.CodeValidation)
	}
	if !h.browserSession(c) {
		return "", utils.E(utils.CodeForbidden)
	}
	return strings.TrimSpace(v), nil
}
