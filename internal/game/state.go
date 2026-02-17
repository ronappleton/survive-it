package game

import (
	"fmt"
	"go/doc"
	"math/rand/v2"
)

type PlayerState struct {
	Energy    int
	Hydration int
	Morale    int
}

type RunState struct {
	Config      RunConfig
	Scenario    Scenario
	SeasonSetID SeasonSetID
	Day         int
	Players     []PlayerState
}

func NewRunState(config RunConfig) (RunState, error) {
	var valid = config.Validate()
	if valid != nil {
		return RunState{}, valid
	}

	if config.ScenarioID == ScenarioRandomID {
		config.ScenarioID = BuiltInScenarios()[rand.Int64()%int64(len(BuiltInScenarios()))].ID
	}

	config.Scenario =

	return RunState{Config: config}
}
