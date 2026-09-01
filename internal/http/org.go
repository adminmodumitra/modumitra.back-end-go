package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireOrgs(c *gin.Context) bool {
	if h.orgs == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

func requireIdempotencyKey(c *gin.Context) bool {
	if c.GetHeader(utils.HeaderIdempotency) == "" {
		middleware.Abort(c, utils.E(utils.CodeValidation).WithDetails(map[string]any{
			"field":  "Idempotency-Key",
			"reason": "required",
		}))
		return false
	}
	return true
}

type createOrgBody struct {
	Name string `json:"name"`
}

type patchOrgBody struct {
	Name string `json:"name"`
}

type transferBody struct {
	UserID         string `json:"user_id"`
	NewOwnerUserID string `json:"new_owner_user_id"`
}

func (b transferBody) target() string {
	if b.UserID != "" {
		return b.UserID
	}
	return b.NewOwnerUserID
}

type patchMemberBody struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}

type createInviteBody struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *handlers) memberUserView(c *gin.Context, userID string) *MemberUserView {
	if h.orgs == nil {
		return nil
	}
	u, err := h.orgs.MemberUser(c.Request.Context(), userID)
	if err != nil || u == nil {
		return nil
	}
	return &MemberUserView{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		EmailVerified: u.EmailVerified(),
	}
}

func (h *handlers) primaryOrgView(c *gin.Context, userID string) *OrganizationView {
	if h.orgs == nil {
		return nil
	}
	o, err := h.orgs.PrimaryForUser(c.Request.Context(), userID)
	if err != nil || o == nil {
		return nil
	}
	v := organizationView(o)
	return &v
}

func (h *handlers) createOrg(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body createOrgBody
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
	middleware.WriteJSON(c, utils.StatusCreated, OrgCreatedOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
		Membership:   membershipView(m, nil),
		TrialEndsAt:  trial,
	})
}

func (h *handlers) listOrgs(c *gin.Context) {
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
	out := make([]OrgListItem, 0, len(items))
	for _, it := range items {
		o := it.Organization
		out = append(out, OrgListItem{
			Organization:     organizationView(&o),
			Role:             it.Membership.Role,
			MembershipStatus: it.Membership.Status,
		})
	}
	middleware.WriteJSON(c, utils.StatusOK, OrgListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) getOrg(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	o, err := h.orgs.Get(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, OrgOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
	})
}

func (h *handlers) patchOrg(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	var body patchOrgBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	o, err := h.orgs.UpdateName(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.Name)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, OrgOK{
		Meta:         metaFrom(c),
		Organization: organizationView(o),
	})
}

func (h *handlers) deleteOrg(c *gin.Context) {
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
	middleware.WriteJSON(c, utils.StatusOK, OrgDeletedOK{
		Meta:   metaFrom(c),
		Status: "deleted",
		ID:     id,
	})
}

func (h *handlers) leaveOrg(c *gin.Context) {
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
	middleware.WriteJSON(c, utils.StatusOK, LeaveOK{
		Meta:           metaFrom(c),
		Status:         "left",
		OrganizationID: id,
	})
}

func (h *handlers) transferOwnership(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body transferBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	actor := middleware.UserIDFrom(c)
	o, err := h.orgs.TransferOwnership(c.Request.Context(), actor, c.Param("id"), body.target())
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, TransferOK{
		Meta:                metaFrom(c),
		Organization:        organizationView(o),
		PreviousOwnerUserID: actor,
		NewOwnerUserID:      o.OwnerUserID,
	})
}

func (h *handlers) orgRoles(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	roles, err := h.orgs.Roles(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, RoleCatalogOK{
		Meta:          metaFrom(c),
		SchemaVersion: 1,
		Roles:         roles,
	})
}

func (h *handlers) listMembers(c *gin.Context) {
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
	out := make([]MembershipView, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, membershipView(&m, h.memberUserView(c, m.UserID)))
	}
	middleware.WriteJSON(c, utils.StatusOK, MemberListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) patchMember(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	var body patchMemberBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	m, err := h.orgs.PatchMember(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Param("userId"), body.Role, body.Status)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, MemberOK{
		Meta:       metaFrom(c),
		Membership: membershipView(m, h.memberUserView(c, m.UserID)),
	})
}

func (h *handlers) suspendMember(c *gin.Context) {
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
	middleware.WriteJSON(c, utils.StatusOK, MemberOK{
		Meta:       metaFrom(c),
		Membership: membershipView(m, h.memberUserView(c, m.UserID)),
	})
}

func (h *handlers) removeMember(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	id := c.Param("id")
	uid := c.Param("userId")
	if err := h.orgs.RemoveMember(c.Request.Context(), middleware.UserIDFrom(c), id, uid); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, MemberRemovedOK{
		Meta:           metaFrom(c),
		Status:         "removed",
		UserID:         uid,
		OrganizationID: id,
	})
}

func (h *handlers) createInvitation(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body createInviteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	inv, err := h.orgs.CreateInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.Email, body.Role)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusCreated, InvitationOK{
		Meta:       metaFrom(c),
		Invitation: invitationView(inv),
	})
}

func (h *handlers) listInvitations(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	limit, err := org.ParseLimit(c.Query("limit"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	items, next, err := h.orgs.ListInvitations(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Query("status"), c.Query("cursor"), limit)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	out := make([]InvitationView, 0, len(items))
	for i := range items {
		inv := items[i]
		out = append(out, invitationView(&inv))
	}
	middleware.WriteJSON(c, utils.StatusOK, InvitationListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) revokeInvitation(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	id := c.Param("invitationId")
	if err := h.orgs.RevokeInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), id); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, InvitationRevokedOK{
		Meta:   metaFrom(c),
		Status: "revoked",
		ID:     id,
	})
}

func (h *handlers) resendInvitation(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	inv, err := h.orgs.ResendInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), c.Param("invitationId"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusAccepted, InvitationResentOK{
		Meta:      metaFrom(c),
		Status:    "resent",
		ID:        inv.ID,
		ExpiresAt: rfcTime(inv.ExpiresAt),
	})
}

func (h *handlers) previewInvitation(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	prev, err := h.orgs.PreviewInvitation(c.Request.Context(), c.Param("token"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, InvitePreviewOK{
		Meta:               metaFrom(c),
		OrganizationName:   prev.OrganizationName,
		Role:               prev.Role,
		InviterDisplayName: prev.InviterDisplayName,
		Email:              prev.Email,
		ExpiresAt:          rfcTime(prev.ExpiresAt),
		Status:             prev.Status,
	})
}

func (h *handlers) acceptInvitation(c *gin.Context) {
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	o, m, err := h.orgs.AcceptInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("token"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, InviteAcceptOK{
		Meta:         metaFrom(c),
		Membership:   membershipView(m, nil),
		Organization: organizationView(o),
	})
}
