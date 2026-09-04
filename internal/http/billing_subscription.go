package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// cancelSubscription godoc
// @Summary Cancel subscription
// @Tags billing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string false "Organization ID (org-scoped route)"
// @Param body body dto.CancelBody true "Cancel options"
// @Success 200 {object} dto.SubscriptionOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/subscription/cancel [post]
// @Router /v1/organizations/{id}/subscription/cancel [post]
func (h *handlers) cancelSubscription(c *gin.Context) {
	logEnter(c, "cancel_subscription")
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.CancelBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	atEnd := body.AtPeriodEnd != nil && *body.AtPeriodEnd
	uid := middleware.UserIDFrom(c)
	orgID := h.billingTenant(c)
	sub, err := h.billing.Cancel(c.Request.Context(), uid, orgID, atEnd)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateBillingSub(c, uid, orgID)
	middleware.WriteJSON(c, utils.StatusOK, dto.SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

// resumeSubscription godoc
// @Summary Resume subscription
// @Tags billing
// @Security BearerAuth
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string false "Organization ID (org-scoped route)"
// @Success 200 {object} dto.SubscriptionOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/subscription/resume [post]
// @Router /v1/organizations/{id}/subscription/resume [post]
func (h *handlers) resumeSubscription(c *gin.Context) {
	logEnter(c, "resume_subscription")
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	uid := middleware.UserIDFrom(c)
	orgID := h.billingTenant(c)
	sub, err := h.billing.Resume(c.Request.Context(), uid, orgID)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateBillingSub(c, uid, orgID)
	middleware.WriteJSON(c, utils.StatusOK, dto.SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}
