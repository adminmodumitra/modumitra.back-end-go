package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// checkoutSession godoc
// @Summary Create checkout session
// @Tags billing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string false "Organization ID (org-scoped route)"
// @Param body body dto.CheckoutBody true "Checkout"
// @Success 200 {object} dto.CheckoutOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/checkout-session [post]
// @Router /v1/organizations/{id}/checkout-session [post]
func (h *handlers) checkoutSession(c *gin.Context) {
	logEnter(c, "checkout_session")
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.CheckoutBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	sess, err := h.billing.Checkout(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c), body.PlanID, body.SuccessURL, body.CancelURL)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.CheckoutOK{
		Meta:              metaFrom(c),
		CheckoutSessionID: sess.ID,
		URL:               sess.URL,
		ExpiresAt:         rfcTime(sess.ExpiresAt),
	})
}

// portalSession godoc
// @Summary Create billing portal session
// @Tags billing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "Idempotency key"
// @Param id path string false "Organization ID (org-scoped route)"
// @Param body body dto.PortalBody true "Portal"
// @Success 200 {object} dto.PortalOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/portal-session [post]
// @Router /v1/organizations/{id}/portal-session [post]
func (h *handlers) portalSession(c *gin.Context) {
	logEnter(c, "portal_session")
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	var body dto.PortalBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Abort(c, utils.E(utils.CodeValidation))
		return
	}
	sess, err := h.billing.Portal(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c), body.ReturnURL)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.PortalOK{
		Meta:            metaFrom(c),
		PortalSessionID: sess.ID,
		URL:             sess.URL,
		ExpiresAt:       rfcTime(sess.ExpiresAt),
	})
}
