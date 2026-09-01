package httpapi

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func NewRouter(cfg config.Config, pinger store.Pinger, ident *identity.Service, orgs *org.Service) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	h := &handlers{cfg: cfg, pinger: pinger, ident: ident, orgs: orgs}

	// Process-wide only: not limiters. Needed so healthz still has request_id/trace_id.
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(middleware.OTEL(cfg.ServiceName)...)

	public := r.Group("/")
	public.Use(middleware.BBR(cfg.BBR)) // ONLY limiter on free/non-auth APIs
	{
		public.GET("/healthz", h.healthz)
		public.GET("/readyz", h.readyz)
		public.GET("/v1/public/config", h.publicConfig)
		public.POST("/v1/auth/register", h.register)
		public.POST("/v1/auth/verify-email", h.verifyEmail)
		public.POST("/v1/auth/resend-verification", h.resendVerification)
		public.POST("/v1/auth/login", h.login)
		public.POST("/v1/auth/refresh", h.refresh)
		public.POST("/v1/auth/logout", h.logout)
		public.POST("/v1/auth/password-reset", h.passwordReset)
		public.POST("/v1/auth/password-reset/confirm", h.passwordResetConfirm)
		public.POST("/v1/auth/otp/request", h.otpRequest)
		public.POST("/v1/auth/otp/verify", h.otpVerify)
		public.POST("/v1/auth/otp/resend", h.otpRequest)
		public.GET("/v1/invitations/:token", h.previewInvitation)
	}

	authed := r.Group("/")
	var verifier middleware.AccessVerifier
	if ident != nil {
		verifier = ident
	}
	authed.Use(middleware.BBR(cfg.BBR))
	authed.Use(middleware.Timeout(cfg.RequestTimeout))
	authed.Use(middleware.BodyLimit(cfg.BodyLimitBytes))
	authed.Use(middleware.IPRateLimit(rate.Limit(30), 60))
	authed.Use(middleware.UserRateLimit())
	authed.Use(middleware.JWTAuth(verifier)) // no skip list; this group is auth-only
	authed.Use(middleware.Tenant(orgs))
	authed.Use(middleware.RBAC())
	authed.Use(middleware.Idempotency())
	{
		authed.POST("/v1/auth/logout-all", h.logoutAll)
		authed.GET("/v1/me", h.me)
		authed.POST("/v1/me/password", h.changePassword)
		authed.POST("/v1/organizations", h.createOrg)
		authed.GET("/v1/organizations", h.listOrgs)
		authed.GET("/v1/organizations/:id", h.getOrg)
		authed.PATCH("/v1/organizations/:id", h.patchOrg)
		authed.DELETE("/v1/organizations/:id", h.deleteOrg)
		authed.POST("/v1/organizations/:id/leave", h.leaveOrg)
		authed.POST("/v1/organizations/:id/transfer-ownership", h.transferOwnership)
		authed.GET("/v1/organizations/:id/roles", h.orgRoles)
		authed.GET("/v1/organizations/:id/members", h.listMembers)
		authed.PATCH("/v1/organizations/:id/members/:userId", h.patchMember)
		authed.POST("/v1/organizations/:id/members/:userId/suspend", h.suspendMember)
		authed.DELETE("/v1/organizations/:id/members/:userId", h.removeMember)
		authed.POST("/v1/organizations/:id/invitations", h.createInvitation)
		authed.GET("/v1/organizations/:id/invitations", h.listInvitations)
		authed.DELETE("/v1/organizations/:id/invitations/:invitationId", h.revokeInvitation)
		authed.POST("/v1/organizations/:id/invitations/:invitationId/resend", h.resendInvitation)
		authed.POST("/v1/invitations/:token/accept", h.acceptInvitation)
	}

	r.NoRoute(h.noRoute)
	return r
}
