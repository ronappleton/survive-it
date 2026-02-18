package game

import "fmt"

func (s *RunState) CatchAndConsume(playerID int, domain AnimalDomain, choice MealChoice) (CatchResult, MealOutcome, error) {
	if s == nil {
		return CatchResult{}, MealOutcome{}, fmt.Errorf("run state is nil")
	}
	s.EnsurePlayerRuntimeStats()
	playerIndex := -1
	for i := range s.Players {
		if s.Players[i].ID == playerID {
			playerIndex = i
			break
		}
	}
	if playerIndex < 0 {
		return CatchResult{}, MealOutcome{}, fmt.Errorf("player %d not found", playerID)
	}

	catch, err := RandomCatch(s.Config.Seed, s.Scenario.Biome, domain, s.Day, playerID)
	if err != nil {
		return CatchResult{}, MealOutcome{}, err
	}

	outcome := ConsumeCatch(s.Config.Seed, s.Day, &s.Players[playerIndex], catch, choice)
	return catch, outcome, nil
}

func (s *RunState) ActiveAilmentNames(playerID int) []string {
	for i := range s.Players {
		if s.Players[i].ID != playerID {
			continue
		}
		out := make([]string, 0, len(s.Players[i].Ailments))
		for _, ailment := range s.Players[i].Ailments {
			name := ailment.Name
			if name == "" {
				name = string(ailment.Type)
			}
			out = append(out, name)
		}
		return out
	}
	return nil
}
