package httpapi

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/http/middleware"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// callerOrg resolves the organization a non-path-scoped route acts for:
// X-Organization-Id when sent (active membership required, else 404), or the
// caller's primary organization. A personal account with no organization gets
// ("", "", nil). Sets organization_id and role on the context for logs.
func (h *handlers) callerOrg(c *gin.Context) (orgID, role string, err error) {
	if h.orgs == nil {
		return "", "", utils.E(utils.CodeServiceUnavailable)
	}
	ctx := c.Request.Context()
	uid := middleware.UserIDFrom(c)
	orgID = strings.TrimSpace(c.GetHeader(utils.HeaderOrganization))
	if orgID == "" {
		o, err := h.orgs.PrimaryForUser(ctx, uid)
		if err != nil {
			return "", "", err
		}
		if o == nil {
			return "", "", nil
		}
		orgID = o.ID
	}
	role, err = h.orgs.Resolve(ctx, orgID, uid)
	if err != nil {
		return "", "", err
	}
	c.Set(utils.CtxOrgID, orgID)
	c.Set(utils.CtxRole, role)
	return orgID, role, nil
}
