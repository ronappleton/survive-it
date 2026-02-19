package game

import (
	"fmt"
	"math"
)

type HuntResult struct {
	PlayerID    int
	Domain      AnimalDomain
	AnimalID    string
	AnimalName  string
	WeightGrams int
	CarcassID   string
	CarcassKg   float64
	StoredAt    string
	HoursSpent  float64
}

func carcassIDForDomain(domain AnimalDomain) string {
	switch domain {
	case AnimalDomainWater:
		return "fish_carcass"
	case AnimalDomainAir:
		return "bird_carcass"
	default:
		return "small_game_carcass"
	}
}

func (s *RunState) catchWithSkillBonus(playerID int, domain AnimalDomain) (CatchResult, *PlayerState, error) {
	if s == nil {
		return CatchResult{}, nil, fmt.Errorf("run state is nil")
	}
	s.EnsurePlayerRuntimeStats()
	player, ok := s.playerByID(playerID)
	if !ok {
		return CatchResult{}, nil, fmt.Errorf("player %d not found", playerID)
	}
	catch, err := RandomCatch(s.Config.Seed, s.Scenario.Biome, domain, s.Day, playerID)
	if err != nil {
		return CatchResult{}, nil, err
	}

	switch domain {
	case AnimalDomainLand:
		applySkillEffort(&player.Hunting, 18, true)
	case AnimalDomainWater:
		applySkillEffort(&player.Fishing, 18, true)
	default:
		applySkillEffort(&player.Hunting, 12, true)
	}

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
	catch.EdibleGrams = max(1, int(math.Round(float64(catch.WeightGrams)*catch.Animal.EdibleYieldRatio)))
	return catch, player, nil
}

func (s *RunState) HuntAndCollectCarcass(playerID int, domain AnimalDomain) (HuntResult, error) {
	catch, player, err := s.catchWithSkillBonus(playerID, domain)
	if err != nil {
		return HuntResult{}, err
	}

	carcassID := carcassIDForDomain(domain)
	carcass, ok := carcassCatalog[carcassID]
	if !ok {
		return HuntResult{}, fmt.Errorf("no carcass profile for domain %s", domain)
	}
	kg := math.Round(max(0.1, float64(catch.WeightGrams)/1000.0)*10) / 10
	item := InventoryItem{
		ID:       carcassID,
		Name:     carcass.Name,
		Unit:     "kg",
		Qty:      kg,
		WeightKg: 1.2,
		Category: "carcass",
		AgeDays:  0,
	}

	storedAt := ""
	if err := s.AddPersonalInventoryItem(playerID, item); err == nil {
		storedAt = "personal"
	} else if err := s.addCampInventoryItem(item); err == nil {
		storedAt = "camp"
	} else {
		return HuntResult{}, fmt.Errorf("caught %s (%.1fkg), but no storage space", catch.Animal.Name, kg)
	}

	baseHours := 1.8
	switch domain {
	case AnimalDomainWater:
		baseHours = 1.6
	case AnimalDomainAir:
		baseHours = 1.4
	}
	skillFactor := float64(player.Hunting) / 40.0
	if domain == AnimalDomainWater {
		skillFactor = float64(player.Fishing) / 40.0
	}
	hours := clampFloat(baseHours+(kg*0.05)-(skillFactor*0.25), 0.5, 10)
	_ = s.AdvanceActionClock(hours)

	player.Energy = clamp(player.Energy-int(math.Ceil(hours*2.0)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*1.4)), 0, 100)
	player.Morale = clamp(player.Morale+1, 0, 100)
	refreshEffectBars(player)

	return HuntResult{
		PlayerID:    playerID,
		Domain:      domain,
		AnimalID:    catch.Animal.ID,
		AnimalName:  catch.Animal.Name,
		WeightGrams: catch.WeightGrams,
		CarcassID:   carcassID,
		CarcassKg:   kg,
		StoredAt:    storedAt,
		HoursSpent:  hours,
	}, nil
}

func (s *RunState) CatchAndConsume(playerID int, domain AnimalDomain, choice MealChoice) (CatchResult, MealOutcome, error) {
	catch, player, err := s.catchWithSkillBonus(playerID, domain)
	if err != nil {
		return CatchResult{}, MealOutcome{}, err
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
