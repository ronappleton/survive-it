package game

import (
	"strings"
	"testing"
)

func newRunForCommands(t *testing.T) RunState {
	t.Helper()

	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        909,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	run.Players[0].Kit = []KitItem{
		KitParacord50ft,
		KitFirstAidKit,
		KitEmergencyRations,
	}
	run.Config.IssuedKit = []KitItem{KitTarp}
	return run
}

func TestRunCommandActionsListsEquipment(t *testing.T) {
	run := newRunForCommands(t)

	res := run.ExecuteRunCommand("actions")
	if !res.Handled {
		t.Fatalf("expected command to be handled")
	}
	if !strings.Contains(res.Message, "paracord") {
		t.Fatalf("expected paracord in actions list, got: %s", res.Message)
	}
}

func TestRunCommandUseParacordTieSticksTogether(t *testing.T) {
	run := newRunForCommands(t)
	run.Players[0].Morale = 90
	beforeMorale := run.Players[0].Morale

	res := run.ExecuteRunCommand("use paracord tie sticks together p1")
	if !res.Handled {
		t.Fatalf("expected use command to be handled")
	}
	if !strings.Contains(res.Message, "tie_sticks_together") {
		t.Fatalf("expected tie_sticks_together action confirmation, got: %s", res.Message)
	}
	if run.Players[0].Morale <= beforeMorale {
		t.Fatalf("expected morale to increase after successful paracord action")
	}
}

func TestRunCommandUseTreatWoundRemovesAilment(t *testing.T) {
	run := newRunForCommands(t)
	run.Players[0].Ailments = []Ailment{
		{Type: AilmentVomiting, Name: "Vomiting", DaysRemaining: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 3},
	}

	res := run.ExecuteRunCommand("use first aid kit treat wound")
	if !res.Handled {
		t.Fatalf("expected command to be handled")
	}
	if len(run.Players[0].Ailments) != 0 {
		t.Fatalf("expected first-aid action to remove one ailment")
	}
	if !strings.Contains(strings.ToLower(res.Message), "treated") {
		t.Fatalf("expected treated message, got: %s", res.Message)
	}
}

func TestRunCommandUseRationsAddsNutrition(t *testing.T) {
	run := newRunForCommands(t)
	beforeCalories := run.Players[0].Nutrition.CaloriesKcal

	res := run.ExecuteRunCommand("use rations eat")
	if !res.Handled {
		t.Fatalf("expected ration command to be handled")
	}
	if run.Players[0].Nutrition.CaloriesKcal <= beforeCalories {
		t.Fatalf("expected ration command to increase nutrition calories")
	}
}

func TestRunCommandUseRejectsMissingItem(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 20},
		Seed:        111,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	res := run.ExecuteRunCommand("use paracord tie sticks together")
	if !res.Handled {
		t.Fatalf("expected command to be handled with validation message")
	}
	if !strings.Contains(strings.ToLower(res.Message), "not in player") {
		t.Fatalf("expected missing-item validation, got: %s", res.Message)
	}
}

func TestRunCommandGoWithoutDistancePrompts(t *testing.T) {
	run := newRunForCommands(t)

	res := run.ExecuteRunCommand("go north")
	if !res.Handled {
		t.Fatalf("expected command to be handled")
	}
	if !strings.Contains(strings.ToLower(res.Message), "how far?") {
		t.Fatalf("expected distance prompt, got: %s", res.Message)
	}
}

func TestRunCommandGoDistanceUnits(t *testing.T) {
	run := newRunForCommands(t)

	before := run.Travel.TotalKm
	beforeClock := run.ClockHours
	res := run.ExecuteRunCommand("go north 500m")
	if !res.Handled {
		t.Fatalf("expected command to be handled")
	}
	if run.Travel.TotalKm <= before {
		t.Fatalf("expected movement for 500m command")
	}
	if run.ClockHours <= beforeClock && run.Day == 1 {
		t.Fatalf("expected clock/day to advance for travel command")
	}

	before = run.Travel.TotalKm
	res = run.ExecuteRunCommand("go north 2km")
	if !res.Handled {
		t.Fatalf("expected command to be handled")
	}
	if run.Travel.TotalKm <= before {
		t.Fatalf("expected movement for 2km command")
	}
}

func TestRunCommandLookLeftProvidesDirectionalInfo(t *testing.T) {
	run := newRunForCommands(t)
	run.Travel.Direction = "north"

	res := run.ExecuteRunCommand("look left")
	if !res.Handled {
		t.Fatalf("expected look command to be handled")
	}
	msg := strings.ToLower(res.Message)
	if !strings.Contains(msg, "left") && !strings.Contains(msg, "west") {
		t.Fatalf("expected directional description for look left, got: %s", res.Message)
	}
}

func TestRunCommandLookCloserAtPlantsFindsSpecificPlant(t *testing.T) {
	run := newRunForCommands(t)

	res := run.ExecuteRunCommand("look closer at plants")
	if !res.Handled {
		t.Fatalf("expected look closer command to be handled")
	}
	msg := strings.ToLower(res.Message)
	if !strings.Contains(msg, "locate") || !strings.Contains(msg, "plant") {
		t.Fatalf("expected closer plant detail, got: %s", res.Message)
	}
}

func TestGoCommandAppendsForwardViewSummary(t *testing.T) {
	run := newRunForCommands(t)

	res := run.ExecuteRunCommand("go north 500m")
	if !res.Handled {
		t.Fatalf("expected go command to be handled")
	}
	msg := strings.ToLower(res.Message)
	if !strings.Contains(msg, "looking in front of you") && !strings.Contains(msg, "in front of you") {
		t.Fatalf("expected forward visibility summary after travel, got: %s", res.Message)
	}
}

func TestGoCommandStopsAtShorelineAndReportsIt(t *testing.T) {
	run := newRunForCommands(t)
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
	run.Travel.Direction = "east"

	res := run.ExecuteRunCommand("go east 500m")
	if !res.Handled {
		t.Fatalf("expected go command handled")
	}
	msg := strings.ToLower(res.Message)
	if !strings.Contains(msg, "shore") {
		t.Fatalf("expected shoreline stop message, got: %s", res.Message)
	}
	if run.Travel.PosX != 0 || run.Travel.PosY != 0 {
		t.Fatalf("expected player to remain on shore, pos=(%d,%d)", run.Travel.PosX, run.Travel.PosY)
	}
}

func TestRunCommandEnterShelter(t *testing.T) {
	run := newRunForCommands(t)

	// Fails when no shelter is built
	res := run.ExecuteRunCommand("enter shelter")
	if !res.Handled || !strings.Contains(res.Message, "no shelter built here") {
		t.Fatalf("expected to fail entering missing shelter, got: %s", res.Message)
	}

	// Build a shelter
	run.Shelter.Type = "debris_hut"
	run.Shelter.Durability = 100

	res = run.ExecuteRunCommand("enter shelter")
	if !res.Handled || !strings.Contains(res.Message, "crawls inside") {
		t.Fatalf("expected to enter shelter smoothly, got: %s", res.Message)
	}

	if run.Players[0].MicroLocation != LocationInsideShelter {
		t.Fatalf("expected micro-location to be inside")
	}

	// Fails when already inside
	res = run.ExecuteRunCommand("enter shelter")
	if !res.Handled || !strings.Contains(res.Message, "already inside") {
		t.Fatalf("expected error while already inside, got: %s", res.Message)
	}
}

func TestRunCommandExitShelter(t *testing.T) {
	run := newRunForCommands(t)
	run.Shelter.Type = "debris_hut"
	run.Shelter.Durability = 100

	// Fails when already outside
	res := run.ExecuteRunCommand("exit shelter")
	if !res.Handled || !strings.Contains(res.Message, "already outside") {
		t.Fatalf("expected error when exiting while outside, got: %s", res.Message)
	}

	// Move inside
	run.Players[0].MicroLocation = LocationInsideShelter

	// Succeeds exiting
	res = run.ExecuteRunCommand("exit shelter")
	if !res.Handled || !strings.Contains(res.Message, "steps outside") {
		t.Fatalf("expected to exit shelter smoothly, got: %s", res.Message)
	}

	if run.Players[0].MicroLocation != LocationOutside {
		t.Fatalf("expected micro-location to be outside")
	}
}
