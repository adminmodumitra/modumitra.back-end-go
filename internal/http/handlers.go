package httpapi

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type handlers struct {
	cfg     config.Config
	pinger  store.Pinger
	ident   *identity.Service
	orgs    *org.Service
	devices *device.Service
	billing *billing.Service
	sso     *sso.Service
}

func (h *handlers) healthz(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	slog.Info("healthz", "request_id", rid, "trace_id", tid)
	middleware.WriteJSON(c, utils.StatusOK, HealthOK{
		Meta:   metaFrom(c),
		Status: "ok",
	})
}

func (h *handlers) readyz(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	if h.pinger == nil {
		slog.Error("readyz missing pinger", "request_id", rid, "trace_id", tid)
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return
	}
	checks := map[string]string{}
	if checker, ok := h.pinger.(interface {
		Checks(context.Context) (map[string]string, error)
	}); ok {
		m, err := checker.Checks(c.Request.Context())
		for k, v := range m {
			checks[k] = v
		}
		if err != nil {
			slog.Error("readyz datastore ping failed", "request_id", rid, "trace_id", tid, "error", err)
			middleware.Abort(c, utils.E(utils.CodeServiceUnavailable).WithDetails(map[string]any{"checks": checks}))
			return
		}
	} else if err := h.pinger.Ping(c.Request.Context()); err != nil {
		slog.Error("readyz datastore ping failed", "request_id", rid, "trace_id", tid)
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return
	} else {
		checks["store"] = "ok"
	}
	slog.Info("readyz", "request_id", rid, "trace_id", tid, "checks", checks)
	middleware.WriteJSON(c, utils.StatusOK, ReadyOK{
		Meta:   metaFrom(c),
		Status: "ready",
		Checks: checks,
	})
}

func (h *handlers) publicConfig(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	slog.Info("public_config", "request_id", rid, "trace_id", tid)
	middleware.WriteJSON(c, utils.StatusOK, PublicConfigOK{
		Meta:           metaFrom(c),
		TermsVersion:   h.cfg.TermsVersion,
		PrivacyVersion: h.cfg.PrivacyVersion,
		OAuthProviders: h.cfg.OAuthProviders,
		MinAppVersion:  h.cfg.MinAppVersion,
		TrialDays:      h.cfg.TrialDays,
	})
}

func (h *handlers) noRoute(c *gin.Context) {
	// JWT is group-scoped, so unmatched paths 404 rather than 401.
	if !c.Writer.Written() {
		middleware.Abort(c, utils.E(utils.CodeNotFound))
	}
}
