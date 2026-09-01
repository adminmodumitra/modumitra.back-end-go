package billing

import "time"

func catalogEpoch() time.Time {
	return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
}

// SeedPlans returns the three catalog rows from schema-v1.sql.
func SeedPlans() []Plan {
	at := catalogEpoch()
	return []Plan{
		{
			ID: PlanIDTrial, Code: PlanTrial, Name: "Trial", Interval: IntervalNone,
			Currency: "usd", UnitAmountMinor: 0, SeatsIncluded: 3, MaxDevicesPerUser: 3,
			TrialDays: 14, Sellable: false, Active: true,
			Capabilities: Capabilities{
				CloudProjects: true, MaxSeats: 3, MaxDevices: 3, MaxCloudProjects: 10,
			},
			SortOrder: 0, CreatedAt: at, UpdatedAt: at,
		},
		{
			ID: PlanIDStudio, Code: PlanStudio, Name: "Studio", Interval: IntervalMonth,
			Currency: "usd", UnitAmountMinor: 2900, SeatsIncluded: 1, MaxDevicesPerUser: 3,
			TrialDays: 14, Sellable: true, Active: true,
			Capabilities: Capabilities{
				CloudProjects: true, OrganizationAssets: true, ProposalBranding: true,
				MaxSeats: 1, MaxDevices: 3, MaxCloudProjects: 50,
			},
			Provider: ProviderStripe, ProviderPriceID: "price_studio_monthly_placeholder",
			SortOrder: 10, CreatedAt: at, UpdatedAt: at,
		},
		{
			ID: PlanIDTeam, Code: PlanTeam, Name: "Team", Interval: IntervalMonth,
			Currency: "usd", UnitAmountMinor: 4900, SeatsIncluded: 5, MaxDevicesPerUser: 3,
			TrialDays: 14, Sellable: true, Active: true,
			Capabilities: Capabilities{
				CloudProjects: true, OrganizationAssets: true, ProposalBranding: true,
				EngineeringHandoff: true, MaxSeats: 5, MaxDevices: 3, MaxCloudProjects: 500,
			},
			Provider: ProviderStripe, ProviderPriceID: "price_team_monthly_placeholder",
			SortOrder: 20, CreatedAt: at, UpdatedAt: at,
		},
	}
}

func clonePlan(p *Plan) *Plan {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}
