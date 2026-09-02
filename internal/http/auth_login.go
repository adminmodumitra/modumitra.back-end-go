package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) login(c *gin.Context) {
	logEnter(c, "login")
	if !h.requireIdent(c) {
		return
	}
	var body dto.LoginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, pair, err := h.ident.Login(c.Request.Context(), identity.LoginInput{
		Email:      body.Email,
		Password:   body.Password,
		DeviceName: body.DeviceName,
	})
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	c.Set(utils.CtxUserID, u.ID)
	middleware.WriteJSON(c, utils.StatusOK, dto.LoginOK{
		TokenOK: dto.TokenOK{
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

func (h *handlers) refresh(c *gin.Context) {
	logEnter(c, "refresh")
	if !h.requireIdent(c) {
		return
	}
	var body dto.RefreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	pair, err := h.ident.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.TokenOK{
		Meta:         metaFrom(c),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	})
}

func (h *handlers) logout(c *gin.Context) {
	logEnter(c, "logout")
	if !h.requireIdent(c) {
		return
	}
	var body dto.RefreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "logged_out",
	})
}

func (h *handlers) logoutAll(c *gin.Context) {
	logEnter(c, "logout_all")
	if !h.requireIdent(c) {
		return
	}
	n, err := h.ident.LogoutAll(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.LogoutAllOK{
		Meta:            metaFrom(c),
		Status:          "logged_out_all",
		RevokedSessions: n,
	})
}
