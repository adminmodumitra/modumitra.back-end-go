package httpapi

import (
	"encoding/json"
	"io"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// billingWebhook godoc
// @Summary Billing provider webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param Stripe-Signature header string false "Stripe signature"
// @Param X-Billing-Signature header string false "Billing HMAC signature"
// @Param body body object true "Provider event payload"
// @Success 200 {object} dto.WebhookOK
// @Failure 401 {object} utils.ErrorResponse
// @Router /v1/webhooks/billing [post]
func (h *handlers) billingWebhook(c *gin.Context) {
	logEnter(c, "billing_webhook")
	if !h.requireBilling(c) {
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		middleware.Abort(c, utils.E(utils.CodeInvalidPayload))
		return
	}
	secret := h.cfg.BillingWebhookSecret
	if err := billing.VerifyWebhookSignature(secret, raw, c.GetHeader(utils.HeaderStripeSignature), c.GetHeader(utils.HeaderBillingSignature)); err != nil {
		middleware.Abort(c, err)
		return
	}
	var meta struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object map[string]any `json:"object"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &meta) == nil {
		if meta.ID != "" {
			c.Set("event_id", meta.ID)
		}
		if meta.Type != "" {
			c.Set("event_type", meta.Type)
		}
	}
	if err := h.billing.HandleWebhook(c.Request.Context(), raw); err != nil {
		middleware.Abort(c, err)
		return
	}
	h.invalidateWebhookBilling(c, meta.Data.Object)
	rid := middleware.RequestIDFrom(c)
	if rid == "" {
		rid = "req_wh_billing"
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.WebhookOK{
		Meta:     utils.Meta{RequestID: rid, TraceID: middleware.TraceIDFrom(c)},
		Received: true,
	})
}
