package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// suspendMember godoc
// @Summary Suspend organization member
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Success 200 {object} dto.MemberOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/members/{userId}/suspend [post]
func (h *handlers) suspendMember(c *gin.Context) {
	logEnter(c, "suspend_member")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	orgID, uid := c.Param("id"), c.Param("userId")
	m, err := h.orgs.SuspendMember(c.Request.Context(), middleware.UserIDFrom(c), orgID, uid)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateOrgMembers(c, orgID, uid)
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberOK{
		Meta:       metaFrom(c),
		Membership: membershipView(m, h.memberUserView(c, m.UserID)),
	})
}

// removeMember godoc
// @Summary Remove organization member
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Success 200 {object} dto.MemberRemovedOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/members/{userId} [delete]
func (h *handlers) removeMember(c *gin.Context) {
	logEnter(c, "remove_member")
	if !h.requireOrgs(c) {
		return
	}
	id, uid := c.Param("id"), c.Param("userId")
	if err := h.orgs.RemoveMember(c.Request.Context(), middleware.UserIDFrom(c), id, uid); err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateOrgMembers(c, id, uid)
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberRemovedOK{
		Meta:           metaFrom(c),
		Status:         "removed",
		UserID:         uid,
		OrganizationID: id,
	})
}
