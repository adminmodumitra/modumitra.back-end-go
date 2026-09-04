package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// deviceCode godoc
// @Summary Start device authorization
// @Tags sso
// @Accept json
// @Produce json
// @Param body body dto.DeviceCodeBody true "Device code request"
// @Success 200 {object} dto.DeviceCodeOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/device/code [post]
func (h *handlers) deviceCode(c *gin.Context) {
	logEnter(c, "device_code")
	if !h.requireSSO(c) {
		return
	}
	var body dto.DeviceCodeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	res, err := h.sso.CreateDeviceCode(c.Request.Context(), body.ClientID, body.Provider)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.DeviceCodeOK{
		Meta:                    metaFrom(c),
		DeviceCode:              res.DeviceCode,
		UserCode:                res.UserCode,
		VerificationURI:         res.VerificationURI,
		VerificationURIComplete: res.VerificationURIComplete,
		ExpiresIn:               res.ExpiresIn,
		Interval:                res.Interval,
	})
}

// deviceToken godoc
// @Summary Poll device authorization token
// @Tags sso
// @Accept json
// @Produce json
// @Param body body dto.DeviceTokenBody true "Device code"
// @Success 200 {object} dto.LoginOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/device/token [post]
func (h *handlers) deviceToken(c *gin.Context) {
	logEnter(c, "device_token")
	if !h.requireSSO(c) {
		return
	}
	var body dto.DeviceTokenBody
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
