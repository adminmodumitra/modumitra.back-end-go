package device

import "time"

func ValidPlatform(p string) bool {
	switch p {
	case PlatformMacOS, PlatformWindows, PlatformLinux:
		return true
	default:
		return false
	}
}

func planOf(o *OrgRef, now time.Time) (plan, status string) {
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
	c := Capabilities{MaxSeats: 3, MaxDevices: 3, MaxCloudProjects: 10}
	switch role {
	case "owner", "admin":
		c.CloudProjects = true
		c.OrganizationAssets = true
		c.ProposalBranding = true
		c.EngineeringHandoff = true
	case "billing_admin":
		// all product flags remain false
	case "sales_designer":
		c.CloudProjects = true
		c.OrganizationAssets = true
		c.ProposalBranding = true
	case "engineer":
		c.CloudProjects = true
		c.OrganizationAssets = true
		c.EngineeringHandoff = true
	case "production":
		c.CloudProjects = true
		c.OrganizationAssets = true
	case "viewer":
		// none
	}
	return c
}
