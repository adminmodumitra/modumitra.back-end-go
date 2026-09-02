package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) transferOwnership(c *gin.Context) {
	logEnter(c, "transfer_ownership")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.TransferBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	actor := middleware.UserIDFrom(c)
	o, err := h.orgs.TransferOwnership(c.Request.Context(), actor, c.Param("id"), transferTarget(body))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.TransferOK{
		Meta:                metaFrom(c),
		Organization:        organizationView(o),
		PreviousOwnerUserID: actor,
		NewOwnerUserID:      o.OwnerUserID,
	})
}

func (h *handlers) orgRoles(c *gin.Context) {
	logEnter(c, "org_roles")
	if !h.requireOrgs(c) {
		return
	}
	roles, err := h.orgs.Roles(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.RoleCatalogOK{
		Meta:          metaFrom(c),
		SchemaVersion: 1,
		Roles:         roles,
	})
}
