package httpapi

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// deviceApprove godoc
// @Summary Approve a desktop sign-in code as the signed-in user
// @Tags sso
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.DeviceApproveBody true "User code shown by the desktop app"
// @Success 200 {object} dto.StatusOK
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/auth/device/approve [post]
func (h *handlers) deviceApprove(c *gin.Context) {
	logEnter(c, "device_approve")
	if !h.requireSSO(c) {
		return
	}
	var body dto.DeviceApproveBody
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.UserCode) == "" {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	code := strings.ToUpper(strings.TrimSpace(body.UserCode))
	if err := h.sso.AuthorizeDevice(c.Request.Context(), code, middleware.UserIDFrom(c)); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.StatusOK{Meta: metaFrom(c), Status: "approved"})
}
