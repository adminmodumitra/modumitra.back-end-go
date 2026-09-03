package cache

import (
	"fmt"
	"strings"
	"time"
)

// Key scheme (cabinet read-through):
//
//	cabinet:me:{userID}
//	cabinet:me:capabilities:{userID}                 — no org scope
//	cabinet:me:capabilities:{userID}:org:{orgID}     — org-scoped (header/query)
//	cabinet:org:{orgID}:members[:{status}:{role}:{cursor}:{limit}]
//	cabinet:org:{orgID}:roles
//	cabinet:billing:plans
//	cabinet:billing:sub:user:{userID}
//	cabinet:billing:sub:org:{orgID}
//	cabinet:devices:user:{userID}[:inc={0|1}:c={cursor}:l={limit}]
//
// Domain payloads only are meaningful; handlers re-wrap request meta on hit.
// List keys embed filter/cursor so pages do not collide; writers DeletePrefix.

const (
	TTLMe      = 60 * time.Second
	TTLCaps    = 60 * time.Second
	TTLOrg     = 60 * time.Second
	TTLBilling = 180 * time.Second
	TTLDevices = 60 * time.Second
)

func Me(userID string) string { return "cabinet:me:" + userID }

func Capabilities(userID, orgID string) string {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return "cabinet:me:capabilities:" + userID
	}
	return "cabinet:me:capabilities:" + userID + ":org:" + orgID
}

func CapabilitiesPrefix(userID string) string {
	return "cabinet:me:capabilities:" + userID
}

func OrgMembers(orgID, status, role, cursor string, limit int) string {
	base := "cabinet:org:" + orgID + ":members"
	status, role, cursor = strings.TrimSpace(status), strings.TrimSpace(role), strings.TrimSpace(cursor)
	if status == "" && role == "" && cursor == "" && limit == 0 {
		return base
	}
	return fmt.Sprintf("%s:%s:%s:%s:%d", base, status, role, cursor, limit)
}

func OrgMembersPrefix(orgID string) string { return "cabinet:org:" + orgID + ":members" }

func OrgRoles(orgID string) string { return "cabinet:org:" + orgID + ":roles" }

func BillingPlans() string { return "cabinet:billing:plans" }

func BillingSubUser(userID string) string { return "cabinet:billing:sub:user:" + userID }

func BillingSubOrg(orgID string) string { return "cabinet:billing:sub:org:" + orgID }

func Devices(userID string, includeDeactivated bool, cursor string, limit int) string {
	inc := 0
	if includeDeactivated {
		inc = 1
	}
	cursor = strings.TrimSpace(cursor)
	if cursor == "" && limit == 0 && !includeDeactivated {
		return "cabinet:devices:user:" + userID
	}
	return fmt.Sprintf("cabinet:devices:user:%s:inc=%d:c=%s:l=%d", userID, inc, cursor, limit)
}

func DevicesPrefix(userID string) string { return "cabinet:devices:user:" + userID }
