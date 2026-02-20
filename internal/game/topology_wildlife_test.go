package game

import (
	"strings"
	"testing"
)

func TestDeterministicEncounterRollStable(t *testing.T) {
	a := deterministicEncounterRoll(44, 7, 9, 3, TimeBlockDay, "hunt", 2, "species")
	b := deterministicEncounterRoll(44, 7, 9, 3, TimeBlockDay, "hunt", 2, "species")
	if a != b {
		t.Fatalf("expected same deterministic roll, got %f vs %f", a, b)
	}
	if a < 0 || a > 1 {
		t.Fatalf("expected roll in [0,1], got %f", a)
	}
}

func TestFogRevealPermanenceAlone(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 30},
		Seed:        2001,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.EnsureTopology()
	x, y := run.CurrentMapPosition()
	if !run.IsRevealed(x, y) {
		t.Fatalf("start cell should be revealed")
	}
	targetX, targetY := x+2, y
	run.RevealFog(targetX, targetY, 1)
	if !run.IsRevealed(targetX, targetY) {
		t.Fatalf("target cell expected revealed")
	}
	run.RevealFog(x-2, y, 1)
	if !run.IsRevealed(targetX, targetY) {
		t.Fatalf("revealed cell should remain revealed")
	}
}

func TestGenerateWorldTopologyDeterministicForSameSeed(t *testing.T) {
	a := GenerateWorldTopology(3001, "temperate_rainforest", 36, 36)
	b := GenerateWorldTopology(3001, "temperate_rainforest", 36, 36)
	if a.Width != b.Width || a.Height != b.Height {
		t.Fatalf("expected same dimensions, got %dx%d vs %dx%d", a.Width, a.Height, b.Width, b.Height)
	}
	if len(a.Cells) != len(b.Cells) {
		t.Fatalf("expected same cell count")
	}
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			t.Fatalf("topology mismatch at cell %d: %+v vs %+v", i, a.Cells[i], b.Cells[i])
		}
	}
}

func TestHuntPressureReducesPreyEncounterRate(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 30},
		Seed:        4001,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.EnsureTopology()
	x, y := run.CurrentMapPosition()
	idx, ok := run.topoIndex(x, y)
	if !ok {
		t.Fatalf("expected valid start index")
	}
	if len(run.CellStates) <= idx {
		t.Fatalf("missing cell state for index")
	}

	countPrey := func(huntPressure uint8) int {
		run.CellStates[idx] = CellState{HuntPressure: huntPressure}
		prey := 0
		for i := 0; i < 350; i++ {
			event, ok := run.RollWildlifeEncounter(1, x, y, "hunt", i)
			if ok && event.Prey {
				prey++
			}
		}
		return prey
	}

	low := countPrey(0)
	high := countPrey(220)
	if high >= low {
		t.Fatalf("expected high hunt pressure to reduce prey encounters, low=%d high=%d", low, high)
	}
}

func TestTopologySizeForScenarioUsesScenarioConfig(t *testing.T) {
	w, h := topologySizeForScenario(ModeNakedAndAfraid, Scenario{
		MapWidthCells:  92,
		MapHeightCells: 101,
	})
	if w != 92 || h != 101 {
		t.Fatalf("expected explicit scenario size to be used, got %dx%d", w, h)
	}
}

func TestTopologySizeForScenarioClampsByMode(t *testing.T) {
	w, h := topologySizeForScenario(ModeAlone, Scenario{
		MapWidthCells:  500,
		MapHeightCells: 9,
	})
	if w != 46 || h != 28 {
		t.Fatalf("expected isolation_protocol clamp to 46x28, got %dx%d", w, h)
	}
}

func TestSwampBiomeProducesMoreInsectEncountersThanDesert(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 30},
		Seed:        5101,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}

	cells := make([]TopoCell, 9)
	for i := range cells {
		cells[i] = TopoCell{
			Biome: TopoBiomeForest,
		}
	}
	run.Topology = WorldTopology{
		Width:  3,
		Height: 3,
		Cells:  cells,
	}
	run.CellStates = make([]CellState, 9)
	run.FogMask = make([]bool, 9)
	run.Travel.PosX = 1
	run.Travel.PosY = 1
	run.ClockHours = 19.5 // dusk
	run.Day = 4

	idx, ok := run.topoIndex(1, 1)
	if !ok {
		t.Fatalf("expected center topology index")
	}

	countInsects := func(biome uint8) int {
		run.Topology.Cells[idx] = TopoCell{Biome: biome}
		run.CellStates[idx] = CellState{}
		hits := 0
		for i := 0; i < 900; i++ {
			ev, ok := run.RollWildlifeEncounter(1, 1, 1, "forage", i)
			if ok && ev.Channel == "insect" {
				hits++
			}
		}
		return hits
	}

	desert := countInsects(TopoBiomeDesert)
	swamp := countInsects(TopoBiomeSwamp)
	if swamp <= desert {
		t.Fatalf("expected swamp to produce more insect encounters than desert, desert=%d swamp=%d", desert, swamp)
	}
}

func TestWildlifeEncounterSpeciesIDsResolveToAnimalCatalog(t *testing.T) {
	animalIDs := map[string]bool{}
	for _, animal := range AnimalCatalog() {
		animalIDs[animal.ID] = true
	}

	ids := wildlifeEncounterSpeciesIDs()
	if len(ids) == 0 {
		t.Fatalf("expected wildlife encounter species ids")
	}
	for _, id := range ids {
		if !animalIDs[id] {
			t.Fatalf("encounter species id %q missing from AnimalCatalog", id)
		}
	}
}

func TestAlaskaClimateTopologyNeverGeneratesDesert(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  "naa_alaska",
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        6188,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.EnsureTopology()
	for i, cell := range run.Topology.Cells {
		if cell.Biome == TopoBiomeDesert {
			t.Fatalf("desert biome generated in alaska climate at cell %d", i)
		}
	}
}

func TestColdWeatherSuppressesMosquitoEncounters(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  "naa_alaska",
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        7782,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.Topology = WorldTopology{
		Width: 1, Height: 1,
		Cells: []TopoCell{{Biome: TopoBiomeWetland, Flags: TopoFlagWater}},
	}
	run.CellStates = make([]CellState, 1)
	run.FogMask = []bool{true}
	run.Travel.PosX, run.Travel.PosY = 0, 0
	run.Day = 2
	run.ClockHours = 13
	run.Weather = WeatherState{Day: 2, Type: WeatherSnow, TemperatureC: -18}

	for i := 0; i < 1200; i++ {
		ev, ok := run.RollWildlifeEncounter(1, 0, 0, "forage", i)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(ev.Species), "mosquito") {
			t.Fatalf("unexpected mosquito encounter at -18C: %+v", ev)
		}
	}
}

func TestAlaskaWinterFiltersWarmBirdsFromEncounters(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  "naa_alaska",
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        9122,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.Topology = WorldTopology{
		Width: 1, Height: 1,
		Cells: []TopoCell{{Biome: TopoBiomeGrassland}},
	}
	run.CellStates = make([]CellState, 1)
	run.FogMask = []bool{true}
	run.Travel.PosX, run.Travel.PosY = 0, 0
	run.Day = 3
	run.ClockHours = 9
	run.Weather = WeatherState{Day: 3, Type: WeatherSnow, TemperatureC: -14}

	for i := 0; i < 1600; i++ {
		ev, ok := run.RollWildlifeEncounter(1, 0, 0, "hunt", i)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(ev.Species), "quail") {
			t.Fatalf("unexpected warm-season bird encounter in alaska winter: %+v", ev)
		}
	}
}
