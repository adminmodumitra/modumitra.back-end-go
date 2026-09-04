package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// passwordReset godoc
// @Summary Request password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.EmailBody true "Email"
// @Success 202 {object} dto.StatusOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/password-reset [post]
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

// passwordResetConfirm godoc
// @Summary Confirm password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param body body dto.ResetConfirmBody true "Token and new password"
// @Success 200 {object} dto.StatusOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/password-reset/confirm [post]
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
