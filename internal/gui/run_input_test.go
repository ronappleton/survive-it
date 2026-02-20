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

	if ui.pendingIntent == nil {
		t.Fatalf("expected pending intent for missing go distance")
	}
	if ui.pendingIntent.OriginalVerb != "go" {
		t.Fatalf("expected pending verb go, got %q", ui.pendingIntent.OriginalVerb)
	}
	if len(ui.pendingIntent.FilledArgs) == 0 || ui.pendingIntent.FilledArgs[0] != "north" {
		t.Fatalf("expected pending filled direction north, got %+v", ui.pendingIntent.FilledArgs)
	}
	if !pendingMissingField(ui.pendingIntent, "distance") {
		t.Fatalf("expected pending missing field to include distance, got %+v", ui.pendingIntent.MissingFields)
	}
	if len(ui.runMessages) == 0 || !strings.Contains(strings.ToLower(ui.runMessages[len(ui.runMessages)-1]), "how far") {
		t.Fatalf("expected distance prompt in message log, got: %+v", ui.runMessages)
	}
}

func TestSubmitRunInputPendingDistanceRetriesThenCompletesTravel(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	ui.runInput = "go north"
	ui.submitRunInput()
	if ui.pendingIntent == nil {
		t.Fatalf("expected pending go intent")
	}

	ui.runInput = "far"
	ui.submitRunInput()
	if ui.pendingIntent == nil {
		t.Fatalf("expected pending intent to remain after invalid distance answer")
	}
	if !strings.Contains(strings.ToLower(ui.status), "distance required") {
		t.Fatalf("expected clearer retry guidance, got status: %q", ui.status)
	}

	beforeKm := ui.run.Travel.TotalKm

	ui.runInput = "500m"
	ui.submitRunInput()
	if ui.pendingIntent != nil {
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
	if ui.pendingIntent != nil {
		t.Fatalf("did not expect pending intent when distance is provided")
	}
	ui.processIntentQueue()
	if ui.run.Travel.TotalKm <= beforeKm {
		t.Fatalf("expected movement from immediate go command; before %.2f after %.2f", beforeKm, ui.run.Travel.TotalKm)
	}
}

func TestSubmitRunInputCraftMissingTypeUsesPendingIntent(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	ui.runInput = "craft"
	ui.submitRunInput()

	if ui.pendingIntent == nil {
		t.Fatalf("expected pending intent for craft missing item")
	}
	if ui.pendingIntent.OriginalVerb != "craft" {
		t.Fatalf("expected pending craft verb, got %q", ui.pendingIntent.OriginalVerb)
	}
	if !pendingMissingField(ui.pendingIntent, "craft_item") {
		t.Fatalf("expected missing craft_item field, got %+v", ui.pendingIntent.MissingFields)
	}
	if len(ui.runMessages) == 0 || !strings.Contains(strings.ToLower(ui.runMessages[len(ui.runMessages)-1]), "craft") {
		t.Fatalf("expected craft prompt in message log, got: %+v", ui.runMessages)
	}

	answer := "1"
	if len(ui.pendingIntent.Options) == 0 {
		craftables := game.CraftablesForBiome(ui.run.Scenario.Biome)
		if len(craftables) == 0 {
			t.Skip("no craftables available for test scenario")
		}
		answer = craftables[0].ID
	}

	ui.runInput = answer
	ui.submitRunInput()
	if ui.pendingIntent != nil {
		t.Fatalf("expected pending intent to clear after craft answer")
	}
	ui.processIntentQueue()
	if strings.Contains(strings.ToLower(ui.status), "unknown command") {
		t.Fatalf("expected resolved craft command to be handled, got status: %q", ui.status)
	}
}

func TestSubmitRunInputCancelClearsPendingIntent(t *testing.T) {
	ui := newGameUI(AppConfig{NoUpdate: true})
	ui.run = testRunState(t)

	ui.runInput = "go north"
	ui.submitRunInput()
	if ui.pendingIntent == nil {
		t.Fatalf("expected pending intent before cancel")
	}

	ui.runInput = "cancel"
	ui.submitRunInput()
	if ui.pendingIntent != nil {
		t.Fatalf("expected pending intent to clear on cancel")
	}
	if ui.status != "Cancelled." {
		t.Fatalf("expected cancel status, got %q", ui.status)
	}
	if len(ui.runMessages) == 0 || !strings.Contains(strings.ToLower(ui.runMessages[len(ui.runMessages)-1]), "cancelled") {
		t.Fatalf("expected cancel feedback in message log, got: %+v", ui.runMessages)
	}
}
