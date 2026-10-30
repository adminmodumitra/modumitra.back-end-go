package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// gateCookie unlocks the editor files on app.; the CloudFront gate checks it
// (cabinet-designer-mvp edge/app-gate). It proves only "signed in until exp"
// and is signed with SESSION_GATE_KEY, so the edge never holds JWT_SECRET.
const gateCookie = "ms_gate"

// gateValue is "v1.<exp unix>.<base64url HMAC-SHA256(key, "v1.<exp>")>".
// The edge test uses the same vector as gate_cookie_test.go.
func gateValue(key string, exp time.Time) string {
	payload := "v1." + strconv.FormatInt(exp.Unix(), 10)
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (h *handlers) setGateCookie(c *gin.Context, maxAge int) {
	if h.cfg.Session.GateKey == "" || maxAge <= 0 {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name: gateCookie, Value: gateValue(h.cfg.Session.GateKey, time.Now().Add(time.Duration(maxAge)*time.Second)),
		Domain: h.cfg.Session.CookieDomain, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
}

func (h *handlers) clearGateCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: gateCookie, Domain: h.cfg.Session.CookieDomain, Path: "/",
		MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
}
