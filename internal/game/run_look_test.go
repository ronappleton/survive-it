package game

import (
	"strings"
	"testing"
)

func TestTundraLookDescriptionIsColdCoherent(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  "naa_alaska",
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        7412,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}

	cells := make([]TopoCell, 9)
	for i := range cells {
		cells[i] = TopoCell{Biome: TopoBiomeTundra}
	}
	// North/front cell has water so the text path exercises frozen-water wording.
	cells[1] = TopoCell{Biome: TopoBiomeTundra, Flags: TopoFlagWater | TopoFlagLake}
	run.Topology = WorldTopology{Width: 3, Height: 3, Cells: cells}
	run.CellStates = make([]CellState, len(cells))
	run.FogMask = make([]bool, len(cells))
	for i := range run.FogMask {
		run.FogMask[i] = true
	}
	run.Travel.PosX = 1
	run.Travel.PosY = 1
	run.Travel.Direction = "north"
	run.Day = 2
	run.Weather = WeatherState{Day: 2, Type: WeatherSnow, TemperatureC: -18}

	msg := strings.ToLower(run.describeDirectionalView(1, "front", false, ""))
	if !strings.Contains(msg, "tundra") {
		t.Fatalf("expected tundra description, got: %q", msg)
	}
	if strings.Contains(msg, "desert") {
		t.Fatalf("unexpected desert description in tundra view: %q", msg)
	}
	if !strings.Contains(msg, "frozen water") {
		t.Fatalf("expected frozen-water wording, got: %q", msg)
	}
	if strings.Contains(msg, "mosquito") {
		t.Fatalf("unexpected mosquito mention in freezing look text: %q", msg)
	}
}

func TestLookCloserPlantsRespectsColdSeasonFiltering(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  "naa_alaska",
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        8821,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}

	cells := make([]TopoCell, 9)
	for i := range cells {
		cells[i] = TopoCell{Biome: TopoBiomeTundra}
	}
	run.Topology = WorldTopology{Width: 3, Height: 3, Cells: cells}
	run.CellStates = make([]CellState, len(cells))
	run.FogMask = make([]bool, len(cells))
	for i := range run.FogMask {
		run.FogMask[i] = true
	}
	run.Travel.PosX = 1
	run.Travel.PosY = 1
	run.Travel.Direction = "north"
	run.Day = 5
	run.Weather = WeatherState{Day: 5, Type: WeatherCloudy, TemperatureC: -15}

	msg := strings.ToLower(run.describeDirectionalView(1, "front", true, "plants"))
	if strings.Contains(msg, "wild garlic") {
		t.Fatalf("unexpected warm-season plant in alaska winter look: %q", msg)
	}
}
