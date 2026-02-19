package gui

import (
	"strings"
	"testing"

	"github.com/appengine-ltd/survive-it/internal/game"
)

func testRunState(t *testing.T) *game.RunState {
	t.Helper()
	run, err := game.NewRunState(game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   game.RunLength{Days: 20},
		Seed:        4242,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	return &run
}

func TestSubmitRunInputGoWithoutDistanceQueuesPendingPrompt(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	ui.runInput = "move north"
	ui.submitRunInput()

	if ui.pendingGo == nil {
		t.Fatalf("expected pending go intent")
	}
	if ui.pendingGo.Direction != "north" {
		t.Fatalf("expected pending direction north, got %q", ui.pendingGo.Direction)
	}
	if len(ui.runMessages) == 0 || !strings.Contains(strings.ToLower(ui.runMessages[len(ui.runMessages)-1]), "how far?") {
		t.Fatalf("expected distance prompt in message log, got: %+v", ui.runMessages)
	}
}

func TestSubmitRunInputPendingDistanceCompletesTravel(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	ui.runInput = "go north"
	ui.submitRunInput()
	if ui.pendingGo == nil {
		t.Fatalf("expected pending go intent")
	}
	beforeKm := ui.run.Travel.TotalKm

	ui.runInput = "500m"
	ui.submitRunInput()
	if ui.pendingGo != nil {
		t.Fatalf("expected pending go intent to be cleared after distance input")
	}
	ui.processIntentQueue()
	if ui.run.Travel.TotalKm <= beforeKm {
		t.Fatalf("expected travel progress after distance input; before %.2f after %.2f", beforeKm, ui.run.Travel.TotalKm)
	}
}

func TestSubmitRunInputGoWithDistanceTravelsImmediately(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	beforeKm := ui.run.Travel.TotalKm
	ui.runInput = "go north 2km"
	ui.submitRunInput()
	if ui.pendingGo != nil {
		t.Fatalf("did not expect pending go intent when distance is provided")
	}
	ui.processIntentQueue()
	if ui.run.Travel.TotalKm <= beforeKm {
		t.Fatalf("expected movement from immediate go command; before %.2f after %.2f", beforeKm, ui.run.Travel.TotalKm)
	}
}
