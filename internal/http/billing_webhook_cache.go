package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
)

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
