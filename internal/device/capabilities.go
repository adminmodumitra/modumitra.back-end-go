package device

import (
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func ValidPlatform(p string) bool {
	switch p {
	case PlatformMacOS, PlatformWindows, PlatformLinux:
		return true
	default:
		return false
	}
}

func planOf(o *org.Organization, now time.Time) (plan, status string) {
	if o == nil {
		return PlanNone, SubNone
	}
	if o.TrialEndsAt != nil && o.TrialEndsAt.After(now) {
		return PlanTrial, SubTrialing
	}
	return PlanNone, SubNone
}

// Intersect is role catalog ∩ plan defaults. Plan none → all false / zeros.
// Trial and studio-like plans grant the four product flags; numeric limits
// are 3 / 3 / 10. Role catalog gates the boolean product keys.
func Intersect(role, plan string) Capabilities {
	if plan == PlanNone || plan == "" {
		return Capabilities{}
	}
	catalog := org.RolesCatalog()
	r := catalog[role]
	return Capabilities{
		CloudProjects:      r["cloud_projects"],
		OrganizationAssets: r["organization_assets"],
		ProposalBranding:   r["proposal_branding"],
		EngineeringHandoff: r["engineering_handoff"],
		MaxSeats:           3,
		MaxDevices:         3,
		MaxCloudProjects:   10,
	}
}
