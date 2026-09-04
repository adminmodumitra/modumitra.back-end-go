package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// listMembers godoc
// @Summary List organization members
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Param status query string false "Status filter"
// @Param role query string false "Role filter"
// @Param cursor query string false "Cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.MemberListOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/members [get]
func (h *handlers) listMembers(c *gin.Context) {
	logEnter(c, "list_members")
	if !h.requireOrgs(c) {
		return
	}
	limit, err := org.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	orgID := c.Param("id")
	status, role, cursor := c.Query("status"), c.Query("role"), c.Query("cursor")
	key := cache.OrgMembers(orgID, status, role, cursor, limit)
	var cached dto.MemberListOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
		return
	}
	items, next, err := h.orgs.ListMembers(c.Request.Context(), middleware.UserIDFrom(c), orgID, status, role, cursor, limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	out := make([]dto.MembershipView, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, membershipView(&m, h.memberUserView(c, m.UserID)))
	}
	body := dto.MemberListOK{Meta: metaFrom(c), Items: out, NextCursor: cursorPtr(next)}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLOrg)
	middleware.WriteJSON(c, utils.StatusOK, body)
}

// patchMember godoc
// @Summary Update member role/status
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Param body body dto.PatchMemberBody true "Patch"
// @Success 200 {object} dto.MemberOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/members/{userId} [patch]
func (h *handlers) patchMember(c *gin.Context) {
	logEnter(c, "patch_member")
	if !h.requireOrgs(c) {
		return
	}
	var body dto.PatchMemberBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	orgID, uid := c.Param("id"), c.Param("userId")
	m, err := h.orgs.PatchMember(c.Request.Context(), middleware.UserIDFrom(c), orgID, uid, body.Role, body.Status)
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
