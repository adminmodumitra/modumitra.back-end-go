package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func actionFor(method, fullPath string) (string, bool) {
	if !strings.HasPrefix(fullPath, "/v1/organizations/:id") {
		return "", false
	}
	rest := strings.TrimPrefix(fullPath, "/v1/organizations/:id")
	switch {
	case rest == "" && method == http.MethodGet:
		return org.ActionOrgRead, true
	case rest == "" && method == http.MethodPatch:
		return org.ActionOrgWrite, true
	case rest == "" && method == http.MethodDelete:
		return org.ActionOrgDelete, true
	case rest == "/leave" && method == http.MethodPost:
		return org.ActionOrgLeave, true
	case rest == "/transfer-ownership" && method == http.MethodPost:
		return org.ActionOrgTransfer, true
	case rest == "/roles" && method == http.MethodGet:
		return org.ActionOrgRoles, true
	case rest == "/entitlements" && method == http.MethodGet:
		return org.ActionOrgRead, true
	case rest == "/subscription" && method == http.MethodGet:
		return org.ActionBillingRead, true
	case rest == "/checkout-session" && method == http.MethodPost:
		return org.ActionBillingWrite, true
	case rest == "/portal-session" && method == http.MethodPost:
		return org.ActionBillingWrite, true
	case rest == "/subscription/cancel" && method == http.MethodPost:
		return org.ActionBillingWrite, true
	case rest == "/subscription/resume" && method == http.MethodPost:
		return org.ActionBillingWrite, true
	case rest == "/invoices" && method == http.MethodGet:
		return org.ActionBillingRead, true
	case rest == "/invoices/:invoiceId" && method == http.MethodGet:
		return org.ActionBillingRead, true
	case rest == "/seats" && method == http.MethodGet:
		return org.ActionBillingRead, true
	case rest == "/seats/preview-downgrade" && method == http.MethodPost:
		return org.ActionBillingWrite, true
	case rest == "/members" && method == http.MethodGet:
		return org.ActionMembersList, true
	case rest == "/members/:userId" && method == http.MethodPatch:
		return org.ActionMembersWrite, true
	case rest == "/members/:userId/suspend" && method == http.MethodPost:
		return org.ActionMembersSuspend, true
	case rest == "/members/:userId" && method == http.MethodDelete:
		return org.ActionMembersRemove, true
	case rest == "/invitations" && method == http.MethodPost:
		return org.ActionInvitesCreate, true
	case rest == "/invitations" && method == http.MethodGet:
		return org.ActionInvitesList, true
	case rest == "/invitations/:invitationId" && method == http.MethodDelete:
		return org.ActionInvitesRevoke, true
	case rest == "/invitations/:invitationId/resend" && method == http.MethodPost:
		return org.ActionInvitesResend, true
	default:
		return "", true // org-scoped but unknown: require active member only
	}
}

// RBAC enforces RolePermits for org-scoped authed routes. Unscoped routes no-op.
func RBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		action, scoped := actionFor(c.Request.Method, c.FullPath())
		if !scoped {
			c.Next()
			return
		}
		role := RoleFrom(c)
		if role == "" {
			Abort(c, utils.E(utils.CodeNotFound))
			return
		}
		if action != "" && !org.RolePermits(role, action) {
			Abort(c, utils.E(utils.CodeForbidden))
			return
		}
		c.Next()
	}
}
