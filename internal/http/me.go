package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) me(c *gin.Context) {
	logEnter(c, "me")
	if !h.requireIdent(c) {
		return
	}
	u, err := h.ident.Me(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.UserOK{
		Meta: metaFrom(c),
		User: userView(u),
	})
}

func (h *handlers) changePassword(c *gin.Context) {
	logEnter(c, "change_password")
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
	var body dto.ChangePasswordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	err := h.ident.ChangePassword(c.Request.Context(), middleware.UserIDFrom(c), body.CurrentPassword, body.NewPassword)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "password_changed",
	})
}
