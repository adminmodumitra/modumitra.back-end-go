package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// createInvitation godoc
// @Summary Create invitation
// @Tags invitations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string true "Organization ID"
// @Param body body dto.CreateInviteBody true "Invite"
// @Success 201 {object} dto.InvitationOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/invitations [post]
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

// listInvitations godoc
// @Summary List invitations
// @Tags invitations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Param status query string false "Status filter"
// @Param cursor query string false "Cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.InvitationListOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/invitations [get]
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

// revokeInvitation godoc
// @Summary Revoke invitation
// @Tags invitations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Param invitationId path string true "Invitation ID"
// @Success 200 {object} dto.InvitationRevokedOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/organizations/{id}/invitations/{invitationId} [delete]
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
