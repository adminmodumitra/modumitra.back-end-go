package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
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
	orgID := c.Param("id")
	o, err := h.orgs.TransferOwnership(c.Request.Context(), actor, orgID, transferTarget(body))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateOrgMembers(c, orgID, actor, o.OwnerUserID)
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
	orgID := c.Param("id")
	key := cache.OrgRoles(orgID)
	var cached dto.RoleCatalogOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
		return
	}
	roles, err := h.orgs.Roles(c.Request.Context(), middleware.UserIDFrom(c), orgID)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	body := dto.RoleCatalogOK{Meta: metaFrom(c), SchemaVersion: 1, Roles: roles}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLOrg)
	middleware.WriteJSON(c, utils.StatusOK, body)
}
