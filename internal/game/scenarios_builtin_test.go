package game

import "testing"

func TestBuiltInScenariosAreValid(t *testing.T) {
	scenarios := BuiltInScenarios()
	if len(scenarios) < 20 {
		t.Fatalf("expected expanded built-in scenario catalog, got %d", len(scenarios))
	}

	seen := make(map[ScenarioID]bool, len(scenarios))
	modeCount := map[GameMode]int{
		ModeAlone:            0,
		ModeNakedAndAfraid:   0,
		ModeNakedAndAfraidXL: 0,
	}
	for _, scenario := range scenarios {
		if scenario.ID == "" {
			t.Fatalf("scenario has empty ID: %+v", scenario)
		}
		if seen[scenario.ID] {
			t.Fatalf("duplicate scenario ID: %s", scenario.ID)
		}
		seen[scenario.ID] = true

		if scenario.Name == "" {
			t.Fatalf("scenario %s has empty name", scenario.ID)
		}
		if scenario.Description == "" {
			t.Fatalf("scenario %s has empty description", scenario.ID)
		}
		if scenario.Daunting == "" {
			t.Fatalf("scenario %s has empty daunting text", scenario.ID)
		}
		if scenario.Motivation == "" {
			t.Fatalf("scenario %s has empty motivation text", scenario.ID)
		}
		if len(scenario.SupportedModes) == 0 {
			t.Fatalf("scenario %s has no supported modes", scenario.ID)
		}
		for _, mode := range scenario.SupportedModes {
			modeCount[mode]++
		}
		if scenario.DefaultDays <= 0 {
			t.Fatalf("scenario %s must have positive DefaultDays, got %d", scenario.ID, scenario.DefaultDays)
		}
		if len(scenario.SeasonSets) == 0 {
			t.Fatalf("scenario %s must have at least one season set", scenario.ID)
		}
		if scenario.DefaultSeasonSetID == "" {
			t.Fatalf("scenario %s must have a default season set", scenario.ID)
		}
	}

	for mode, count := range modeCount {
		if count == 0 {
			t.Fatalf("expected at least one scenario for mode %s", mode)
		}
	}
}
