// Package engine calls the private pricing engine over the private network.
// api. is the only caller; the engine has no public route except /healthz.
package engine

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

// Ops the proxy forwards. Anything else is refused before a network call.
const (
	OpQuote    = "quote"
	OpEstimate = "estimate"
)

// MaxRequestBytes caps a client body (the design document) on /v1/engine/*.
const MaxRequestBytes = 4 << 20

// MaxResponseBytes caps what is copied back from the engine.
const MaxResponseBytes = 16 << 20

// Headers sent with every engine call.
const (
	HeaderKey       = "X-Engine-Key"
	HeaderOrgID     = "X-Org-Id"
	HeaderUserID    = "X-User-Id"
	HeaderRequestID = "X-Request-Id"
)

// Caller identifies who the engine is computing for (logs, per-org metering).
type Caller struct {
	OrgID     string
	UserID    string
	RequestID string
}

// Client is safe for concurrent use; one per process keeps the pool warm.
type Client struct {
	base   string
	secret string
	http   *http.Client
}

// New returns nil when the engine is not configured (ENGINE_URL empty).
func New(cfg config.Engine) *Client {
	if !cfg.Enabled() {
		return nil
	}
	tr := &http.Transport{
		Proxy:                 nil, // private network: never via an env proxy
		DialContext:           (&net.Dialer{Timeout: 2 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: cfg.Timeout,
		ForceAttemptHTTP2:     true,
	}
	return &Client{base: cfg.URL, secret: cfg.SharedSecret, http: &http.Client{Transport: tr, Timeout: cfg.Timeout}}
}

// Do POSTs body to {ENGINE_URL}/v1/ops/<op>. The caller closes resp.Body.
func (c *Client) Do(ctx context.Context, op string, body []byte, who Caller) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/ops/"+op, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set(HeaderKey, c.secret)
	req.Header.Set(HeaderOrgID, who.OrgID)
	req.Header.Set(HeaderUserID, who.UserID)
	req.Header.Set(HeaderRequestID, who.RequestID)
	return c.http.Do(req)
}

// IsTimeout reports a client timeout or deadline (→ 504), as opposed to a
// refused or reset connection (→ 502).
func IsTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}
