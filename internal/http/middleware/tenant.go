package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// TenantResolver maps (orgID, userID) to an active membership role.
type TenantResolver interface {
	Resolve(ctx context.Context, orgID, userID string) (role string, err error)
}

// orgScoped reports whether this authed route is organization-scoped and
// returns the path organization id. Unscoped authed routes (me, logout-all,
// create/list orgs, accept invitation) are skipped.
func orgScoped(c *gin.Context) (orgID string, ok bool) {
	fp := c.FullPath()
	if !strings.HasPrefix(fp, "/v1/organizations/:id") {
		return "", false
	}
	id := c.Param("id")
	if id == "" {
		return "", false
	}
	return id, true
}

// Tenant resolves the caller's membership in the path organization and sets
// organization_id + role on the context. Missing/inactive membership → 404.
func Tenant(resolver TenantResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, scoped := orgScoped(c)
		if !scoped {
			c.Next()
			return
		}
		if hdr := strings.TrimSpace(c.GetHeader(utils.HeaderOrganization)); hdr != "" && hdr != orgID {
			Abort(c, utils.E(utils.CodeTenantMismatch))
			return
		}
		if resolver == nil {
			Abort(c, utils.E(utils.CodeNotFound))
			return
		}
		uid := UserIDFrom(c)
		role, err := resolver.Resolve(c.Request.Context(), orgID, uid)
		if err != nil {
			Abort(c, err)
			return
		}
		c.Set(utils.CtxOrgID, orgID)
		c.Set(utils.CtxRole, role)
		c.Next()
	}
}

func OrgIDFrom(c *gin.Context) string {
	if v, ok := c.Get(utils.CtxOrgID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return c.Param("id")
}

func RoleFrom(c *gin.Context) string {
	if v, ok := c.Get(utils.CtxRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
