package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireIdent(c *gin.Context) bool {
	if h.ident == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

// register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param body body dto.RegisterBody true "Registration"
// @Success 201 {object} dto.UserOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/register [post]
func (h *handlers) register(c *gin.Context) {
	logEnter(c, "register")
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
	var body dto.RegisterBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, err := h.ident.Register(c.Request.Context(), identity.RegisterInput{
		Email:          body.Email,
		Password:       body.Password,
		DisplayName:    body.DisplayName,
		TermsVersion:   body.TermsVersion,
		PrivacyVersion: body.PrivacyVersion,
	})
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	c.Set(utils.CtxUserID, u.ID)
	middleware.WriteJSON(c, utils.StatusCreated, dto.UserOK{
		Meta: metaFrom(c),
		User: userView(u),
	})
}

// verifyEmail godoc
// @Summary Verify email with token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.TokenBody true "Verification token"
// @Success 200 {object} dto.UserOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/verify-email [post]
func (h *handlers) verifyEmail(c *gin.Context) {
	logEnter(c, "verify_email")
	if !h.requireIdent(c) {
		return
	}
	var body dto.TokenBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, err := h.ident.VerifyEmail(c.Request.Context(), body.Token)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	c.Set(utils.CtxUserID, u.ID)
	middleware.WriteJSON(c, utils.StatusOK, dto.UserOK{
		Meta: metaFrom(c),
		User: userView(u),
	})
}

// resendVerification godoc
// @Summary Resend email verification
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.EmailBody true "Email"
// @Success 202 {object} dto.StatusOK
// @Failure 422 {object} utils.ErrorResponse
// @Router /v1/auth/resend-verification [post]
func (h *handlers) resendVerification(c *gin.Context) {
	logEnter(c, "resend_verification")
	if !h.requireIdent(c) {
		return
	}
	var body dto.EmailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.ResendVerification(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}
