package httpapi

import (
	"encoding/json"
	"io"

	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/billing"
	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

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

func (h *handlers) invalidateWebhookBilling(c *gin.Context, obj map[string]any) {
	if h.cache == nil {
		return
	}
	ctx := c.Request.Context()
	orgID := ""
	userID := ""
	if obj != nil {
		for _, k := range []string{"organization_id", "client_reference_id"} {
			if v, ok := obj[k].(string); ok && v != "" {
				orgID = v
				break
			}
		}
	}
	if orgID != "" {
		h.cache.Delete(ctx, cache.BillingSubOrg(orgID))
		// Best-effort: drop org-scoped capability entries (*:org:{orgID}).
		h.cache.DeletePrefix(ctx, "cabinet:me:capabilities:")
	}
	// Personal checkout payloads may only carry billing_account_id; drop user sub
	// keys when we can read a user owner id from the object.
	if obj != nil {
		if v, ok := obj["owner_user_id"].(string); ok {
			userID = v
		}
	}
	if userID != "" {
		h.cache.Delete(ctx, cache.BillingSubUser(userID))
		h.cache.DeletePrefix(ctx, cache.CapabilitiesPrefix(userID))
	}
	h.cache.Delete(ctx, cache.BillingPlans())
}
