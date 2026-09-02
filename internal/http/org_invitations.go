package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) createInvitation(c *gin.Context) {
	logEnter(c, "create_invitation")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.CreateInviteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	inv, err := h.orgs.CreateInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), body.Email, body.Role)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusCreated, dto.InvitationOK{
		Meta:       metaFrom(c),
		Invitation: invitationView(inv),
	})
}

func (h *handlers) listInvitations(c *gin.Context) {
	logEnter(c, "list_invitations")
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
	out := make([]dto.InvitationView, 0, len(items))
	for i := range items {
		inv := items[i]
		out = append(out, invitationView(&inv))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InvitationListOK{
		Meta:       metaFrom(c),
		Items:      out,
		NextCursor: cursorPtr(next),
	})
}

func (h *handlers) revokeInvitation(c *gin.Context) {
	logEnter(c, "revoke_invitation")
	if !h.requireOrgs(c) {
		return
	}
	id := c.Param("invitationId")
	if err := h.orgs.RevokeInvitation(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), id); err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InvitationRevokedOK{
		Meta:   metaFrom(c),
		Status: "revoked",
		ID:     id,
	})
}

func (h *handlers) resendInvitation(c *gin.Context) {
	logEnter(c, "resend_invitation")
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
	middleware.WriteJSON(c, utils.StatusAccepted, dto.InvitationResentOK{
		Meta:      metaFrom(c),
		Status:    "resent",
		ID:        inv.ID,
		ExpiresAt: rfcTime(inv.ExpiresAt),
	})
}

func (h *handlers) previewInvitation(c *gin.Context) {
	logEnter(c, "preview_invitation")
	if !h.requireOrgs(c) {
		return
	}
	prev, err := h.orgs.PreviewInvitation(c.Request.Context(), c.Param("token"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InvitePreviewOK{
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
	logEnter(c, "accept_invitation")
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
	middleware.WriteJSON(c, utils.StatusOK, dto.InviteAcceptOK{
		Meta:         metaFrom(c),
		Membership:   membershipView(m, nil),
		Organization: organizationView(o),
	})
}
