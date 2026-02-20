package game

import (
	"fmt"
	"math/rand"
	"time"
)

type RunState struct {
	Config      RunConfig
	Scenario    Scenario
	SeasonSetID SeasonSetID
	Day         int
	ClockHours  float64
	Players     []PlayerState
	Contestants []ContestantState `json:"contestants,omitempty"`
	Weather     WeatherState

	MetabolismProgress  float64         `json:"metabolism_progress"`
	WoodStock           []WoodStock     `json:"wood_stock,omitempty"`
	ResourceStock       []ResourceStock `json:"resource_stock,omitempty"`
	CampInventory       []InventoryItem `json:"camp_inventory,omitempty"`
	Travel              TravelState     `json:"travel"`
	Fire                FireState       `json:"fire"`
	FirePrep            FirePrepState   `json:"fire_prep"`
	Shelter             ShelterState    `json:"shelter"`
	CraftedItems        []string        `json:"crafted_items,omitempty"`
	PlacedTraps         []PlacedTrap    `json:"placed_traps,omitempty"`
	FireAttemptCount    int             `json:"fire_attempt_count"`
	ProcessAttemptCount int             `json:"process_attempt_count"`
	Topology            WorldTopology   `json:"topology"`
	FogMask             []bool          `json:"fog_mask,omitempty"`
	CellStates          []CellState     `json:"cell_states,omitempty"`
}

func NewRunState(config RunConfig) (RunState, error) {
	resolvedConfig := config

	if err := resolvedConfig.Validate(); err != nil {
		return RunState{}, err
	}

	if resolvedConfig.Seed == 0 {
		resolvedConfig.Seed = time.Now().UnixNano()
	}

	scenarios := AllScenarios()

	if resolvedConfig.ScenarioID == ScenarioRandomID {
		rng := seededRNG(resolvedConfig.Seed)
		resolvedConfig.ScenarioID = scenarios[rng.IntN(len(scenarios))].ID
	}

	scenario, found := GetScenario(scenarios, resolvedConfig.ScenarioID)

	if !found {
		return RunState{}, fmt.Errorf("scenario not found: %s", resolvedConfig.ScenarioID)
	}

	state := RunState{
		Config:      resolvedConfig,
		Scenario:    scenario,
		SeasonSetID: scenario.DefaultSeasonSetID,
		Day:         1,
		ClockHours:  7,
		Players:     CreatePlayers(resolvedConfig),
	}

	if resolvedConfig.Mode == ModeAlone {
		numCompetitors := 10 - resolvedConfig.PlayerCount
		if numCompetitors > 0 {
			state.Contestants = initialContestants(numCompetitors, resolvedConfig.Seed)
		}
	}
	state.EnsureWeather()
	state.EnsurePlayerRuntimeStats()
	state.initTopology()

	return state, nil
}

func GetScenario(scenarios []Scenario, id ScenarioID) (Scenario, bool) {
	for _, scenario := range scenarios {
		if scenario.ID == id {
			return scenario, true
		}
	}

	return Scenario{}, false
}

func (s *RunState) ProcessContestantSimulation(delta time.Duration) []string {
	if s.Config.Mode != ModeAlone || len(s.Contestants) == 0 {
		return nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var messages []string
	for i := range s.Contestants {
		if hasEvent, msg := s.Contestants[i].processMacroTick(delta, s.Weather, r); hasEvent {
			messages = append(messages, msg)
		}
	}
	return messages
}
