package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) passwordReset(c *gin.Context) {
	logEnter(c, "password_reset")
	if !h.requireIdent(c) {
		return
	}
	var body dto.EmailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.RequestPasswordReset(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}

func (h *handlers) passwordResetConfirm(c *gin.Context) {
	logEnter(c, "password_reset_confirm")
	if !h.requireIdent(c) {
		return
	}
	if c.GetHeader(utils.HeaderIdempotency) == "" {
		middleware.Abort(c, utils.E(utils.CodeValidation).WithDetails(map[string]any{
			"field":  "Idempotency-Key",
			"reason": "required",
		}))
		return
	}
	var body dto.ResetConfirmBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.ConfirmPasswordReset(c.Request.Context(), body.Token, body.NewPassword); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "password_reset",
	})
}

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
