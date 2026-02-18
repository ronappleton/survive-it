package game

import (
	"fmt"
	"time"
)

type RunState struct {
	Config      RunConfig
	Scenario    Scenario
	SeasonSetID SeasonSetID
	Day         int
	Players     []PlayerState
	Weather     WeatherState
}

func NewRunState(config RunConfig) (RunState, error) {
	resolvedConfig := config

	if err := resolvedConfig.Validate(); err != nil {
		return RunState{}, err
	}

	if resolvedConfig.Seed == 0 {
		resolvedConfig.Seed = time.Now().UnixNano()
	}

	scenarios := BuiltInScenarios()

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
		Players:     CreatePlayers(resolvedConfig),
	}
	state.EnsureWeather()
	state.EnsurePlayerRuntimeStats()

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
