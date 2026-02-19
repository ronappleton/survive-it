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

func TestResourcesIncludeClayInRealisticBiomes(t *testing.T) {
	withClay := ResourcesForBiome("river_delta_wetlands")
	hasClay := false
	for _, resource := range withClay {
		if resource.ID == "clay" {
			hasClay = true
			break
		}
	}
	if !hasClay {
		t.Fatalf("expected clay resource in delta/wetlands biome")
	}
}

func TestEmberChanceDropsWithWetWood(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 10},
		Seed:        922,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 2
	run.FirePrep = FirePrepState{
		TinderBundles:   2,
		KindlingBundles: 2,
		FeatherSticks:   1,
		TinderQuality:   0.9,
		KindlingQuality: 0.9,
		FeatherQuality:  0.9,
	}

	dryRun := run
	dryRun.WoodStock = []WoodStock{{Type: WoodTypeResinous, Kg: 2.0, Wetness: 0.1}}
	dryRun.Weather = WeatherState{Type: WeatherClear, TemperatureC: 10}
	dryChance := dryRun.emberChance(FireMethodBowDrill, dryRun.Players[0], WoodTypeResinous)

	wetRun := run
	wetRun.WoodStock = []WoodStock{{Type: WoodTypeResinous, Kg: 2.0, Wetness: 0.9}}
	wetRun.Weather = WeatherState{Type: WeatherHeavyRain, TemperatureC: 8}
	wetChance := wetRun.emberChance(FireMethodBowDrill, wetRun.Players[0], WoodTypeResinous)

	if wetChance >= dryChance {
		t.Fatalf("expected wet conditions to reduce ember chance, dry=%.2f wet=%.2f", dryChance, wetChance)
	}
}

func TestRunCommandPrimitiveFireFlowHandled(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioJungleID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 7},
		Seed:        707,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 3

	if res := run.ExecuteRunCommand("wood gather 4 p1"); !res.Handled {
		t.Fatalf("expected wood gather to be handled")
	}
	if res := run.ExecuteRunCommand("collect vine_fiber 2 p1"); !res.Handled {
		t.Fatalf("expected collect vine_fiber to be handled")
	}
	if res := run.ExecuteRunCommand("collect inner_bark_fiber 4 p1"); !res.Handled {
		t.Fatalf("expected collect inner_bark_fiber to be handled")
	}
	if res := run.ExecuteRunCommand("fire prep tinder 2 p1"); !res.Handled {
		t.Fatalf("expected fire prep tinder handled")
	}
	if res := run.ExecuteRunCommand("fire prep kindling 2 p1"); !res.Handled {
		t.Fatalf("expected fire prep kindling handled")
	}
	if res := run.ExecuteRunCommand("fire prep feather 1 p1"); !res.Handled {
		t.Fatalf("expected fire prep feather handled")
	}

	for _, craft := range []string{
		"bow_drill_spindle",
		"bow_drill_hearth_board",
		"bow_drill_bow",
		"bearing_block",
	} {
		if _, err := run.CraftItem(1, craft); err != nil {
			t.Fatalf("expected craft %s to succeed: %v", craft, err)
		}
	}

	ember := run.ExecuteRunCommand("fire ember bow p1")
	if !ember.Handled {
		t.Fatalf("expected fire ember command handled")
	}
	if !strings.Contains(strings.ToLower(ember.Message), "chance") {
		t.Fatalf("expected chance feedback in ember command message, got: %s", ember.Message)
	}

	ignite := run.ExecuteRunCommand("fire ignite hardwood 0.8 p1")
	if !ignite.Handled {
		t.Fatalf("expected fire ignite command handled")
	}
}

func TestCraftableCatalogExpandedForTreeProducts(t *testing.T) {
	catalog := CraftableCatalog()
	if len(catalog) < 30 {
		t.Fatalf("expected expanded craft catalog with >= 30 entries, got %d", len(catalog))
	}

	ids := map[string]bool{}
	for _, entry := range catalog {
		ids[entry.ID] = true
	}
	required := []string{
		"bow_drill_spindle",
		"bow_drill_hearth_board",
		"bow_drill_bow",
		"bearing_block",
		"hand_drill_spindle",
		"hand_drill_hearth_board",
		"ridge_pole_kit",
		"trap_trigger_set",
		"fish_spear_shaft",
		"resin_torch",
		"pitch_glue",
		"clay_pot",
		"clay_heat_core",
	}
	for _, id := range required {
		if !ids[id] {
			t.Fatalf("expected craftable catalog to include %s", id)
		}
	}
}

func TestCraftItemReturnsQualityAndAdvancesClock(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 14},
		Seed:        1201,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 3
	if _, _, err := run.CollectResource(1, "inner_bark_fiber", 2); err != nil {
		t.Fatalf("collect inner_bark_fiber: %v", err)
	}

	beforeClock := run.ClockHours
	outcome, err := run.CraftItem(1, "natural_twine")
	if err != nil {
		t.Fatalf("craft natural_twine: %v", err)
	}
	if outcome.Quality == "" {
		t.Fatalf("expected craft quality")
	}
	if outcome.HoursSpent <= 0 {
		t.Fatalf("expected positive craft time")
	}
	if run.ClockHours <= beforeClock && run.Day == 1 {
		t.Fatalf("expected clock to advance after craft")
	}
	if !strings.Contains(strings.ToLower(run.PersonalInventorySummary(1)), "natural_twine") && !strings.Contains(strings.ToLower(run.CampInventorySummary()), "natural_twine") {
		t.Fatalf("expected natural_twine to be stored in inventory")
	}
}

func TestInventoryStashAndTakeFlow(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 14},
		Seed:        1202,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	if err := run.AddPersonalInventoryItem(1, InventoryItem{ID: "field_stick", Name: "Field Stick", Unit: "piece", Qty: 3, WeightKg: 0.2, Category: "field"}); err != nil {
		t.Fatalf("add personal item: %v", err)
	}
	if _, err := run.StashPersonalItem(1, "field_stick", 2); err != nil {
		t.Fatalf("stash personal item: %v", err)
	}
	if _, err := run.TakeCampItem(1, "field_stick", 1); err != nil {
		t.Fatalf("take camp item: %v", err)
	}

	player, ok := run.playerByID(1)
	if !ok {
		t.Fatalf("player not found")
	}
	item, ok := inventoryItemByID(player.PersonalItems, "field_stick")
	if !ok {
		t.Fatalf("expected field_stick in personal inventory")
	}
	if item.Qty != 2 {
		t.Fatalf("expected personal qty=2, got %.0f", item.Qty)
	}
}

func TestTrapCheckCollectsPendingCatch(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1203,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 3

	if _, _, err := run.CollectResource(1, "inner_bark_fiber", 2); err != nil {
		t.Fatalf("collect inner_bark_fiber: %v", err)
	}
	if _, err := run.CraftItem(1, "natural_twine"); err != nil {
		t.Fatalf("craft natural_twine: %v", err)
	}
	if _, err := run.SetTrap(1, "peg_snare"); err != nil {
		t.Fatalf("set trap: %v", err)
	}
	if len(run.PlacedTraps) == 0 {
		t.Fatalf("expected placed trap")
	}
	run.PlacedTraps[0].PendingCatchKg = 0.6
	run.PlacedTraps[0].PendingCatchType = "small_game"
	run.PlacedTraps[0].Armed = false

	check := run.CheckTraps()
	if check.CollectedKg <= 0 {
		t.Fatalf("expected trap check to collect catch, got %+v", check)
	}
	if !strings.Contains(strings.ToLower(run.CampInventorySummary()), "small_game_carcass") {
		t.Fatalf("expected small_game_carcass in camp inventory after check")
	}
}

func TestGutCookEatFlowFromTrappedCatch(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1301,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Bushcraft = 2
	run.Players[0].Crafting = 25
	run.Players[0].Kit = []KitItem{KitSixInchKnife}

	if err := run.addCampInventoryItem(InventoryItem{ID: "small_game_carcass", Name: "Small Game Carcass", Unit: "kg", Qty: 1.2, WeightKg: 1.2, Category: "carcass"}); err != nil {
		t.Fatalf("add carcass: %v", err)
	}

	gut := run.ExecuteRunCommand("gut small_game_carcass 1 p1")
	if !gut.Handled {
		t.Fatalf("expected gut command handled")
	}
	if !strings.Contains(strings.ToLower(gut.Message), "gutted") {
		t.Fatalf("expected gut result message, got: %s", gut.Message)
	}

	if res := run.ExecuteRunCommand("wood gather 4 p1"); !res.Handled {
		t.Fatalf("expected wood gather handled")
	}
	if res := run.ExecuteRunCommand("fire build 1 p1"); !res.Handled {
		t.Fatalf("expected fire build handled")
	}

	cook := run.ExecuteRunCommand("cook raw_small_game_meat 0.3 p1")
	if !cook.Handled {
		t.Fatalf("expected cook command handled, got: %+v", cook)
	}
	eat := run.ExecuteRunCommand("eat cooked_small_game_meat 220 p1")
	if !eat.Handled {
		t.Fatalf("expected eat command handled, got: %+v", eat)
	}
	if !strings.Contains(strings.ToLower(eat.Message), "ate") {
		t.Fatalf("expected eat message, got: %s", eat.Message)
	}
}

func TestGoTravelUsesWatercraftWhenAvailable(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1302,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Agility = 1
	run.Players[0].Endurance = 1
	startX, startY := run.CurrentMapPosition()

	without, err := run.TravelMove(1, "north", 3)
	if err != nil {
		t.Fatalf("travel without watercraft: %v", err)
	}

	run.CraftedItems = append(run.CraftedItems, "brush_raft")
	run.Travel.PosX = startX
	run.Travel.PosY = startY
	run.Players[0].Energy = 100
	run.Players[0].Hydration = 100
	with, err := run.TravelMove(1, "north", 3)
	if err != nil {
		t.Fatalf("travel with watercraft: %v", err)
	}
	if with.WatercraftUsed == "" {
		t.Fatalf("expected watercraft to be used in water biome")
	}
	if with.TravelSpeedKmph <= without.TravelSpeedKmph {
		t.Fatalf("expected watercraft travel speed boost, without=%.2f with=%.2f", without.TravelSpeedKmph, with.TravelSpeedKmph)
	}
}

func TestCraftedClothingModifiesWeatherImpact(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioArcticID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 10},
		Seed:        1303,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	player := run.Players[0]
	base := statDelta{Energy: -2, Hydration: -1, Morale: -2}
	noGear := run.applyCraftedWeatherModifiersForPlayer(base, player, WeatherSnow, -10)
	player.PersonalItems = append(player.PersonalItems, InventoryItem{ID: "hide_jacket", Name: "Hide Jacket", Unit: "set", Qty: 1, WeightKg: 2})
	withGear := run.applyCraftedWeatherModifiersForPlayer(base, player, WeatherSnow, -10)
	if withGear.Energy <= noGear.Energy {
		t.Fatalf("expected clothing to reduce cold energy penalty, noGear=%+v withGear=%+v", noGear, withGear)
	}
	if withGear.Morale <= noGear.Morale {
		t.Fatalf("expected clothing to improve morale in cold weather, noGear=%+v withGear=%+v", noGear, withGear)
	}
}

func TestPreserveSmokeCommandCreatesSmokedMeat(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1401,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Crafting = 30
	if err := run.addCampInventoryItem(InventoryItem{ID: "raw_fish_meat", Name: "Raw Fish Meat", Unit: "kg", Qty: 1.1, WeightKg: 1, Category: "food"}); err != nil {
		t.Fatalf("add raw fish: %v", err)
	}
	if res := run.ExecuteRunCommand("wood gather 3 p1"); !res.Handled {
		t.Fatalf("expected wood gather handled")
	}
	if res := run.ExecuteRunCommand("fire build 1 p1"); !res.Handled {
		t.Fatalf("expected fire build handled")
	}

	preserve := run.ExecuteRunCommand("preserve smoke raw_fish_meat 0.8 p1")
	if !preserve.Handled {
		t.Fatalf("expected preserve command handled")
	}
	if !strings.Contains(strings.ToLower(preserve.Message), "smoked_fish_meat") {
		t.Fatalf("expected smoked output in message, got: %s", preserve.Message)
	}
	if inventoryTotalQtyByID(run.CampInventory, "smoked_fish_meat") <= 0 {
		t.Fatalf("expected smoked fish meat in inventory")
	}
}

func TestShelterStageProgressionAndLocationModifiers(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1501,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	shelter, err := run.BuildShelter(1, string(ShelterLeanTo))
	if err != nil {
		t.Fatalf("build shelter stage 1: %v", err)
	}
	if run.Shelter.Stage != 1 {
		t.Fatalf("expected stage 1 after first build, got %d", run.Shelter.Stage)
	}

	spec, ok := shelterByID(shelter.ID)
	if !ok {
		t.Fatalf("expected shelter spec")
	}
	base := shelterMetricsForState(spec, run.Shelter)

	thatch, ok := run.findResourceForBiome("thatch_bundle")
	if !ok {
		t.Fatalf("expected thatch_bundle resource")
	}
	if err := run.addResourceStock(thatch, 1); err != nil {
		t.Fatalf("add thatch bundle: %v", err)
	}

	if _, err := run.BuildShelter(1, string(ShelterLeanTo)); err != nil {
		t.Fatalf("build shelter stage 2: %v", err)
	}
	if run.Shelter.Stage != 2 {
		t.Fatalf("expected stage 2 after second build, got %d", run.Shelter.Stage)
	}

	idx, ok := run.topoIndex(run.Shelter.SiteX, run.Shelter.SiteY)
	if !ok {
		t.Fatalf("expected valid shelter topology index")
	}
	cell := run.Topology.Cells[idx]
	cell.Flags |= TopoFlagWater
	run.Topology.Cells[idx] = cell

	withLocation, ok := run.currentShelterMetrics()
	if !ok {
		t.Fatalf("expected active shelter metrics")
	}
	if withLocation.DrynessProtection >= base.DrynessProtection {
		t.Fatalf("expected near-water site to reduce dryness protection, base=%d with=%d", base.DrynessProtection, withLocation.DrynessProtection)
	}
}

func TestLogCabinBuildRequiresMultipleStages(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 60},
		Seed:        1502,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	addResource := func(id string, qty float64) {
		spec, ok := run.findResourceForBiome(id)
		if !ok {
			t.Fatalf("resource %s not found", id)
		}
		if err := run.addResourceStock(spec, qty); err != nil {
			t.Fatalf("add resource %s: %v", id, err)
		}
	}
	addResource("stone_cobble", 3)
	addResource("thatch_bundle", 3)
	addResource("mud", 2.0)

	spec, ok := shelterByID(ShelterLogCabin)
	if !ok {
		t.Fatalf("expected log cabin spec")
	}
	if len(spec.Stages) < 5 {
		t.Fatalf("expected log cabin to have many stages, got %d", len(spec.Stages))
	}

	for i := 1; i <= len(spec.Stages); i++ {
		if _, err := run.BuildShelter(1, string(ShelterLogCabin)); err != nil {
			t.Fatalf("stage %d build failed: %v", i, err)
		}
		if run.Shelter.Stage != i {
			t.Fatalf("expected stage %d after build, got %d", i, run.Shelter.Stage)
		}
	}
	if _, err := run.BuildShelter(1, string(ShelterLogCabin)); err == nil {
		t.Fatalf("expected final build attempt to fail once fully built")
	}
}

func TestAdvanceDayFoodDegradationRawVsPreserved(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioJungleID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        1402,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	if err := run.addCampInventoryItem(InventoryItem{ID: "raw_small_game_meat", Name: "Raw Small Game Meat", Unit: "kg", Qty: 1.0, WeightKg: 1, Category: "food"}); err != nil {
		t.Fatalf("add raw meat: %v", err)
	}
	if err := run.addCampInventoryItem(InventoryItem{ID: "smoked_small_game_meat", Name: "Smoked Small Game Meat", Unit: "kg", Qty: 1.0, WeightKg: 1, Category: "food"}); err != nil {
		t.Fatalf("add smoked meat: %v", err)
	}

	for i := 0; i < 4; i++ {
		run.AdvanceDay()
	}

	rawRemaining := inventoryTotalQtyByID(run.CampInventory, "raw_small_game_meat")
	smokedRemaining := inventoryTotalQtyByID(run.CampInventory, "smoked_small_game_meat")
	spoiled := inventoryTotalQtyByID(run.CampInventory, "spoiled_meat")

	if rawRemaining >= smokedRemaining {
		t.Fatalf("expected raw meat to degrade faster than smoked meat, raw=%.2f smoked=%.2f", rawRemaining, smokedRemaining)
	}
	if spoiled <= 0 {
		t.Fatalf("expected spoiled meat to accumulate from degradation")
	}
}

func TestAlaskaWinterPlantFilteringRemovesWarmSeasonFlora(t *testing.T) {
	scenario, ok := GetScenario(BuiltInScenarios(), "naa_alaska")
	if !ok || scenario.Climate == nil {
		t.Fatalf("expected alaska scenario climate profile")
	}
	plants := PlantsForBiomeSeason("tundra", PlantCategoryAny, SeasonWinter)
	filtered := filterPlantsForClimate(plants, scenario.Climate, SeasonWinter, -18)
	if len(filtered) == 0 {
		t.Fatalf("expected some winter-capable flora in tundra")
	}
	coldBerrySeen := false
	for _, plant := range filtered {
		id := strings.ToLower(plant.ID)
		if strings.Contains(id, "wild_garlic") {
			t.Fatalf("unexpected warm-season flora survived cold filter: %s", plant.ID)
		}
		if id == "crowberry" || id == "lingonberry" {
			coldBerrySeen = true
		}
	}
	if !coldBerrySeen {
		t.Fatalf("expected at least one cold-viable berry in alaska winter filtered plants")
	}
}
