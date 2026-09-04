package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// patchOrg godoc
// @Summary Update organization
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param body body dto.PatchOrgBody true "Patch"
// @Success 200 {object} dto.OrgOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id} [patch]
func (h *handlers) patchOrg(c *gin.Context) {
	logEnter(c, "patch_org")
	if !h.requireOrgs(c) {
		return
	}
	var body dto.PatchOrgBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	o, err := h.orgs.UpdateName(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.Name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
	})
}

// deleteOrg godoc
// @Summary Delete organization
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string true "Organization ID"
// @Success 200 {object} dto.OrgDeletedOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id} [delete]
func (h *handlers) deleteOrg(c *gin.Context) {
	logEnter(c, "delete_org")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	id := c.Param("id")
	if err := h.orgs.Delete(c.Request.Context(), middleware.UserIDFrom(c), id); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgDeletedOK{
		Meta:   metaFrom(c),
		Status: "deleted",
		ID:     id,
	})
}

// leaveOrg godoc
// @Summary Leave organization
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string true "Organization ID"
// @Success 200 {object} dto.LeaveOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/leave [post]
func (h *handlers) leaveOrg(c *gin.Context) {
	logEnter(c, "leave_org")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	id := c.Param("id")
	uid := middleware.UserIDFrom(c)
	if err := h.orgs.Leave(c.Request.Context(), uid, id); err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateOrgMembers(c, id, uid)
	middleware.WriteJSON(c, utils.StatusOK, dto.LeaveOK{
		Meta:           metaFrom(c),
		Status:         "left",
		OrganizationID: id,
	})
}
