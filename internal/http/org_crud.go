package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) createOrg(c *gin.Context) {
	logEnter(c, "create_org")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.CreateOrgBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	o, m, err := h.orgs.Create(c.Request.Context(), middleware.UserIDFrom(c), body.Name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	trial := ""
	if o.TrialEndsAt != nil {
		trial = rfcTime(*o.TrialEndsAt)
	}
	middleware.WriteJSON(c, utils.StatusCreated, dto.OrgCreatedOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
		Membership:   membershipView(m, nil),
		TrialEndsAt:  trial,
	})
}

func (h *handlers) listOrgs(c *gin.Context) {
	logEnter(c, "list_orgs")
	if !h.requireOrgs(c) {
		return
	}
	limit, err := org.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, err := h.orgs.List(c.Request.Context(), middleware.UserIDFrom(c), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	out := make([]dto.OrgListItem, 0, len(items))
	for _, it := range items {
		o := it.Organization
		out = append(out, dto.OrgListItem{
			Organization:     organizationView(&o),
			Role:             it.Membership.Role,
			MembershipStatus: it.Membership.Status,
		})
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) getOrg(c *gin.Context) {
	logEnter(c, "get_org")
	if !h.requireOrgs(c) {
		return
	}
	o, err := h.orgs.Get(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.OrgOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
	})
}

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

func (h *handlers) leaveOrg(c *gin.Context) {
	logEnter(c, "leave_org")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	id := c.Param("id")
	if err := h.orgs.Leave(c.Request.Context(), middleware.UserIDFrom(c), id); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.LeaveOK{
		Meta:           metaFrom(c),
		Status:         "left",
		OrganizationID: id,
	})
}
