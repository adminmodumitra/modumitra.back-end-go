package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func (h *handlers) requireOrgs(c *gin.Context) bool {
	if h.orgs == nil {
		middleware.Abort(c, utils.E(utils.CodeServiceUnavailable))
		return false
	}
	return true
}

func requireIdempotencyKey(c *gin.Context) bool {
	if c.GetHeader(utils.HeaderIdempotency) == "" {
		middleware.Abort(c, utils.E(utils.CodeValidation).WithDetails(map[string]any{
			"field":  "Idempotency-Key",
			"reason": "required",
		}))
		return false
	}
	return true
}

func (h *handlers) memberUserView(c *gin.Context, userID string) *dto.MemberUserView {
	if h.orgs == nil {
		return nil
	}
	u, err := h.orgs.MemberUser(c.Request.Context(), userID)
	if err != nil || u == nil {
		return nil
	}
	return &dto.MemberUserView{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		EmailVerified: u.EmailVerified(),
	}
}

func (h *handlers) primaryOrgView(c *gin.Context, userID string) *dto.OrganizationView {
	if h.orgs == nil {
		return nil
	}
	o, err := h.orgs.PrimaryForUser(c.Request.Context(), userID)
	if err != nil || o == nil {
		return nil
	}
	v := organizationView(o)
	return &v
}
