package httpapi

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func NewRouter(cfg config.Config, pinger store.Pinger, ident *identity.Service) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	useMiddleware(r, cfg, ident)
	registerRoutes(r, cfg, pinger, ident)
	return r
}

func useMiddleware(r *gin.Engine, cfg config.Config, ident *identity.Service) {
	var verifier middleware.AccessVerifier
	if ident != nil {
		verifier = ident
	}
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(middleware.OTEL(cfg.ServiceName)...)
	r.Use(middleware.BBR(cfg.BBR))
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(middleware.BodyLimit(cfg.BodyLimitBytes))
	r.Use(middleware.IPRateLimit(rate.Limit(30), 60))
	r.Use(middleware.UserRateLimit())
	r.Use(middleware.JWTAuth(verifier))
	r.Use(middleware.Tenant())
	r.Use(middleware.RBAC())
	r.Use(middleware.Idempotency())
}

func registerRoutes(r *gin.Engine, cfg config.Config, pinger store.Pinger, ident *identity.Service) {
	h := &handlers{cfg: cfg, pinger: pinger, ident: ident}
	r.GET("/healthz", h.healthz)
	r.GET("/readyz", h.readyz)
	r.GET("/v1/public/config", h.publicConfig)

	r.POST("/v1/auth/register", h.register)
	r.POST("/v1/auth/verify-email", h.verifyEmail)
	r.POST("/v1/auth/resend-verification", h.resendVerification)
	r.POST("/v1/auth/login", h.login)
	r.POST("/v1/auth/refresh", h.refresh)
	r.POST("/v1/auth/logout", h.logout)
	r.POST("/v1/auth/logout-all", h.logoutAll)
	r.POST("/v1/auth/password-reset", h.passwordReset)
	r.POST("/v1/auth/password-reset/confirm", h.passwordResetConfirm)
	r.POST("/v1/auth/otp/request", h.otpRequest)
	r.POST("/v1/auth/otp/verify", h.otpVerify)
	r.POST("/v1/auth/otp/resend", h.otpRequest)

	r.GET("/v1/me", h.me)
	r.POST("/v1/me/password", h.changePassword)

	r.NoRoute(h.noRoute)
}
