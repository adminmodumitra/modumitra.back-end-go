package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// IPAllowlist rejects callers whose IP is not in allowed.
// An empty list allows every IP (check off). /healthz, /readyz, and /swagger
// stay open so probes and docs keep working when the list is set. Every other
// route returns 401.
func IPAllowlist(allowed []string) gin.HandlerFunc {
	ips, nets := parseAllowlist(allowed)
	if len(ips) == 0 && len(nets) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" || strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}
		if !ipAllowed(clientIP(c), ips, nets) {
			Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		c.Next()
	}
}

// clientIP is the caller address the allowlist compares.
// Render's edge is Cloudflare, so the socket address is a private proxy hop.
// Cloudflare overwrites CF-Connecting-IP with the real client. That header is
// read only when the direct peer is a private or loopback address, so a
// client connecting straight to the process cannot spoof it.
func clientIP(c *gin.Context) string {
	peer := net.ParseIP(c.RemoteIP())
	if peer != nil && (peer.IsPrivate() || peer.IsLoopback() || peer.IsLinkLocalUnicast()) {
		for _, header := range []string{"CF-Connecting-IP", "True-Client-IP"} {
			if ip := headerIP(c, header); ip != "" {
				return ip
			}
		}
	}
	return c.ClientIP()
}

func headerIP(c *gin.Context, name string) string {
	raw := strings.TrimSpace(c.GetHeader(name))
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	return ip.String()
}

func parseAllowlist(allowed []string) ([]net.IP, []*net.IPNet) {
	var ips []net.IP
	var nets []*net.IPNet
	for _, raw := range allowed {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, n, err := net.ParseCIDR(entry)
			if err == nil {
				nets = append(nets, n)
			}
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips, nets
}

func ipAllowed(client string, ips []net.IP, nets []*net.IPNet) bool {
	ip := net.ParseIP(strings.TrimSpace(client))
	if ip == nil {
		return false
	}
	for _, allowed := range ips {
		if ip.Equal(allowed) {
			return true
		}
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
