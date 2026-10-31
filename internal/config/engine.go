package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Engine is the private pricing engine behind POST /v1/engine/*. Off while URL
// is empty: those routes then answer 503 engine_unavailable.
type Engine struct {
	// URL is the private-network base, e.g. http://cabinet-engine:10000.
	// Env ENGINE_URL.
	URL string
	// SharedSecret is sent as X-Engine-Key. Env ENGINE_SHARED_SECRET.
	SharedSecret string
	// Timeout bounds one engine call; past it the client gets 504.
	// Env ENGINE_TIMEOUT (Go duration, default 8s).
	Timeout time.Duration
	// Quotas on quote+estimate calls. Env ENGINE_QUOTA_USER_PER_MIN,
	// ENGINE_QUOTA_ORG_PER_MIN, ENGINE_QUOTA_ORG_PER_DAY.
	UserPerMinute int
	OrgPerMinute  int
	OrgPerDay     int
	// Phase 4: a separate bucket for cut list, CNC and production-package
	// calls (the app computes them ahead after edits, up to three per edit).
	// Env ENGINE_MFG_QUOTA_USER_PER_MIN, ENGINE_MFG_QUOTA_ORG_PER_MIN,
	// ENGINE_MFG_QUOTA_ORG_PER_DAY.
	MfgUserPerMinute int
	MfgOrgPerMinute  int
	MfgOrgPerDay     int
	// Phase 5a: documents built on the server (proposal and revision summary
	// PDFs, quote export). User-initiated and CPU-heavy, so a small bucket of
	// their own. Env ENGINE_DOC_QUOTA_USER_PER_MIN, ENGINE_DOC_QUOTA_ORG_PER_MIN,
	// ENGINE_DOC_QUOTA_ORG_PER_DAY.
	DocUserPerMinute int
	DocOrgPerMinute  int
	DocOrgPerDay     int
	// DocTimeout bounds one document call (PDFs take seconds, not ms).
	// Env ENGINE_DOC_TIMEOUT (Go duration, default 30s).
	DocTimeout time.Duration
}

func defaultEngine() Engine {
	return Engine{
		Timeout: 8 * time.Second, UserPerMinute: 60, OrgPerMinute: 300, OrgPerDay: 5000,
		MfgUserPerMinute: 120, MfgOrgPerMinute: 600, MfgOrgPerDay: 20000,
		DocUserPerMinute: 10, DocOrgPerMinute: 60, DocOrgPerDay: 2000, DocTimeout: 30 * time.Second,
	}
}

func (e Engine) Enabled() bool { return strings.TrimSpace(e.URL) != "" }

func applyEngineEnv(e *Engine) {
	e.URL = strings.TrimRight(getenv("ENGINE_URL", e.URL), "/")
	e.SharedSecret = getenv("ENGINE_SHARED_SECRET", e.SharedSecret)
	e.Timeout = getenvDuration("ENGINE_TIMEOUT", e.Timeout)
	e.UserPerMinute = positive(getenvInt("ENGINE_QUOTA_USER_PER_MIN", e.UserPerMinute), e.UserPerMinute)
	e.OrgPerMinute = positive(getenvInt("ENGINE_QUOTA_ORG_PER_MIN", e.OrgPerMinute), e.OrgPerMinute)
	e.OrgPerDay = positive(getenvInt("ENGINE_QUOTA_ORG_PER_DAY", e.OrgPerDay), e.OrgPerDay)
	e.MfgUserPerMinute = positive(getenvInt("ENGINE_MFG_QUOTA_USER_PER_MIN", e.MfgUserPerMinute), e.MfgUserPerMinute)
	e.MfgOrgPerMinute = positive(getenvInt("ENGINE_MFG_QUOTA_ORG_PER_MIN", e.MfgOrgPerMinute), e.MfgOrgPerMinute)
	e.MfgOrgPerDay = positive(getenvInt("ENGINE_MFG_QUOTA_ORG_PER_DAY", e.MfgOrgPerDay), e.MfgOrgPerDay)
	e.DocUserPerMinute = positive(getenvInt("ENGINE_DOC_QUOTA_USER_PER_MIN", e.DocUserPerMinute), e.DocUserPerMinute)
	e.DocOrgPerMinute = positive(getenvInt("ENGINE_DOC_QUOTA_ORG_PER_MIN", e.DocOrgPerMinute), e.DocOrgPerMinute)
	e.DocOrgPerDay = positive(getenvInt("ENGINE_DOC_QUOTA_ORG_PER_DAY", e.DocOrgPerDay), e.DocOrgPerDay)
	if d := getenvDuration("ENGINE_DOC_TIMEOUT", e.DocTimeout); d > 0 {
		e.DocTimeout = d
	}
}

func positive(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

// Validate runs whenever ENGINE_URL is set, local dev included: the engine
// refuses calls without the right key, so a weak key is a broken deploy.
func (e Engine) Validate() error {
	if !e.Enabled() {
		return nil
	}
	u, err := url.Parse(e.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("ENGINE_URL must be an http(s) base URL")
	}
	if reason := weakSecret(e.SharedSecret); reason != "" {
		return fmt.Errorf("ENGINE_SHARED_SECRET %s; set a random value of at least %d characters (openssl rand -base64 48)", reason, minSecretLen)
	}
	return nil
}
