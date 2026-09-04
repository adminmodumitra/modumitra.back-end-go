package httpapi

import (
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
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

// listPlans godoc
// @Summary List billing plans
// @Tags billing
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.PlanListOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/plans [get]
func (h *handlers) listPlans(c *gin.Context) {
	logEnter(c, "list_plans")
	if !h.requireBilling(c) {
		return
	}
	key := cache.BillingPlans()
	var cached dto.PlanListOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
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
	body := dto.PlanListOK{Meta: metaFrom(c), Items: views, NextCursor: nil}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLBilling)
	middleware.WriteJSON(c, utils.StatusOK, body)
}

// getSubscription godoc
// @Summary Get subscription
// @Tags billing
// @Security BearerAuth
// @Produce json
// @Param id path string false "Organization ID (org-scoped route)"
// @Param X-Organization-Id header string false "Organization ID (personal route)"
// @Success 200 {object} dto.SubscriptionOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/billing/subscription [get]
// @Router /v1/organizations/{id}/subscription [get]
func (h *handlers) getSubscription(c *gin.Context) {
	logEnter(c, "get_subscription")
	if !h.requireBilling(c) {
		return
	}
	uid := middleware.UserIDFrom(c)
	orgID := h.billingTenant(c)
	key := cache.BillingSubOrg(orgID)
	if orgID == "" {
		key = cache.BillingSubUser(uid)
	}
	var cached dto.SubscriptionOK
	if cache.GetJSON(c.Request.Context(), h.cache, key, &cached) {
		cached.Meta = metaFrom(c)
		middleware.WriteJSON(c, utils.StatusOK, cached)
		return
	}
	sub, err := h.billing.GetSubscription(c.Request.Context(), uid, orgID)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	body := dto.SubscriptionOK{Meta: metaFrom(c), Subscription: subscriptionView(sub)}
	cache.SetJSON(c.Request.Context(), h.cache, key, body, cache.TTLBilling)
	middleware.WriteJSON(c, utils.StatusOK, body)
}
