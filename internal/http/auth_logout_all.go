package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// logoutAll godoc
// @Summary Logout all sessions
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.LogoutAllOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/auth/logout-all [post]
func (h *handlers) logoutAll(c *gin.Context) {
	logEnter(c, "logout_all")
	if !h.requireIdent(c) {
		return
	}
	uid := middleware.UserIDFrom(c)
	n, err := h.ident.LogoutAll(c.Request.Context(), uid)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.cache.Delete(c.Request.Context(), cache.Me(uid))
	h.clearSessionCookies(c)
	middleware.WriteJSON(c, utils.StatusOK, dto.LogoutAllOK{
		Meta:            metaFrom(c),
		Status:          "logged_out_all",
		RevokedSessions: n,
	})
}
