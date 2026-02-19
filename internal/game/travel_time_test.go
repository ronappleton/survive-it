package game

import (
	"strings"
	"testing"
)

func newRunForTravelTime(t *testing.T) RunState {
	t.Helper()
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        8080,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Travel.PosX = 10
	run.Travel.PosY = 10
	return run
}

func TestTravelMoveAdvancesClock(t *testing.T) {
	run := newRunForTravelTime(t)
	startDay := run.Day
	startClock := run.ClockHours

	res, err := run.TravelMove(1, "east", 0.1)
	if err != nil {
		t.Fatalf("travel move: %v", err)
	}
	if !((run.Day > startDay) || (run.ClockHours > startClock)) {
		t.Fatalf("expected clock/day to advance, before day=%d clock=%.2f after day=%d clock=%.2f", startDay, startClock, run.Day, run.ClockHours)
	}
	if res.HoursSpent <= 0 {
		t.Fatalf("expected positive hours spent, got %.2f", res.HoursSpent)
	}
}

func TestTravelMoveMultiStepConsumesMoreTime(t *testing.T) {
	shortRun := newRunForTravelTime(t)
	shortRes, err := shortRun.TravelMove(1, "east", 0.1)
	if err != nil {
		t.Fatalf("short travel: %v", err)
	}

	longRun := newRunForTravelTime(t)
	longRes, err := longRun.TravelMove(1, "east", 0.5)
	if err != nil {
		t.Fatalf("long travel: %v", err)
	}

	if longRes.HoursSpent <= shortRes.HoursSpent {
		t.Fatalf("expected long travel hours %.2f > short %.2f", longRes.HoursSpent, shortRes.HoursSpent)
	}
}

func TestTravelMinutesForStepUphillHigherThanFlat(t *testing.T) {
	run := newRunForTravelTime(t)
	run.Topology = WorldTopology{
		Width:  2,
		Height: 2,
		Cells: []TopoCell{
			{Elevation: 10, Biome: TopoBiomeGrassland, Roughness: 2},
			{Elevation: 10, Biome: TopoBiomeGrassland, Roughness: 2},
			{Elevation: 34, Biome: TopoBiomeGrassland, Roughness: 2},
			{Elevation: 10, Biome: TopoBiomeGrassland, Roughness: 2},
		},
	}
	flat := TravelMinutesForStep(&run, 0, 0, 1, 0, &run.Players[0])
	uphill := TravelMinutesForStep(&run, 0, 0, 0, 1, &run.Players[0])
	if uphill <= flat {
		t.Fatalf("expected uphill minutes %d > flat %d", uphill, flat)
	}
}

func TestTravelMoveCanCrossTimeBlock(t *testing.T) {
	run := newRunForTravelTime(t)
	run.ClockHours = 17.9
	startBlock := run.CurrentTimeBlock()

	res, err := run.TravelMove(1, "east", 0.2)
	if err != nil {
		t.Fatalf("travel move: %v", err)
	}
	if res.BlocksCrossed < 1 {
		t.Fatalf("expected at least one time block crossing, got %d", res.BlocksCrossed)
	}
	if res.StartBlock != startBlock {
		t.Fatalf("expected start block %s, got %s", startBlock, res.StartBlock)
	}
	if res.EndBlock != run.CurrentTimeBlock() {
		t.Fatalf("expected end block %s to match run block %s", res.EndBlock, run.CurrentTimeBlock())
	}
}

func TestTravelStopsAtShoreWithoutWatercraft(t *testing.T) {
	run := newRunForTravelTime(t)
	run.Topology = WorldTopology{
		Width:  3,
		Height: 1,
		Cells: []TopoCell{
			{Biome: TopoBiomeGrassland},
			{Biome: TopoBiomeWetland, Flags: TopoFlagWater | TopoFlagLake},
			{Biome: TopoBiomeGrassland},
		},
	}
	run.CellStates = make([]CellState, 3)
	run.FogMask = []bool{true, true, true}
	run.Travel.PosX = 0
	run.Travel.PosY = 0
	startClock := run.ClockHours

	res, err := run.TravelMove(1, "east", 0.5)
	if err != nil {
		t.Fatalf("travel move: %v", err)
	}
	if res.StepsMoved != 0 {
		t.Fatalf("expected no movement into water without watercraft, moved %d", res.StepsMoved)
	}
	if !strings.Contains(strings.ToLower(res.StopReason), "shore") {
		t.Fatalf("expected shoreline stop reason, got: %q", res.StopReason)
	}
	if run.ClockHours != startClock {
		t.Fatalf("expected no time advance when blocked at shore, before=%.2f after=%.2f", startClock, run.ClockHours)
	}
}

func TestTravelCanEnterWaterWithWatercraft(t *testing.T) {
	run := newRunForTravelTime(t)
	run.Topology = WorldTopology{
		Width:  3,
		Height: 1,
		Cells: []TopoCell{
			{Biome: TopoBiomeGrassland},
			{Biome: TopoBiomeWetland, Flags: TopoFlagWater | TopoFlagLake},
			{Biome: TopoBiomeGrassland},
		},
	}
	run.CellStates = make([]CellState, 3)
	run.FogMask = []bool{true, true, true}
	run.Travel.PosX = 0
	run.Travel.PosY = 0
	run.CraftedItems = append(run.CraftedItems, "brush_raft")

	res, err := run.TravelMove(1, "east", 0.1)
	if err != nil {
		t.Fatalf("travel move: %v", err)
	}
	if res.StepsMoved < 1 {
		t.Fatalf("expected movement into water with watercraft, moved %d", res.StepsMoved)
	}
	if res.WatercraftUsed == "" {
		t.Fatalf("expected watercraft to be used")
	}
}
