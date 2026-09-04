package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

var (
	_ = dto.OAuthLoginOK{}
	_ = utils.ErrorResponse{}
)

// oauthCallback godoc
// @Summary OAuth provider callback
// @Tags sso
// @Produce json
// @Param provider path string true "Provider id"
// @Param code query string false "Authorization code"
// @Param state query string false "OAuth state"
// @Param error query string false "Provider error"
// @Success 200 {object} dto.OAuthLoginOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/auth/oauth/{provider}/callback [get]
func (h *handlers) oauthCallback(c *gin.Context) {
	logEnter(c, "oauth_callback")
	if !h.requireSSO(c) {
		return
	}
	u, pair, err := h.sso.Callback(c.Request.Context(), c.Param("provider"), c.Query("code"), c.Query("state"), c.Query("error"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.writeOAuthLogin(c, u, pair, c.Param("provider"))
}
