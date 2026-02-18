package game

import (
	"strings"
	"testing"
)

func TestRandomForageRespectsCategoryAndNutrition(t *testing.T) {
	result, err := RandomForage(101, "tropical_jungle", PlantCategoryFruits, 3, 1)
	if err != nil {
		t.Fatalf("random forage: %v", err)
	}
	if result.Plant.Category != PlantCategoryFruits {
		t.Fatalf("expected fruit category, got %s", result.Plant.Category)
	}
	if result.HarvestGrams <= 0 {
		t.Fatalf("expected positive harvest grams")
	}
	if result.Nutrition.CaloriesKcal <= 0 || result.Nutrition.SugarG < 0 {
		t.Fatalf("expected positive nutrition values, got %+v", result.Nutrition)
	}
}

func TestForageAndConsumeUpdatesPlayerNutrition(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioJungleID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 10},
		Seed:        411,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	beforeSugar := run.Players[0].Nutrition.SugarG
	beforeCalories := run.Players[0].CaloriesReserveKcal

	result, err := run.ForageAndConsume(1, PlantCategoryBerries, 120)
	if err != nil {
		t.Fatalf("forage and consume: %v", err)
	}
	if result.HarvestGrams <= 0 {
		t.Fatalf("expected consumed harvest")
	}
	if run.Players[0].Nutrition.SugarG <= beforeSugar {
		t.Fatalf("expected sugar nutrition total to increase")
	}
	if run.Players[0].CaloriesReserveKcal <= beforeCalories {
		t.Fatalf("expected calorie reserve to increase")
	}
}

func TestWoodFireShelterCraftFlow(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioArcticID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 10},
		Seed:        511,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 2
	run.Weather = WeatherState{Day: 1, Type: WeatherSnow, TemperatureC: -12, StreakDays: 1}

	tree, kg, err := run.GatherWood(1, 2.0)
	if err != nil {
		t.Fatalf("gather wood: %v", err)
	}
	if kg <= 0 {
		t.Fatalf("expected gathered kg > 0")
	}
	if run.totalWoodKg() <= 0 {
		t.Fatalf("expected wood stock")
	}

	if err := run.StartFire(1, tree.WoodType, 1.0); err != nil {
		t.Fatalf("start fire: %v", err)
	}
	if !run.Fire.Lit || run.Fire.Intensity <= 0 || run.Fire.HeatC <= 0 {
		t.Fatalf("expected lit fire with intensity and heat")
	}

	shelters := SheltersForBiome(run.Scenario.Biome)
	if len(shelters) == 0 {
		t.Fatalf("expected shelters for biome")
	}
	if _, err := run.BuildShelter(1, string(shelters[0].ID)); err != nil {
		t.Fatalf("build shelter: %v", err)
	}

	craftables := CraftablesForBiome(run.Scenario.Biome)
	if len(craftables) == 0 {
		t.Fatalf("expected craftables for biome")
	}
	chosen := craftables[0]
	if chosen.RequiresFire && !run.Fire.Lit {
		t.Fatalf("test setup expected lit fire")
	}
	if chosen.RequiresShelter && run.Shelter.Type == "" {
		t.Fatalf("test setup expected shelter")
	}
	_, _, _ = run.GatherWood(1, 2.0) // ensure there is wood for recipes that require it.
	if _, err := run.CraftItem(1, chosen.ID); err != nil {
		t.Fatalf("craft item: %v", err)
	}
	if len(run.CraftedItems) == 0 {
		t.Fatalf("expected crafted inventory to track item")
	}

	impact := run.campImpactForDay()
	if impact.Energy <= 0 {
		t.Fatalf("expected cold-camp setup to provide positive energy impact, got %+v", impact)
	}
}

func TestRunCommandIncludesNewCampSystems(t *testing.T) {
	run := newRunForCommands(t)

	forage := run.ExecuteRunCommand("forage berries p1 120")
	if !forage.Handled || !strings.Contains(strings.ToLower(forage.Message), "foraged") {
		t.Fatalf("expected forage command to be handled, got: %+v", forage)
	}

	wood := run.ExecuteRunCommand("wood gather 1.5 p1")
	if !wood.Handled || !strings.Contains(strings.ToLower(wood.Message), "gathered") {
		t.Fatalf("expected wood gather command handled, got: %+v", wood)
	}

	fire := run.ExecuteRunCommand("fire build 0.7 p1")
	if !fire.Handled {
		t.Fatalf("expected fire build command handled")
	}
	if !run.Fire.Lit {
		t.Fatalf("expected fire to be lit after command, got message: %s", fire.Message)
	}

	shelterList := run.ExecuteRunCommand("shelter list")
	if !shelterList.Handled || !strings.Contains(strings.ToLower(shelterList.Message), "shelters") {
		t.Fatalf("expected shelter list command handled, got: %+v", shelterList)
	}
}
