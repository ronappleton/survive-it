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
