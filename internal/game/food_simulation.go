package game

import (
	"fmt"
	"hash/fnv"
)

type MealChoice struct {
	PortionGrams int
	Cooked       bool
	EatLiver     bool
}

type DiseaseEvent struct {
	DiseaseID DiseaseID
	Name      string
	Ailment   Ailment
}

type MealOutcome struct {
	PlayerID       int
	AnimalName     string
	PortionGrams   int
	Nutrition      NutritionTotals
	EnergyDelta    int
	HydrationDelta int
	MoraleDelta    int
	DiseaseEvents  []DiseaseEvent
}

func ConsumeCatch(seed int64, day int, player *PlayerState, catch CatchResult, choice MealChoice) MealOutcome {
	if player == nil {
		return MealOutcome{}
	}
	if day < 1 {
		day = 1
	}

	portion := choice.PortionGrams
	if portion <= 0 {
		portion = clamp(catch.EdibleGrams/2, 80, catch.EdibleGrams)
	}
	if portion > catch.EdibleGrams {
		portion = catch.EdibleGrams
	}
	if portion < 1 {
		portion = 1
	}

	nutrition := catch.NutritionForGrams(portion)
	energyGain, hydrationGain, moraleGain := nutritionToPlayerEffects(nutrition)

	player.Energy = clamp(player.Energy+energyGain, 0, 100)
	player.Hydration = clamp(player.Hydration+hydrationGain, 0, 100)
	player.Morale = clamp(player.Morale+moraleGain, 0, 100)
	player.Nutrition = player.Nutrition.add(nutrition)

	outcome := MealOutcome{
		PlayerID:       player.ID,
		AnimalName:     catch.Animal.Name,
		PortionGrams:   portion,
		Nutrition:      nutrition,
		EnergyDelta:    energyGain,
		HydrationDelta: hydrationGain,
		MoraleDelta:    moraleGain,
	}

	for _, risk := range catch.Animal.DiseaseRisks {
		chance := adjustedDiseaseChance(risk, choice)
		if chance <= 0 {
			continue
		}
		if deterministicRiskRoll(seed, catch.Animal.ID, risk.ID, day, player.ID) > chance {
			continue
		}

		ailment := Ailment{
			Type:             risk.Effect.Type,
			Name:             risk.Effect.Name,
			DaysRemaining:    risk.Effect.Days,
			EnergyPenalty:    risk.Effect.EnergyPenalty,
			HydrationPenalty: risk.Effect.HydrationPenalty,
			MoralePenalty:    risk.Effect.MoralePenalty,
		}

		player.applyAilment(ailment)
		// Initial hit on the day of ingestion.
		player.Energy = clamp(player.Energy-ailment.EnergyPenalty, 0, 100)
		player.Hydration = clamp(player.Hydration-ailment.HydrationPenalty, 0, 100)
		player.Morale = clamp(player.Morale-ailment.MoralePenalty, 0, 100)

		outcome.DiseaseEvents = append(outcome.DiseaseEvents, DiseaseEvent{
			DiseaseID: risk.ID,
			Name:      risk.Name,
			Ailment:   ailment,
		})

		// Vomiting can occur as an additional symptom for many diseases.
		if ailment.Type != AilmentVomiting {
			vomitChance := adjustedVomitChance(risk, choice)
			if vomitChance > 0 {
				vomitID := DiseaseID(fmt.Sprintf("%s_vomit", risk.ID))
				if deterministicRiskRoll(seed, catch.Animal.ID, vomitID, day, player.ID) <= vomitChance {
					vomitAilment := Ailment{
						Type:             AilmentVomiting,
						Name:             "Vomiting",
						DaysRemaining:    1,
						EnergyPenalty:    2,
						HydrationPenalty: 4,
						MoralePenalty:    3,
					}
					player.applyAilment(vomitAilment)
					player.Energy = clamp(player.Energy-vomitAilment.EnergyPenalty, 0, 100)
					player.Hydration = clamp(player.Hydration-vomitAilment.HydrationPenalty, 0, 100)
					player.Morale = clamp(player.Morale-vomitAilment.MoralePenalty, 0, 100)
					outcome.DiseaseEvents = append(outcome.DiseaseEvents, DiseaseEvent{
						DiseaseID: vomitID,
						Name:      risk.Name + " (vomiting symptom)",
						Ailment:   vomitAilment,
					})
				}
			}
		}
	}

	return outcome
}

func (n NutritionTotals) add(other NutritionTotals) NutritionTotals {
	return NutritionTotals{
		CaloriesKcal: n.CaloriesKcal + other.CaloriesKcal,
		ProteinG:     n.ProteinG + other.ProteinG,
		FatG:         n.FatG + other.FatG,
	}
}

func nutritionToPlayerEffects(n NutritionTotals) (energy, hydration, morale int) {
	energy = clamp(n.CaloriesKcal/100+n.ProteinG/40+n.FatG/35, 0, 20)
	hydration = clamp(n.ProteinG/90+n.FatG/120, 0, 4)
	morale = clamp(n.CaloriesKcal/180+n.FatG/70, 0, 8)
	return energy, hydration, morale
}

func adjustedDiseaseChance(risk DiseaseRisk, choice MealChoice) float64 {
	chance := risk.BaseChance
	if chance <= 0 {
		return 0
	}

	switch risk.CarrierPart {
	case "liver":
		if !choice.EatLiver {
			return 0
		}
		chance *= 1.6
	case "blood":
		chance *= 1.25
	case "respiratory":
		chance *= 1.10
	}

	if choice.Cooked {
		switch risk.CarrierPart {
		case "respiratory":
			chance *= 0.85
		default:
			chance *= 0.45
		}
	} else {
		chance *= 1.75
	}

	if chance > 0.95 {
		return 0.95
	}
	return chance
}

func adjustedVomitChance(risk DiseaseRisk, choice MealChoice) float64 {
	chance := risk.VomitChance
	if chance <= 0 {
		chance = inferredVomitChance(risk)
	}
	if chance <= 0 {
		return 0
	}

	switch risk.CarrierPart {
	case "liver":
		if !choice.EatLiver {
			return 0
		}
		chance *= 1.25
	case "blood":
		chance *= 1.15
	case "respiratory":
		chance *= 0.60
	}

	if choice.Cooked {
		chance *= 0.80
	} else {
		chance *= 1.30
	}

	if chance > 0.95 {
		return 0.95
	}
	return chance
}

func inferredVomitChance(risk DiseaseRisk) float64 {
	switch risk.Effect.Type {
	case AilmentVomiting:
		return 0.90
	case AilmentFoodPoison:
		return 0.75
	case AilmentGIInfection:
		return 0.55
	case AilmentParasites:
		return 0.35
	case AilmentEnvenomation:
		return 0.35
	default:
		// Keep some chance for unspecified GI-like reactions.
		return clampFloat(risk.BaseChance*3.0, 0.05, 0.45)
	}
}

func deterministicRiskRoll(seed int64, animalID string, diseaseID DiseaseID, day, playerID int) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s:%s:%d:%d", seed, animalID, diseaseID, day, playerID)))
	return float64(h.Sum64()%10000) / 10000.0
}

func (p *PlayerState) applyAilment(next Ailment) {
	if next.DaysRemaining <= 0 {
		next.DaysRemaining = 1
	}
	for i := range p.Ailments {
		if p.Ailments[i].Type != next.Type {
			continue
		}
		if next.DaysRemaining > p.Ailments[i].DaysRemaining {
			p.Ailments[i].DaysRemaining = next.DaysRemaining
		}
		p.Ailments[i].EnergyPenalty = maxInt(p.Ailments[i].EnergyPenalty, next.EnergyPenalty)
		p.Ailments[i].HydrationPenalty = maxInt(p.Ailments[i].HydrationPenalty, next.HydrationPenalty)
		p.Ailments[i].MoralePenalty = maxInt(p.Ailments[i].MoralePenalty, next.MoralePenalty)
		if p.Ailments[i].Name == "" {
			p.Ailments[i].Name = next.Name
		}
		return
	}
	p.Ailments = append(p.Ailments, next)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampFloat(number, min, max float64) float64 {
	if number < min {
		return min
	}
	if number > max {
		return max
	}
	return number
}
