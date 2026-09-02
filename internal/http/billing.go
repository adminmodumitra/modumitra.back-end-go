package httpapi

import (
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) billingTenant(c *gin.Context) string {
	if id := c.Param("id"); id != "" {
		return id
	}
	return strings.TrimSpace(c.GetHeader(utils.HeaderOrganization))
}

func (h *handlers) requireBilling(c *gin.Context) bool {
	if h.billing == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

func (h *handlers) listPlans(c *gin.Context) {
	logEnter(c, "list_plans")
	if !h.requireBilling(c) {
		return
	}
	items, err := h.billing.ListPlans(c.Request.Context())
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	views := make([]dto.PlanView, 0, len(items))
	for i := range items {
		views = append(views, planView(&items[i]))
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.PlanListOK{
		Meta:       metaFrom(c),
		Items:      views,
		NextCursor: nil,
	})
}

func (h *handlers) getSubscription(c *gin.Context) {
	logEnter(c, "get_subscription")
	if !h.requireBilling(c) {
		return
	}
	sub, err := h.billing.GetSubscription(c.Request.Context(), middleware.UserIDFrom(c), h.billingTenant(c))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

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
