package game

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type RunState struct {
	Config      RunConfig
	Scenario    Scenario
	SeasonSetID SeasonSetID
	Day         int
	Players     []PlayerState
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
		rng := rand.New(rand.NewPCG(uint64(resolvedConfig.Seed), uint64(resolvedConfig.Seed^0x9e3779b97f4a7c15)))
		resolvedConfig.ScenarioID = scenarios[rng.IntN(len(scenarios))].ID
	}

	scenario, found := GetScenario(scenarios, resolvedConfig.ScenarioID)

	if !found {
		return RunState{}, fmt.Errorf("scenario not found: %s", resolvedConfig.ScenarioID)
	}

	return RunState{
		Config:      resolvedConfig,
		Scenario:    scenario,
		SeasonSetID: scenario.DefaultSeasonSetID,
		Day:         1,
		Players:     CreatePlayers(resolvedConfig),
	}, nil
}

func GetScenario(scenarios []Scenario, id ScenarioID) (Scenario, bool) {
	for _, scenario := range scenarios {
		if scenario.ID == id {
			return scenario, true
		}
	}

	return Scenario{}, false
}
