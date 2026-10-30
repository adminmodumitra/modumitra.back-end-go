package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// login godoc
// @Summary Login with email/password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginBody true "Credentials"
// @Success 200 {object} dto.LoginOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/login [post]
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
		TokenOK:      h.tokenView(c, pair),
		User:         userView(u),
		Organization: h.primaryOrgView(c, u.ID),
	})
}

// refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshBody true "Refresh token"
// @Success 200 {object} dto.TokenOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/auth/refresh [post]
func (h *handlers) refresh(c *gin.Context) {
	logEnter(c, "refresh")
	if !h.requireIdent(c) {
		return
	}
	var body dto.RefreshBody
	_ = c.ShouldBindJSON(&body) // an empty body is fine with a refresh cookie
	raw, err := h.refreshTokenFrom(c, body.RefreshToken)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	pair, err := h.ident.Refresh(c.Request.Context(), raw)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, h.tokenView(c, pair))
}

// logout godoc
// @Summary Logout current session
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshBody true "Refresh token"
// @Success 200 {object} dto.StatusOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/logout [post]
func (h *handlers) logout(c *gin.Context) {
	logEnter(c, "logout")
	if !h.requireIdent(c) {
		return
	}
	var body dto.RefreshBody
	_ = c.ShouldBindJSON(&body)
	raw, err := h.refreshTokenFrom(c, body.RefreshToken)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	if err := h.ident.Logout(c.Request.Context(), raw); err != nil {
		middleware.Abort(c, err)
		return
	}
	h.clearSessionCookies(c)
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "logged_out",
	})
}
