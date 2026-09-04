package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// listIdentities godoc
// @Summary List linked identities
// @Tags me
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.IdentityListOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/me/identities [get]
func (h *handlers) listIdentities(c *gin.Context) {
	logEnter(c, "list_identities")
	if !h.requireSSO(c) {
		return
	}
	items, err := h.sso.ListIdentities(c.Request.Context(), middleware.UserIDFrom(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]dto.IdentityView, 0, len(items))
	for i := range items {
		views = append(views, identityView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.IdentityListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: nil,
	})
}

// linkIdentity godoc
// @Summary Link OAuth identity
// @Tags me
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param provider path string true "Provider id"
// @Param body body dto.OAuthLinkBody true "Link"
// @Success 200 {object} dto.IdentityOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/me/identities/{provider}/link [post]
func (h *handlers) linkIdentity(c *gin.Context) {
	logEnter(c, "link_identity")
	if !h.requireSSO(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.OAuthLinkBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	id, err := h.sso.Link(c.Request.Context(), middleware.UserIDFrom(c), c.Param("provider"), body.AuthorizationCode, body.State)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.IdentityOK{
		Meta:     metaFrom(c),
		Identity: identityView(id),
	})
}

// unlinkIdentity godoc
// @Summary Unlink OAuth identity
// @Tags me
// @Security BearerAuth
// @Produce json
// @Param provider path string true "Provider id"
// @Success 200 {object} dto.UnlinkOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/me/identities/{provider} [delete]
func (h *handlers) unlinkIdentity(c *gin.Context) {
	logEnter(c, "unlink_identity")
	if !h.requireSSO(c) {
		return
	}
	provider := c.Param("provider")
	if err := h.sso.Unlink(c.Request.Context(), middleware.UserIDFrom(c), provider); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.UnlinkOK{
		Meta:     metaFrom(c),
		Status:   "unlinked",
		Provider: provider,
	})
}
