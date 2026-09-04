package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// resendInvitation godoc
// @Summary Resend invitation
// @Tags invitations
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string true "Organization ID"
// @Param invitationId path string true "Invitation ID"
// @Success 202 {object} dto.InvitationResentOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/invitations/{invitationId}/resend [post]
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

// previewInvitation godoc
// @Summary Preview invitation by token
// @Tags invitations
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} dto.InvitePreviewOK
// @Failure 404 {object} utils.ErrorResponse
// @Router /v1/invitations/{token} [get]
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

// acceptInvitation godoc
// @Summary Accept invitation
// @Tags invitations
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param token path string true "Invitation token"
// @Success 200 {object} dto.InviteAcceptOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/invitations/{token}/accept [post]
func (h *handlers) acceptInvitation(c *gin.Context) {
	logEnter(c, "accept_invitation")
	if !h.requireOrgs(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	uid := middleware.UserIDFrom(c)
	o, m, err := h.orgs.AcceptInvitation(c.Request.Context(), uid, c.Param("token"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	if o != nil {
		h.invalidateOrgMembers(c, o.ID, uid)
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.InviteAcceptOK{
		Meta:         metaFrom(c),
		Membership:   membershipView(m, nil),
		Organization: organizationView(o),
	})
}
