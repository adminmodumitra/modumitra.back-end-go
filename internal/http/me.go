package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
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
	uid := middleware.UserIDFrom(c)
	key := cache.Me(uid)
	var cached dto.UserOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
		return
	}
	u, err := h.ident.Me(c.Request.Context(), uid)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	body := dto.UserOK{Meta: metaFrom(c), User: userView(u)}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLMe)
	middleware.WriteJSON(c, utils.StatusOK, body)
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
	uid := middleware.UserIDFrom(c)
	err := h.ident.ChangePassword(c.Request.Context(), uid, body.CurrentPassword, body.NewPassword)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.cache.Delete(c.Request.Context(), cache.Me(uid))
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{
		Meta:   metaFrom(c),
		Status: "password_changed",
	})
}
