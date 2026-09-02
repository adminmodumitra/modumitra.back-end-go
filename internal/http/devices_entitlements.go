package httpapi

import (
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) orgEntitlements(c *gin.Context) {
	logEnter(c, "org_entitlements")
	if !h.requireDevices(c) {
		return
	}
	snap, err := h.devices.Issue(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"), strings.TrimSpace(c.Query("device_id")))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.EntitlementsOK{
		Meta:                metaFrom(c),
		EntitlementSnapshot: snapshotView(snap),
	})
}

func (h *handlers) meCapabilities(c *gin.Context) {
	logEnter(c, "me_capabilities")
	if !h.requireDevices(c) {
		return
	}
	res, err := h.devices.Capabilities(c.Request.Context(), middleware.UserIDFrom(c), strings.TrimSpace(c.Query("organization_id")))
	if err != nil {
		middleware.Abort(c, err)
		return
	}
	middleware.WriteJSON(c, utils.StatusOK, dto.CapabilitiesOK{
		Meta:               metaFrom(c),
		UserID:             res.UserID,
		OrganizationID:     res.OrganizationID,
		Role:               res.Role,
		MembershipStatus:   res.MembershipStatus,
		Plan:               res.Plan,
		SubscriptionStatus: res.SubscriptionStatus,
		Capabilities:       capabilitiesView(res.Capabilities),
		IssuedAt:           rfcTime(res.IssuedAt),
	})
}
