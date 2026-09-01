package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

type handlers struct {
	cfg    config.Config
	pinger store.Pinger
}

func (h *handlers) healthz(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	slog.Info("healthz", "request_id", rid, "trace_id", tid)
	middleware.WriteJSON(c, http.StatusOK, gin.H{
		"request_id": rid,
		"trace_id":   tid,
		"status":     "ok",
	})
}

func (h *handlers) readyz(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	if h.pinger == nil {
		slog.Error("readyz missing pinger", "request_id", rid, "trace_id", tid)
		middleware.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "Service is not ready", nil)
		return
	}
	if err := h.pinger.Ping(c.Request.Context()); err != nil {
		slog.Error("readyz postgres ping failed", "request_id", rid, "trace_id", tid)
		middleware.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "Service is not ready", nil)
		return
	}
	slog.Info("readyz", "request_id", rid, "trace_id", tid)
	middleware.WriteJSON(c, http.StatusOK, gin.H{
		"request_id": rid,
		"trace_id":   tid,
		"status":     "ready",
		"checks": gin.H{
			"postgres": "ok",
		},
	})
}

func (h *handlers) publicConfig(c *gin.Context) {
	rid, tid := middleware.RequestIDFrom(c), middleware.TraceIDFrom(c)
	slog.Info("public_config", "request_id", rid, "trace_id", tid)
	middleware.WriteJSON(c, http.StatusOK, gin.H{
		"request_id":      rid,
		"trace_id":        tid,
		"terms_version":   h.cfg.TermsVersion,
		"privacy_version": h.cfg.PrivacyVersion,
		"oauth_providers": h.cfg.OAuthProviders,
		"min_app_version": h.cfg.MinAppVersion,
		"trial_days":      h.cfg.TrialDays,
	})
}

func (h *handlers) noRoute(c *gin.Context) {
	// JWT already 401s unknown /v1 and everything else; this is a fallback.
	if !c.Writer.Written() {
		middleware.WriteError(c, http.StatusNotFound, "not_found", "Not found", nil)
	}
}
