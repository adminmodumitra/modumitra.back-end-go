package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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
	sub, err := h.billing.Cancel(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), atEnd)
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}

func (h *handlers) resumeSubscription(c *gin.Context) {
	logEnter(c, "resume_subscription")
	if !h.requireBilling(c) {
		return
	}
	if !requireIdempotencyKey(c) {
		return
	}
	sub, err := h.billing.Resume(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.SubscriptionOK{
		Meta:         metaFrom(c),
		Subscription: subscriptionView(sub),
	})
}
