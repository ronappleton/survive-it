package gui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/appengine-ltd/survive-it/internal/game"
)

func withTempCWD(t *testing.T) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
}

func TestPlayerProfilesSaveLoadRoundTrip(t *testing.T) {
	withTempCWD(t)
	profiles := []playerProfile{
		newPlayerProfile("Ronan", game.ModeAlone),
		newPlayerProfile("Aria", game.ModeNakedAndAfraid),
	}
	profiles[0].RunsPlayed = 3
	profiles[0].TotalDaysSurvived = 210
	activeID := profiles[1].ID
	path := filepath.Join("profiles.json")
	if err := savePlayerProfiles(path, profiles, activeID); err != nil {
		t.Fatalf("save profiles: %v", err)
	}
	loaded, loadedActiveID, err := loadPlayerProfiles(path)
	if err != nil {
		t.Fatalf("load profiles: %v", err)
	}
	if len(loaded) != len(profiles) {
		t.Fatalf("expected %d profiles, got %d", len(profiles), len(loaded))
	}
	if loadedActiveID != activeID {
		t.Fatalf("expected active profile %q, got %q", activeID, loadedActiveID)
	}
	if loaded[0].Name != profiles[0].Name || loaded[1].Name != profiles[1].Name {
		t.Fatalf("loaded names mismatch: %+v", loaded)
	}
}

func TestInitPlayerProfilesCreatesDefaultAndAppliesToSetup(t *testing.T) {
	withTempCWD(t)
	ui := newGameUI(AppConfig{NoUpdate: true})
	if len(ui.profiles) == 0 {
		t.Fatalf("expected at least one profile")
	}
	if ui.selectedProfileID == "" {
		t.Fatalf("expected selected profile id")
	}
	if len(ui.pcfg.Players) == 0 {
		t.Fatalf("expected setup players to be initialised")
	}
	profile, ok := ui.selectedProfile()
	if !ok {
		t.Fatalf("expected selected profile")
	}
	if ui.pcfg.Players[0].Hunting != profile.Config.Hunting {
		t.Fatalf("expected setup player hunting %d, got %d", profile.Config.Hunting, ui.pcfg.Players[0].Hunting)
	}
}

func TestSelectProfileAppliesPrimaryPlayerConfig(t *testing.T) {
	withTempCWD(t)
	ui := newGameUI(AppConfig{NoUpdate: true})
	a := newPlayerProfile("Alpha", game.ModeAlone)
	a.Config.Hunting = 10
	b := newPlayerProfile("Bravo", game.ModeAlone)
	b.ID = "bravo"
	b.Config.Hunting = 47
	ui.profiles = []playerProfile{a, b}
	ui.selectedProfileID = b.ID
	ui.applySelectedProfileToSetupPrimary()
	ui.ensureSetupPlayers()
	if got := ui.pcfg.Players[0].Hunting; got != 47 {
		t.Fatalf("expected selected profile hunting 47, got %d", got)
	}
}

func TestPersistActiveRunProfileProgress(t *testing.T) {
	withTempCWD(t)
	ui := newGameUI(AppConfig{NoUpdate: true})
	run, err := game.NewRunState(game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 1,
		Players:     []game.PlayerConfig{ui.pcfg.Players[0]},
		RunLength:   game.RunLength{Days: 30},
		Seed:        1337,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	ui.run = &run
	ui.runProfileID = ui.selectedProfileID
	idx := ui.selectedProfileIndex()
	if idx < 0 {
		t.Fatalf("expected selected profile index")
	}
	beforeRuns := ui.profiles[idx].RunsPlayed
	ui.run.Players[0].Hunting = 61
	ui.run.Day = 12
	ui.persistActiveRunProfileProgress()
	if ui.profiles[idx].RunsPlayed != beforeRuns+1 {
		t.Fatalf("expected runs played %d, got %d", beforeRuns+1, ui.profiles[idx].RunsPlayed)
	}
	if ui.profiles[idx].Config.Hunting != 61 {
		t.Fatalf("expected persisted hunting 61, got %d", ui.profiles[idx].Config.Hunting)
	}
	if ui.profiles[idx].TotalDaysSurvived < 12 {
		t.Fatalf("expected persisted days survived to increase, got %d", ui.profiles[idx].TotalDaysSurvived)
	}
}
