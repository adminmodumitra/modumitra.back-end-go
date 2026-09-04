package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// getSeats godoc
// @Summary Get seat usage
// @Tags billing
// @Security BearerAuth
// @Produce json
// @Param id path string false "Organization ID (org-scoped route)"
// @Success 200 {object} dto.SeatsOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/seats [get]
// @Router /v1/organizations/{id}/seats [get]
func (h *handlers) getSeats(c *gin.Context) {
	logEnter(c, "get_seats")
	if !h.requireBilling(c) {
		return
	}
	seats, err := h.billing.Seats(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	members := make([]dto.SeatMemberView, 0, len(seats.Members))
	for _, m := range seats.Members {
		members = append(members, dto.SeatMemberView{
			UserID: m.UserID, Role: m.Role, Status: m.Status, ConsumesSeat: m.ConsumesSeat,
		})
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.SeatsOK{
		Meta:               metaFrom(c),
		PurchasedSeats:     seats.Purchased,
		UsedSeats:          seats.Used,
		PendingInvitations: seats.PendingInvitations,
		AvailableSeats:     seats.Available,
		Members:            members,
	})
}

// previewDowngrade godoc
// @Summary Preview plan downgrade
// @Tags billing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string false "Organization ID (org-scoped route)"
// @Param body body dto.PreviewBody true "Preview"
// @Success 200 {object} dto.PreviewDowngradeOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/seats/preview-downgrade [post]
// @Router /v1/organizations/{id}/seats/preview-downgrade [post]
func (h *handlers) previewDowngrade(c *gin.Context) {
	logEnter(c, "preview_downgrade")
	if !h.requireBilling(c) {
		return
	}
	var body dto.PreviewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	preview, err := h.billing.PreviewDowngrade(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c), body.PlanID, body.Seats)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	over := make([]dto.SeatMemberView, 0, len(preview.MembersOverTarget))
	for _, m := range preview.MembersOverTarget {
		over = append(over, dto.SeatMemberView{UserID: m.UserID, Role: m.Role, Status: m.Status, ConsumesSeat: m.ConsumesSeat})
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.PreviewDowngradeOK{
		Meta:              metaFrom(c),
		PlanID:            preview.PlanID,
		Plan:              preview.Plan,
		PurchasedSeats:    preview.PurchasedSeats,
		TargetSeats:       preview.TargetSeats,
		SeatShortfall:     preview.SeatShortfall,
		MembersOverTarget: over,
		CapabilityLoss:    preview.CapabilityLoss,
		Blocked:           preview.Blocked,
		Message:           preview.Message,
	})
}
