package ui

import (
	"strings"
	"testing"

	"github.com/appengine-ltd/survive-it/internal/game"
	tea "github.com/charmbracelet/bubbletea"
)

func testRunState(t *testing.T) game.RunState {
	t.Helper()
	state, err := game.NewRunState(game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   game.RunLength{Days: 7},
		Seed:        11,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	return state
}

func TestRunBodyTextUsesMessageHistory(t *testing.T) {
	m := newMenuModel(AppConfig{})
	state := testRunState(t)
	m.run = &state
	m.runMessages = []string{
		"[00:00:01] Run started",
		"[00:00:02] Day 2 started",
	}

	got := m.bodyText()
	if !strings.Contains(got, "Message History") {
		t.Fatalf("expected message history header in run body")
	}
	if !strings.Contains(got, "Run started") {
		t.Fatalf("expected history entry in run body")
	}
}

func TestUpdateRunShiftPOpensPlayerUX(t *testing.T) {
	m := newMenuModel(AppConfig{})
	state := testRunState(t)
	m.run = &state
	m.screen = screenRun

	gotModel, _ := m.updateRun(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	got := gotModel.(menuModel)
	if got.screen != screenRunPlayers {
		t.Fatalf("expected Shift+P to open run players screen, got %v", got.screen)
	}
}

func TestSubmitRunInputHelpOpensCommandLibrary(t *testing.T) {
	m := newMenuModel(AppConfig{})
	state := testRunState(t)
	m.run = &state
	m.screen = screenRun
	m.runInput = "help"

	gotModel, _ := m.submitRunInput()
	got := gotModel.(menuModel)
	if got.screen != screenRunCommandLibrary {
		t.Fatalf("expected help command to open run command library, got %v", got.screen)
	}
}
