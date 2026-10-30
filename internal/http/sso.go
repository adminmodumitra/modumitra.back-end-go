package httpapi

import (
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireSSO(c *gin.Context) bool {
	if h.sso == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

// oauthProviders godoc
// @Summary List OAuth providers
// @Tags sso
// @Produce json
// @Success 200 {object} dto.OAuthProvidersOK
// @Router /v1/auth/oauth/providers [get]
func (h *handlers) oauthProviders(c *gin.Context) {
	logEnter(c, "oauth_providers")
	if !h.requireSSO(c) {
		return
	}
	items := h.sso.ListProviders()
	views := make([]dto.OAuthProviderView, 0, len(items))
	for _, p := range items {
		views = append(views, dto.OAuthProviderView{ID: p.ID, Label: p.Label, DeviceCodeSupported: p.DeviceCodeSupported})
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OAuthProvidersOK{
		Meta:      metaFrom(c),
		Providers: views,
	})
}

// oauthStart godoc
// @Summary Start OAuth authorization
// @Tags sso
// @Accept json
// @Produce json
// @Param provider path string true "Provider id"
// @Param body body dto.OAuthStartBody true "Start"
// @Success 200 {object} dto.OAuthStartOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/oauth/{provider}/start [post]
func (h *handlers) oauthStart(c *gin.Context) {
	logEnter(c, "oauth_start")
	if !h.requireSSO(c) {
		return
	}
	var body dto.OAuthStartBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	actorID := ""
	if strings.TrimSpace(body.Intent) == sso.IntentLink {
		if h.ident == nil {
			middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
			return
		}
		hdr := c.GetHeader(utils.HeaderAuthorization)
		const prefix = "Bearer "
		if hdr == "" || !strings.HasPrefix(hdr, prefix) {
			middleware.Abort(c, utils.E(utils.CodeUnauthenticated))
			return
		}
		uid, _, err := h.ident.ValidateAccess(c.Request.Context(), strings.TrimSpace(strings.TrimPrefix(hdr, prefix)))
		if err != nil {
			middleware.Abort(c, err)
			return
		}
		actorID = uid
	}
	res, err := h.sso.Start(c.Request.Context(), sso.StartInput{
		Provider:    c.Param("provider"),
		RedirectURI: body.RedirectURI,
		Intent:      body.Intent,
		ActorID:     actorID,
	})
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OAuthStartOK{
		Meta:             metaFrom(c),
		AuthorizationURL: res.AuthorizationURL,
		State:            res.State,
	})
}

func (h *handlers) writeOAuthLogin(c *gin.Context, u *identity.User, pair *identity.TokenPair, provider string) {
	if pair == nil {
		pair = &identity.TokenPair{TokenType: "Bearer"}
	}
	if u != nil {
		c.Set(utils.CtxUserID, u.ID)
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OAuthLoginOK{
		LoginOK: dto.LoginOK{
			TokenOK:      h.tokenView(c, pair),
			User:         userView(u),
			Organization: h.primaryOrgView(c, u.ID),
		},
		LinkedProvider: provider,
	})
}
