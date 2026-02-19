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

	player := &s.Players[playerIndex]
	switch domain {
	case AnimalDomainLand:
		applySkillEffort(&player.Hunting, 18, true)
	case AnimalDomainWater:
		applySkillEffort(&player.Fishing, 18, true)
	default:
		applySkillEffort(&player.Hunting, 12, true)
	}

	// Skill and trait bonuses increase effective usable catch.
	bonusPct := 0
	switch domain {
	case AnimalDomainLand:
		bonusPct = player.Hunting/8 + player.Strength + player.Agility + positiveTraitModifier(player.Traits)/2
	case AnimalDomainWater:
		bonusPct = player.Fishing/8 + player.Agility + positiveTraitModifier(player.Traits)/2
	default:
		bonusPct = player.Hunting/10 + player.Agility + positiveTraitModifier(player.Traits)/2
	}
	if bonusPct != 0 {
		adjusted := catch.WeightGrams + (catch.WeightGrams*bonusPct)/100
		catch.WeightGrams = max(80, adjusted)
	}

	outcome := ConsumeCatch(s.Config.Seed, s.Day, player, catch, choice)
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
