package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/appengine-ltd/survive-it/internal/game"
)

func withTempCWD(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}

func TestValidateDataFilePath(t *testing.T) {
	allowed := []string{
		defaultCustomScenariosFile,
		savePathForSlot(1),
		savePathForSlot(2),
		savePathForSlot(3),
	}
	for _, path := range allowed {
		if err := validateDataFilePath(path); err != nil {
			t.Fatalf("expected allowed path %q, got error: %v", path, err)
		}
	}

	rejected := []string{
		"/tmp/survive-it-save-1.json",
		"../survive-it-save-1.json",
		"nested/survive-it-save-1.json",
		"survive-it-save-.json",
		"survive-it-save-*.json",
	}
	for _, path := range rejected {
		if err := validateDataFilePath(path); err == nil {
			t.Fatalf("expected path %q to be rejected", path)
		}
	}
}

func TestParseNonNegativeIntStrict(t *testing.T) {
	if v, err := parseNonNegativeInt("42"); err != nil || v != 42 {
		t.Fatalf("expected parse 42, got v=%d err=%v", v, err)
	}
	if _, err := parseNonNegativeInt("12x"); err == nil {
		t.Fatalf("expected malformed numeric input to fail")
	}
	if _, err := parseNonNegativeInt("-1"); err == nil {
		t.Fatalf("expected negative numeric input to fail")
	}
}

func TestSaveAndLoadRunRoundTrip(t *testing.T) {
	withTempCWD(t)

	cfg := game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  game.ScenarioVancouverIslandID,
		PlayerCount: 2,
		RunLength:   game.RunLength{Days: 7},
		Seed:        123,
	}
	state, err := game.NewRunState(cfg)
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	path := savePathForSlot(1)
	if err := saveRunToFile(path, state); err != nil {
		t.Fatalf("save run: %v", err)
	}

	loaded, err := loadRunFromFile(path, game.BuiltInScenarios())
	if err != nil {
		t.Fatalf("load run: %v", err)
	}

	if loaded.Day != state.Day {
		t.Fatalf("day mismatch: got %d want %d", loaded.Day, state.Day)
	}
	if loaded.Config.Mode != state.Config.Mode {
		t.Fatalf("mode mismatch: got %s want %s", loaded.Config.Mode, state.Config.Mode)
	}
	if len(loaded.Players) != len(state.Players) {
		t.Fatalf("player count mismatch: got %d want %d", len(loaded.Players), len(state.Players))
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat saved file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected save file mode 0600, got %o", info.Mode().Perm())
	}
}

func TestLoadCustomScenariosLegacyCompatibility(t *testing.T) {
	withTempCWD(t)

	library := customScenarioLibrary{
		FormatVersion: 1,
		Scenarios: []game.Scenario{
			{
				ID:                 "custom_test",
				Name:               "Legacy Scenario",
				DefaultDays:        30,
				SeasonSets:         []game.SeasonSet{{ID: game.SeasonSetAloneDefaultID, Phases: []game.SeasonPhase{{Season: game.SeasonAutumn, Days: 0}}}},
				DefaultSeasonSetID: game.SeasonSetAloneDefaultID,
			},
		},
	}
	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy library: %v", err)
	}
	if err := os.WriteFile(defaultCustomScenariosFile, data, 0o600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	records, err := loadCustomScenarios(defaultCustomScenariosFile)
	if err != nil {
		t.Fatalf("load custom scenarios: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].PreferredMode != game.ModeAlone {
		t.Fatalf("expected legacy preferred mode default to Alone, got %s", records[0].PreferredMode)
	}
}

func TestReadDataFileSizeLimit(t *testing.T) {
	withTempCWD(t)

	path := savePathForSlot(1)
	tooLarge := make([]byte, maxSaveFileBytes+1)
	if err := os.WriteFile(path, tooLarge, 0o600); err != nil {
		t.Fatalf("write oversized save file: %v", err)
	}

	_, err := readDataFile(path, maxSaveFileBytes)
	if err == nil {
		t.Fatalf("expected oversized read to fail")
	}

	// Ensure file path remains local and expected.
	if _, err := filepath.Abs(path); err != nil {
		t.Fatalf("abs path: %v", err)
	}
}

func TestValidateRunConfigWithScenariosAcceptsXLMode(t *testing.T) {
	cfg := game.RunConfig{
		Mode:        game.ModeNakedAndAfraidXL,
		ScenarioID:  "naaxl_colombia_40",
		PlayerCount: 4,
		RunLength:   game.RunLength{Days: 40},
	}

	if err := validateRunConfigWithScenarios(cfg, game.BuiltInScenarios()); err != nil {
		t.Fatalf("expected XL mode config to validate, got error: %v", err)
	}
}

func TestValidateRunConfigWithScenariosRejectsWrongModeScenarioPair(t *testing.T) {
	cfg := game.RunConfig{
		Mode:        game.ModeAlone,
		ScenarioID:  "naaxl_colombia_40",
		PlayerCount: 1,
		RunLength:   game.RunLength{Days: 60},
	}

	if err := validateRunConfigWithScenarios(cfg, game.BuiltInScenarios()); err == nil {
		t.Fatalf("expected mode/scenario mismatch to fail validation")
	}
}

func TestEnsureSetupPlayersAppliesSeriesKitLimits(t *testing.T) {
	m := newMenuModel(AppConfig{})
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

	m.setup.players[0].KitLimit = 10
	m.setup.players[0].Kit = []game.KitItem{
		game.KitHatchet,
		game.KitSixInchKnife,
		game.KitParacord50ft,
	}

	m.setup.modeIdx = selectedModeIndex(game.ModeNakedAndAfraid)
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

	if got := m.setup.players[0].KitLimit; got != 1 {
		t.Fatalf("expected Naked and Afraid kit limit 1, got %d", got)
	}
	if got := len(m.setup.players[0].Kit); got != 1 {
		t.Fatalf("expected player kit to be trimmed to 1 item, got %d", got)
	}
	if len(m.setup.issuedKit) == 0 {
		t.Fatalf("expected scenario-based issued kit recommendation")
	}
}

func TestStartRunFromSetupUsesConfiguredPlayersAndIssuedKit(t *testing.T) {
	m := newMenuModel(AppConfig{})
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

	m.setup.modeIdx = selectedModeIndex(game.ModeAlone)
	m.setup.playerCountIdx = 1 // 2 players
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

	m.setup.players[0].Name = "Ron"
	m.setup.players[0].Sex = game.SexMale
	m.setup.players[0].BodyType = game.BodyTypeMale
	m.setup.players[0].KitLimit = 2
	m.setup.players[0].Kit = []game.KitItem{game.KitHatchet, game.KitFerroRod}

	m.setup.players[1].Name = "Alex"
	m.setup.players[1].Sex = game.SexNonBinary
	m.setup.players[1].BodyType = game.BodyTypeNeutral
	m.setup.players[1].KitLimit = 2
	m.setup.players[1].Kit = []game.KitItem{game.KitParacord50ft}

	m.setup.issuedKit = []game.KitItem{game.KitCookingPot, game.KitCanteen}
	m.setup.issuedCustom = true

	gotModel, _ := m.startRunFromSetup()
	got, ok := gotModel.(menuModel)
	if !ok {
		t.Fatalf("expected menuModel, got %T", gotModel)
	}
	if got.run == nil {
		t.Fatalf("expected run to start")
	}
	if len(got.run.Config.Players) != 2 {
		t.Fatalf("expected 2 configured players, got %d", len(got.run.Config.Players))
	}
	if got.run.Config.Players[0].Name != "Ron" {
		t.Fatalf("expected first configured player name Ron, got %q", got.run.Config.Players[0].Name)
	}
	if got.run.Config.Players[1].Name != "Alex" {
		t.Fatalf("expected second configured player name Alex, got %q", got.run.Config.Players[1].Name)
	}
	if len(got.run.Config.IssuedKit) != 2 {
		t.Fatalf("expected issued kit to carry into run config, got %d items", len(got.run.Config.IssuedKit))
	}
}
