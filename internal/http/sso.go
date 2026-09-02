package httpapi

import (
	"strings"

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

func (h *handlers) oauthProviders(c *gin.Context) {
	logEnter(c, "oauth_providers")
	if !h.requireSSO(c) {
		return
	}
	items := h.sso.ListProviders()
	views := make([]OAuthProviderView, 0, len(items))
	for _, p := range items {
		views = append(views, OAuthProviderView{ID: p.ID, Label: p.Label, DeviceCodeSupported: p.DeviceCodeSupported})
	}
	middleware.WriteJSON(c, utils.StatusOK, OAuthProvidersOK{
		Meta:      metaFrom(c),
		Providers: views,
	})
}

type oauthStartBody struct {
	RedirectURI string `json:"redirect_uri"`
	Intent      string `json:"intent"`
}

func (h *handlers) oauthStart(c *gin.Context) {
	logEnter(c, "oauth_start")
	if !h.requireSSO(c) {
		return
	}
	var body oauthStartBody
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
	middleware.WriteJSON(c, utils.StatusOK, OAuthStartOK{
		Meta:             metaFrom(c),
		AuthorizationURL: res.AuthorizationURL,
		State:            res.State,
	})
}

func (h *handlers) oauthCallback(c *gin.Context) {
	logEnter(c, "oauth_callback")
	if !h.requireSSO(c) {
		return
	}
	u, pair, err := h.sso.Callback(c.Request.Context(), c.Param("provider"), c.Query("code"), c.Query("state"), c.Query("error"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.writeOAuthLogin(c, u, pair, c.Param("provider"))
}

func (h *handlers) writeOAuthLogin(c *gin.Context, u *identity.User, pair *identity.TokenPair, provider string) {
	if pair == nil {
		pair = &identity.TokenPair{TokenType: "Bearer"}
	}
	if u != nil {
		c.Set(utils.CtxUserID, u.ID)
	}
	middleware.WriteJSON(c, utils.StatusOK, OAuthLoginOK{
		LoginOK: LoginOK{
			TokenOK: TokenOK{
				Meta:         metaFrom(c),
				AccessToken:  pair.AccessToken,
				RefreshToken: pair.RefreshToken,
				ExpiresIn:    pair.ExpiresIn,
				TokenType:    pair.TokenType,
			},
			User:         userView(u),
			Organization: h.primaryOrgView(c, u.ID),
		},
		LinkedProvider: provider,
	})
}

func (h *handlers) listIdentities(c *gin.Context) {
	logEnter(c, "list_identities")
	if !h.requireSSO(c) {
		return
	}
	items, err := h.sso.ListIdentities(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]IdentityView, 0, len(items))
	for i := range items {
		views = append(views, identityView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, IdentityListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: nil,
	})
}

type oauthLinkBody struct {
	AuthorizationCode string `json:"authorization_code"`
	State             string `json:"state"`
}

func (h *handlers) linkIdentity(c *gin.Context) {
	logEnter(c, "link_identity")
	if !h.requireSSO(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body oauthLinkBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	id, err := h.sso.Link(c.Request.Context(), middleware.UserIDFrom(c), c.Param("provider"), body.AuthorizationCode, body.State)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, IdentityOK{
		Meta:     metaFrom(c),
		Identity: identityView(id),
	})
}

func (h *handlers) unlinkIdentity(c *gin.Context) {
	logEnter(c, "unlink_identity")
	if !h.requireSSO(c) {
		return
	}
	provider := c.Param("provider")
	if err := h.sso.Unlink(c.Request.Context(), middleware.UserIDFrom(c), provider); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, UnlinkOK{
		Meta:     metaFrom(c),
		Status:   "unlinked",
		Provider: provider,
	})
}

type deviceCodeBody struct {
	ClientID string `json:"client_id"`
	Provider string `json:"provider"`
}

func (h *handlers) deviceCode(c *gin.Context) {
	logEnter(c, "device_code")
	if !h.requireSSO(c) {
		return
	}
	var body deviceCodeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	res, err := h.sso.CreateDeviceCode(c.Request.Context(), body.ClientID, body.Provider)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, DeviceCodeOK{
		Meta:                    metaFrom(c),
		DeviceCode:              res.DeviceCode,
		UserCode:                res.UserCode,
		VerificationURI:         res.VerificationURI,
		VerificationURIComplete: res.VerificationURIComplete,
		ExpiresIn:               res.ExpiresIn,
		Interval:                res.Interval,
	})
}

type deviceTokenBody struct {
	DeviceCode string `json:"device_code"`
}

func (h *handlers) deviceToken(c *gin.Context) {
	logEnter(c, "device_token")
	if !h.requireSSO(c) {
		return
	}
	var body deviceTokenBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, pair, err := h.sso.PollDeviceToken(c.Request.Context(), body.DeviceCode)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	c.Set(utils.CtxUserID, u.ID)
	middleware.WriteJSON(c, utils.StatusOK, LoginOK{
		TokenOK: TokenOK{
			Meta:         metaFrom(c),
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			ExpiresIn:    pair.ExpiresIn,
			TokenType:    pair.TokenType,
		},
		User:         userView(u),
		Organization: h.primaryOrgView(c, u.ID),
	})
}
