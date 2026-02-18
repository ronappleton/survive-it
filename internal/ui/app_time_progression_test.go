package ui

import (
	"testing"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
)

func TestClockTickAutoDayProgressionAppliesMetabolism(t *testing.T) {
	m := newMenuModel(AppConfig{})
	state, err := game.NewRunState(game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   game.RunLength{Days: 10},
		Seed:        42,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	m.run = &state
	m.screen = screenRun
	m.opts.dayHours = 1
	m.runPlayedFor = 0
	m.lastTickAt = time.Date(2026, 2, 18, 10, 0, 0, 0, time.UTC)

	beforeDay := m.run.Day
	beforeEnergy := m.run.Players[0].Energy
	beforeCalories := m.run.Players[0].CaloriesReserveKcal

	updated, _ := m.Update(clockTickMsg{at: m.lastTickAt.Add(time.Hour)})
	got := updated.(menuModel)

	if got.run.Day != beforeDay+1 {
		t.Fatalf("expected auto day advance, before=%d after=%d", beforeDay, got.run.Day)
	}
	if got.runPlayedFor != 0 {
		t.Fatalf("expected runPlayedFor reset to 0 after full day, got %s", got.runPlayedFor)
	}
	if got.run.Players[0].Energy >= beforeEnergy {
		t.Fatalf("expected energy to decrease after progressed day, before=%d after=%d", beforeEnergy, got.run.Players[0].Energy)
	}
	if got.run.Players[0].CaloriesReserveKcal >= beforeCalories {
		t.Fatalf("expected calorie reserve to deplete after progressed day, before=%d after=%d", beforeCalories, got.run.Players[0].CaloriesReserveKcal)
	}
}

func TestClockTickPartialDayDoesNotAdvanceDay(t *testing.T) {
	m := newMenuModel(AppConfig{})
	state, err := game.NewRunState(game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   game.RunLength{Days: 10},
		Seed:        24,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}
	m.run = &state
	m.screen = screenRun
	m.opts.dayHours = 2
	m.runPlayedFor = 0
	m.lastTickAt = time.Date(2026, 2, 18, 10, 0, 0, 0, time.UTC)

	beforeDay := m.run.Day
	updated, _ := m.Update(clockTickMsg{at: m.lastTickAt.Add(30 * time.Minute)})
	got := updated.(menuModel)

	if got.run.Day != beforeDay {
		t.Fatalf("expected no day advance on partial duration, before=%d after=%d", beforeDay, got.run.Day)
	}
	if got.runPlayedFor <= 0 {
		t.Fatalf("expected runPlayedFor to accumulate on partial day")
	}
}
