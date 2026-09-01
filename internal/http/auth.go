package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type registerBody struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	DisplayName    string `json:"display_name"`
	TermsVersion   string `json:"terms_version"`
	PrivacyVersion string `json:"privacy_version"`
}

type tokenBody struct {
	Token string `json:"token"`
}

type emailBody struct {
	Email string `json:"email"`
}

type loginBody struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceName string `json:"device_name"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

type resetConfirmBody struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *handlers) requireIdent(c *gin.Context) bool {
	if h.ident == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

func (h *handlers) register(c *gin.Context) {
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
	var body registerBody
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
	middleware.WriteJSON(c, utils.StatusCreated, UserOK{
		Meta: metaFrom(c),
		User: userView(u),
	})
}

func (h *handlers) verifyEmail(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body tokenBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, err := h.ident.VerifyEmail(c.Request.Context(), body.Token)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, UserOK{
		Meta: metaFrom(c),
		User: userView(u),
	})
}

func (h *handlers) resendVerification(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body emailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.ResendVerification(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}

func (h *handlers) login(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body loginBody
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
	middleware.WriteJSON(c, utils.StatusOK, LoginOK{
		TokenOK: TokenOK{
			Meta:         metaFrom(c),
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			ExpiresIn:    pair.ExpiresIn,
			TokenType:    pair.TokenType,
		},
		User:         userView(u),
		Organization: nil,
	})
}

func (h *handlers) refresh(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	pair, err := h.ident.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, TokenOK{
		Meta:         metaFrom(c),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	})
}

func (h *handlers) logout(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, StatusOK{
		Meta:   metaFrom(c),
		Status: "logged_out",
	})
}

func (h *handlers) logoutAll(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	n, err := h.ident.LogoutAll(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, LogoutAllOK{
		Meta:            metaFrom(c),
		Status:          "logged_out_all",
		RevokedSessions: n,
	})
}

func (h *handlers) passwordReset(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body emailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.RequestPasswordReset(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}

func (h *handlers) passwordResetConfirm(c *gin.Context) {
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
	var body resetConfirmBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.ConfirmPasswordReset(c.Request.Context(), body.Token, body.NewPassword); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, StatusOK{
		Meta:   metaFrom(c),
		Status: "password_reset",
	})
}

type otpVerifyBody struct {
	Email      string `json:"email"`
	OTP        string `json:"otp"`
	DeviceName string `json:"device_name"`
}

func (h *handlers) otpRequest(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body emailBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	if err := h.ident.RequestOTP(c.Request.Context(), body.Email); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, StatusOK{
		Meta:   metaFrom(c),
		Status: "accepted",
	})
}

func (h *handlers) otpVerify(c *gin.Context) {
	if !h.requireIdent(c) {
		return
	}
	var body otpVerifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	u, pair, err := h.ident.VerifyOTP(c.Request.Context(), body.Email, body.OTP, body.DeviceName)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, LoginOK{
		TokenOK: TokenOK{
			Meta:         metaFrom(c),
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			ExpiresIn:    pair.ExpiresIn,
			TokenType:    pair.TokenType,
		},
		User:         userView(u),
		Organization: nil,
	})
}
