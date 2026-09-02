package billing

import "testing"

func TestSeedPlansUsesInjectedIDs(t *testing.T) {
	plans := SeedPlans("price_1Abc", "price_1Def")
	if len(plans) != 3 {
		t.Fatalf("len=%d", len(plans))
	}
	byCode := map[string]Plan{}
	for _, p := range plans {
		byCode[p.Code] = p
	}
	studio := byCode[PlanStudio]
	if studio.ProviderPriceID != "price_1Abc" {
		t.Fatalf("studio price=%q", studio.ProviderPriceID)
	}
	team := byCode[PlanTeam]
	if team.ProviderPriceID != "price_1Def" {
		t.Fatalf("team price=%q", team.ProviderPriceID)
	}
	if studio.Provider != ProviderStripe || team.Provider != ProviderStripe {
		t.Fatalf("provider studio=%q team=%q", studio.Provider, team.Provider)
	}
	if byCode[PlanTrial].ProviderPriceID != "" {
		t.Fatalf("trial should have no price id, got %q", byCode[PlanTrial].ProviderPriceID)
	}
}
