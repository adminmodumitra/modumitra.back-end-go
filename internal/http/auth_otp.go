package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// otpRequest godoc
// @Summary Request login OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.EmailBody true "Email"
// @Success 202 {object} dto.StatusOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/otp/request [post]
// @Router /v1/auth/otp/resend [post]
func (h *handlers) otpRequest(c *gin.Context) {
	logEnter(c, "otp_request")
	if !h.requireIdent(c) {
		return
	}
	var body dto.EmailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.RequestOTP(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}

// otpVerify godoc
// @Summary Verify login OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.OTPVerifyBody true "OTP"
// @Success 200 {object} dto.LoginOK
// @Failure 401 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/otp/verify [post]
func (h *handlers) otpVerify(c *gin.Context) {
	logEnter(c, "otp_verify")
	if !h.requireIdent(c) {
		return
	}
	var body dto.OTPVerifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, pair, err := h.ident.VerifyOTP(c.Request.Context(), body.Email, body.OTP, body.DeviceName)
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
