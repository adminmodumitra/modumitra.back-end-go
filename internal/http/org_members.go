package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
	items, next, err := h.orgs.ListMembers(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Query("status"), c.Query("role"), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	out := make([]dto.MembershipView, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, membershipView(&m, h.memberUserView(c, m.UserID)))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

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
	m, err := h.orgs.PatchMember(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Param("userId"), body.Role, body.Status)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberOK{
		Meta:       metaFrom(c),
		Membership: membershipView(m, h.memberUserView(c, m.UserID)),
	})
}

func (h *handlers) suspendMember(c *gin.Context) {
	logEnter(c, "suspend_member")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	m, err := h.orgs.SuspendMember(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Param("userId"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberOK{
		Meta:       metaFrom(c),
		Membership: membershipView(m, h.memberUserView(c, m.UserID)),
	})
}

func (h *handlers) removeMember(c *gin.Context) {
	logEnter(c, "remove_member")
	if !h.requireOrgs(c) {
		return
	}
	id := c.Param("id")
	uid := c.Param("userId")
	if err := h.orgs.RemoveMember(c.Request.Context(), middleware.UserIDFrom(c), id, uid); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.MemberRemovedOK{
		Meta:           metaFrom(c),
		Status:         "removed",
		UserID:         uid,
		OrganizationID: id,
	})
}
