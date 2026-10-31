package httpapi

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"

	_ "github.com/Danish-13-Khan/backend-cabinet/docs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/engine"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
)

func NewRouter(cfg config.Config, pinger store.Pinger, ident *identity.Service, orgs *org.Service, devices *device.Service, bills *billing.Service, ssoSvc *sso.Service, opts ...Option) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	var c cache.Cache = cache.Noop{}
	if set, ok := pinger.(*store.Set); ok {
		c = cache.FromSet(set)
	}
	h := &handlers{cfg: cfg, pinger: pinger, cache: cache.New(c), ident: ident, orgs: orgs, devices: devices, billing: bills, sso: ssoSvc}
	h.engine = engine.New(cfg.Engine)
	h.quota = middleware.NewEngineQuota(cfg.Engine.UserPerMinute, cfg.Engine.OrgPerMinute, cfg.Engine.OrgPerDay)
	for _, opt := range opts {
		opt(h)
	}
	if ident != nil && bills != nil {
		ident.SetPersonalProvisioner(billing.Provisioner{Service: bills})
	}
	if orgs != nil && bills != nil {
		orgs.SetOrgProvisioner(billing.Provisioner{Service: bills})
	}
	if devices != nil && bills != nil {
		devices.SetPlanLookup(bills)
	}

	// Process-wide only: not limiters. Needed so healthz still has request_id/trace_id.
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(middleware.OTEL(cfg.ServiceName)...)
	// Before the allowlist: browsers send preflights without credentials.
	r.Use(middleware.CORS(cfg.CORSOrigins, cfg.Session.Origins...))
	r.Use(middleware.IPAllowlist(cfg.AllowedIPs))

	public := r.Group("/")
	public.Use(middleware.BBR(cfg.BBR))
	public.Use(middleware.IPRateLimit(rate.Limit(5), 20)) // health and swagger are exempt
	// Password, OTP and email guesses share one tighter budget per IP: 10 per minute.
	credential := middleware.IPRateLimit(rate.Every(6*time.Second), 10)
	{
		public.GET("/healthz", h.healthz)
		public.GET("/readyz", h.readyz)
		public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		public.GET("/v1/public/config", h.publicConfig)
		public.POST("/v1/auth/register", credential, h.register)
		public.POST("/v1/auth/verify-email", credential, h.verifyEmail)
		public.POST("/v1/auth/resend-verification", credential, h.resendVerification)
		public.POST("/v1/auth/login", credential, h.login)
		public.POST("/v1/auth/refresh", h.refresh)
		public.POST("/v1/auth/logout", h.logout)
		public.POST("/v1/auth/password-reset", credential, h.passwordReset)
		public.POST("/v1/auth/password-reset/confirm", credential, h.passwordResetConfirm)
		public.POST("/v1/auth/otp/request", credential, h.otpRequest)
		public.POST("/v1/auth/otp/verify", credential, h.otpVerify)
		public.POST("/v1/auth/otp/resend", credential, h.otpRequest)
		public.GET("/v1/invitations/:token", h.previewInvitation)
		public.GET("/v1/auth/oauth/providers", h.oauthProviders)
		public.POST("/v1/auth/oauth/:provider/start", h.oauthStart)
		public.GET("/v1/auth/oauth/:provider/callback", h.oauthCallback)
		public.POST("/v1/auth/device/code", h.deviceCode)
		public.POST("/v1/auth/device/token", h.deviceToken)
	}

	authed := r.Group("/")
	var verifier middleware.AccessVerifier
	if ident != nil {
		verifier = ident
	}
	// Shared with the engine group so both draw from the same buckets.
	authIP := middleware.IPRateLimit(rate.Limit(30), 60)
	authJWT := middleware.JWTAuth(verifier) // no skip list; these groups are auth-only
	// Cookie-authenticated writes need a session Origin and the CSRF token.
	authGuard := middleware.SessionGuard(cfg.Session.Enabled(), cfg.Session.Origins, middleware.CSRFKey(cfg.JWTSecret))
	authUser := middleware.UserRateLimit(rate.Limit(10), 40)
	authed.Use(middleware.BBR(cfg.BBR))
	authed.Use(middleware.Timeout(cfg.RequestTimeout))
	authed.Use(middleware.BodyLimit(cfg.BodyLimitBytes))
	authed.Use(authIP, authJWT, authGuard, authUser)
	authed.Use(middleware.Tenant(orgs))
	authed.Use(middleware.RBAC())
	authed.Use(middleware.Idempotency())
	{
		authed.POST("/v1/auth/logout-all", h.logoutAll)
		authed.POST("/v1/auth/device/approve", h.deviceApprove)
		authed.GET("/v1/me", h.me)
		authed.GET("/v1/me/capabilities", h.meCapabilities)
		authed.POST("/v1/me/password", h.changePassword)
		authed.GET("/v1/me/identities", h.listIdentities)
		authed.POST("/v1/me/identities/:provider/link", h.linkIdentity)
		authed.DELETE("/v1/me/identities/:provider", h.unlinkIdentity)
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
		authed.GET("/v1/devices", h.listDevices)
		authed.POST("/v1/devices/activate", h.activateDevice)
		authed.PATCH("/v1/devices/:id", h.patchDevice)
		authed.DELETE("/v1/devices/:id", h.deleteDevice)
		authed.GET("/v1/organizations/:id/entitlements", h.orgEntitlements)
		authed.GET("/v1/billing/plans", h.listPlans)
		authed.GET("/v1/billing/subscription", h.getSubscription)
		authed.POST("/v1/billing/checkout-session", h.checkoutSession)
		authed.POST("/v1/billing/portal-session", h.portalSession)
		authed.POST("/v1/billing/subscription/cancel", h.cancelSubscription)
		authed.POST("/v1/billing/subscription/resume", h.resumeSubscription)
		authed.GET("/v1/billing/invoices", h.listInvoices)
		authed.GET("/v1/billing/invoices/:invoiceId", h.getInvoice)
		authed.GET("/v1/billing/seats", h.getSeats)
		authed.POST("/v1/billing/seats/preview-downgrade", h.previewDowngrade)
		authed.GET("/v1/organizations/:id/subscription", h.getSubscription)
		authed.POST("/v1/organizations/:id/checkout-session", h.checkoutSession)
		authed.POST("/v1/organizations/:id/portal-session", h.portalSession)
		authed.POST("/v1/organizations/:id/subscription/cancel", h.cancelSubscription)
		authed.POST("/v1/organizations/:id/subscription/resume", h.resumeSubscription)
		authed.GET("/v1/organizations/:id/invoices", h.listInvoices)
		authed.GET("/v1/organizations/:id/invoices/:invoiceId", h.getInvoice)
		authed.GET("/v1/organizations/:id/seats", h.getSeats)
		authed.POST("/v1/organizations/:id/seats/preview-downgrade", h.previewDowngrade)
		authed.GET("/v1/org/price-book", h.getPriceBook)
		authed.PUT("/v1/org/price-book", middleware.BodyLimit(pricebook.MaxBookBytes), h.putPriceBook)
	}

	mountEngine(r, h, middleware.BBR(cfg.BBR), authIP, authJWT, authGuard, authUser)

	// Webhook HMAC, not JWT. Process-wide RequestID/Recover/OTEL already applied.
	webhooks := r.Group("/")
	webhooks.Use(middleware.BBR(cfg.BBR))
	webhooks.Use(middleware.Timeout(cfg.RequestTimeout))
	webhooks.Use(middleware.BodyLimit(cfg.BodyLimitBytes))
	{
		webhooks.POST("/v1/webhooks/billing", h.billingWebhook)
	}

	r.NoRoute(h.noRoute)
	return r
}
