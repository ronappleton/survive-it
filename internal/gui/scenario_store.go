package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/appengine-ltd/survive-it/internal/game"
)

const defaultCustomScenariosFile = "survive-it-scenarios.json"

type customScenarioRecord struct {
	Scenario      game.Scenario `json:"scenario"`
	PreferredMode game.GameMode `json:"preferred_mode"`
	Notes         string        `json:"notes,omitempty"`
}

type customScenarioLibrary struct {
	FormatVersion int                    `json:"format_version"`
	Custom        []customScenarioRecord `json:"custom,omitempty"`
	Scenarios     []game.Scenario        `json:"scenarios,omitempty"`
}

func loadCustomScenarios(path string) ([]game.Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var lib customScenarioLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, err
	}

	items := make([]game.Scenario, 0, len(lib.Custom)+len(lib.Scenarios))
	if len(lib.Custom) > 0 {
		for _, record := range lib.Custom {
			s := record.Scenario
			mode := record.PreferredMode
			if mode == "" && len(s.SupportedModes) > 0 {
				mode = s.SupportedModes[0]
			}
			if mode == "" {
				mode = game.ModeAlone
			}
			normalizeScenarioForMode(&s, mode)
			items = append(items, s)
		}
	} else if len(lib.Scenarios) > 0 {
		for _, s := range lib.Scenarios {
			mode := game.ModeAlone
			if len(s.SupportedModes) > 0 {
				mode = s.SupportedModes[0]
			}
			normalizeScenarioForMode(&s, mode)
			items = append(items, s)
		}
	}

	dedup := map[game.ScenarioID]game.Scenario{}
	for _, s := range items {
		if strings.TrimSpace(s.Name) == "" {
			continue
		}
		if s.ID == "" {
			s.ID = game.ScenarioID(generateScenarioID(s.Name, s.SupportedModes[0], items))
		}
		dedup[s.ID] = s
	}

	out := make([]game.Scenario, 0, len(dedup))
	for _, s := range dedup {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func saveCustomScenarios(path string, scenarios []game.Scenario) error {
	records := make([]customScenarioRecord, 0, len(scenarios))
	for _, scenario := range scenarios {
		mode := game.ModeAlone
		if len(scenario.SupportedModes) > 0 {
			mode = scenario.SupportedModes[0]
		}
		records = append(records, customScenarioRecord{
			Scenario:      scenario,
			PreferredMode: mode,
		})
	}
	payload := customScenarioLibrary{FormatVersion: 2, Custom: records}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func normalizeScenarioForMode(s *game.Scenario, mode game.GameMode) {
	if s == nil {
		return
	}
	if strings.TrimSpace(s.Name) == "" {
		s.Name = "Custom Scenario"
	}
	if strings.TrimSpace(s.Location) == "" {
		s.Location = "Wilderness"
	}
	if strings.TrimSpace(s.Biome) == "" {
		s.Biome = "temperate_forest"
	}
	if s.DefaultDays <= 0 {
		s.DefaultDays = defaultRunDaysForMode(mode)
	}
	if s.MapWidthCells <= 0 || s.MapHeightCells <= 0 {
		switch mode {
		case game.ModeAlone:
			s.MapWidthCells, s.MapHeightCells = 36, 36
		case game.ModeNakedAndAfraid:
			s.MapWidthCells, s.MapHeightCells = 100, 100
		case game.ModeNakedAndAfraidXL:
			s.MapWidthCells, s.MapHeightCells = 125, 125
		default:
			s.MapWidthCells, s.MapHeightCells = 72, 72
		}
	}
	s.MapWidthCells, s.MapHeightCells = clampScenarioMapSize(mode, s.MapWidthCells, s.MapHeightCells)
	if len(s.SupportedModes) == 0 {
		s.SupportedModes = []game.GameMode{mode}
	} else {
		s.SupportedModes = []game.GameMode{s.SupportedModes[0]}
	}
	if len(s.Wildlife) == 0 {
		s.Wildlife = game.WildlifeForBiome(s.Biome)
	}
	if len(s.SeasonSets) == 0 {
		set := defaultSeasonSetForMode(mode)
		s.SeasonSets = []game.SeasonSet{set}
		s.DefaultSeasonSetID = set.ID
		return
	}

	fallback := defaultSeasonSetForMode(mode)
	for i := range s.SeasonSets {
		set := &s.SeasonSets[i]
		if strings.TrimSpace(string(set.ID)) == "" {
			set.ID = game.SeasonSetID(fmt.Sprintf("custom_profile_%d", i+1))
		}
		if len(set.Phases) == 0 {
			set.Phases = append([]game.SeasonPhase(nil), fallback.Phases...)
		}
		for j := range set.Phases {
			if strings.TrimSpace(string(set.Phases[j].Season)) == "" {
				set.Phases[j].Season = game.SeasonAutumn
			}
			if set.Phases[j].Days < 0 {
				set.Phases[j].Days = 0
			}
		}
	}
	if strings.TrimSpace(string(s.DefaultSeasonSetID)) == "" {
		s.DefaultSeasonSetID = s.SeasonSets[0].ID
		return
	}
	for _, set := range s.SeasonSets {
		if set.ID == s.DefaultSeasonSetID {
			return
		}
	}
	s.DefaultSeasonSetID = s.SeasonSets[0].ID
}

func newScenarioTemplate(mode game.GameMode) game.Scenario {
	s := game.Scenario{
		Name:        "Custom Scenario",
		Location:    "Wilderness",
		Biome:       "temperate_forest",
		Description: "A custom survival scenario.",
		Daunting:    "Unknown risk profile until tested in run.",
		Motivation:  "Build your own challenge and beat it.",
		DefaultDays: defaultRunDaysForMode(mode),
	}
	normalizeScenarioForMode(&s, mode)
	s.ID = ""
	return s
}

func defaultRunDaysForMode(mode game.GameMode) int {
	switch mode {
	case game.ModeAlone:
		return 365
	case game.ModeNakedAndAfraid:
		return 21
	case game.ModeNakedAndAfraidXL:
		return 40
	default:
		return 30
	}
}

func defaultSeasonSetForMode(mode game.GameMode) game.SeasonSet {
	switch mode {
	case game.ModeAlone:
		return game.SeasonSet{
			ID: "custom_alone_default",
			Phases: []game.SeasonPhase{
				{Season: game.SeasonAutumn, Days: 14},
				{Season: game.SeasonWinter, Days: 0},
			},
		}
	case game.ModeNakedAndAfraid:
		return game.SeasonSet{
			ID: "custom_na_wet",
			Phases: []game.SeasonPhase{
				{Season: game.SeasonWet, Days: 0},
			},
		}
	case game.ModeNakedAndAfraidXL:
		return game.SeasonSet{
			ID: "custom_xl_dry",
			Phases: []game.SeasonPhase{
				{Season: game.SeasonDry, Days: 0},
			},
		}
	default:
		return game.SeasonSet{ID: "custom_default", Phases: []game.SeasonPhase{{Season: game.SeasonAutumn, Days: 0}}}
	}
}

func generateScenarioID(name string, mode game.GameMode, existing []game.Scenario) string {
	base := slugify(fmt.Sprintf("%s_%s", mode, name))
	if base == "" {
		base = "custom_scenario"
	}
	id := base
	used := map[string]bool{}
	for _, scenario := range existing {
		used[string(scenario.ID)] = true
	}
	counter := 2
	for used[id] {
		id = fmt.Sprintf("%s_%d", base, counter)
		counter++
	}
	return id
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
