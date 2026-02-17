package game

import (
	"fmt"
)

type GameMode string

const (
	ModeNakedAndAfraid   GameMode = "naked_and_afraid"
	ModeNakedAndAfraidXL GameMode = "naked_and_afraid_xl"
	ModeAlone            GameMode = "alone"
)

type RunConfig struct {
	Mode        GameMode
	ScenarioID  ScenarioID
	PlayerCount int
	Players     []PlayerConfig
	RunLength   RunLength
	Seed        int64
}

type RunLength struct {
	OpenEnded bool
	Days      int
}

func (r RunLength) IsValid() bool {
	return r.OpenEnded || r.Days > 0
}

func (c RunConfig) Validate() error {
	switch c.Mode {
	case ModeNakedAndAfraid:
	case ModeNakedAndAfraidXL:
	case ModeAlone:
	default:
		return fmt.Errorf("invalid mode: %s", c.Mode)
	}

	if c.PlayerCount < 1 || c.PlayerCount > 8 {
		return fmt.Errorf("player count must be between 1 and 8, got %d", c.PlayerCount)
	}

	found := c.ScenarioID == ScenarioRandomID

	if !found {
		for _, scenario := range BuiltInScenarios() {
			if scenario.ID == c.ScenarioID {
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("scenario not found: %s", c.ScenarioID)
	}

	if !c.RunLength.IsValid() {
		return fmt.Errorf("invalid run length")
	}

	return nil
}
