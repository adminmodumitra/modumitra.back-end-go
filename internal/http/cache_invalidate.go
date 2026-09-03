package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/cache"
)

func (h *handlers) invalidateDevices(c *gin.Context, userID string) {
	if h.cache == nil {
		return
	}
	ctx := c.Request.Context()
	h.cache.DeletePrefix(ctx, cache.DevicesPrefix(userID))
	h.cache.DeletePrefix(ctx, cache.CapabilitiesPrefix(userID))
}

func (h *handlers) invalidateOrgMembers(c *gin.Context, orgID string, userIDs ...string) {
	if h.cache == nil {
		return
	}
	ctx := c.Request.Context()
	h.cache.DeletePrefix(ctx, cache.OrgMembersPrefix(orgID))
	h.cache.Delete(ctx, cache.OrgRoles(orgID))
	for _, uid := range userIDs {
		if uid != "" {
			h.cache.DeletePrefix(ctx, cache.CapabilitiesPrefix(uid))
		}
	}
}

func (h *handlers) invalidateBillingSub(c *gin.Context, userID, orgID string) {
	if h.cache == nil {
		return
	}
	ctx := c.Request.Context()
	if orgID != "" {
		h.cache.Delete(ctx, cache.BillingSubOrg(orgID))
	} else if userID != "" {
		h.cache.Delete(ctx, cache.BillingSubUser(userID))
	}
	if userID != "" {
		h.cache.DeletePrefix(ctx, cache.CapabilitiesPrefix(userID))
	}
}
