package game

import "math"

type DailyNutritionNeeds struct {
	CaloriesKcal int `json:"calories_kcal"`
	ProteinG     int `json:"protein_g"`
	FatG         int `json:"fat_g"`
	SugarG       int `json:"sugar_g"`
}

func DailyNutritionNeedsForPlayer(player PlayerState) DailyNutritionNeeds {
	weight := player.WeightKg
	if weight <= 0 {
		weight = 75
	}

	calories := 1750 + (weight * 9)
	protein := maxInt(50, int(float64(weight)*0.8))
	fat := maxInt(35, int(float64(weight)*0.5))
	sugar := maxInt(90, int(float64(weight)*1.2))

	switch player.BodyType {
	case BodyTypeMale:
		calories += 140
		protein += 6
		fat += 4
		sugar += 10
	case BodyTypeFemale:
		calories -= 40
		protein -= 2
		fat -= 2
		sugar -= 5
	}

	if player.HeightFt >= 6 {
		calories += 60
		sugar += 6
	}

	calories += player.Endurance * 80
	if calories < 1500 {
		calories = 1500
	}
	if protein < 40 {
		protein = 40
	}
	if fat < 28 {
		fat = 28
	}
	if sugar < 70 {
		sugar = 70
	}

	return DailyNutritionNeeds{
		CaloriesKcal: calories,
		ProteinG:     protein,
		FatG:         fat,
		SugarG:       sugar,
	}
}

func (s *RunState) EnsurePlayerRuntimeStats() {
	if s == nil {
		return
	}
	for i := range s.Players {
		initializeRuntimeBars(&s.Players[i])
	}
}

func initializeRuntimeBars(player *PlayerState) {
	if player == nil {
		return
	}
	needs := DailyNutritionNeedsForPlayer(*player)
	likelyLegacyLoad := player.CaloriesReserveKcal == 0 &&
		player.ProteinReserveG == 0 &&
		player.FatReserveG == 0 &&
		player.SugarReserveG == 0 &&
		player.Hunger == 0 &&
		player.Thirst == 0 &&
		player.Fatigue == 0

	if likelyLegacyLoad {
		player.CaloriesReserveKcal = needs.CaloriesKcal * 2
		player.ProteinReserveG = needs.ProteinG * 2
		player.FatReserveG = needs.FatG * 2
		player.SugarReserveG = needs.SugarG * 2
	}
	if player.SugarReserveG == 0 && player.Hunger == 0 && player.Thirst == 0 && player.Fatigue == 0 {
		player.SugarReserveG = needs.SugarG * 2
	}

	player.CaloriesReserveKcal = clamp(player.CaloriesReserveKcal, -5000, 12000)
	player.ProteinReserveG = clamp(player.ProteinReserveG, -250, 800)
	player.FatReserveG = clamp(player.FatReserveG, -250, 600)
	player.SugarReserveG = clamp(player.SugarReserveG, -300, 800)

	refreshEffectBars(player)
}

func applyMealNutritionReserves(player *PlayerState, nutrition NutritionTotals) {
	if player == nil {
		return
	}
	initializeRuntimeBars(player)

	player.CaloriesReserveKcal = clamp(player.CaloriesReserveKcal+nutrition.CaloriesKcal, -5000, 12000)
	player.ProteinReserveG = clamp(player.ProteinReserveG+nutrition.ProteinG, -250, 800)
	player.FatReserveG = clamp(player.FatReserveG+nutrition.FatG, -250, 600)
	player.SugarReserveG = clamp(player.SugarReserveG+nutrition.SugarG, -300, 800)

	refreshEffectBars(player)
}

func applyDailyMetabolism(player *PlayerState) {
	applyMetabolismFraction(player, 1.0)
}

func applyMetabolismFraction(player *PlayerState, fraction float64) {
	if player == nil || fraction <= 0 {
		return
	}
	if fraction > 1 {
		fraction = 1
	}
	initializeRuntimeBars(player)

	needs := DailyNutritionNeedsForPlayer(*player)
	consumeReserveWithCarry(&player.CaloriesReserveKcal, &player.metabolismCarryCalories, float64(needs.CaloriesKcal)*fraction, -5000, 12000)
	consumeReserveWithCarry(&player.ProteinReserveG, &player.metabolismCarryProtein, float64(needs.ProteinG)*fraction, -250, 800)
	consumeReserveWithCarry(&player.FatReserveG, &player.metabolismCarryFat, float64(needs.FatG)*fraction, -250, 600)
	consumeReserveWithCarry(&player.SugarReserveG, &player.metabolismCarrySugar, float64(needs.SugarG)*fraction, -300, 800)

	refreshEffectBars(player)
	penalty := effectBarPenalty(*player)
	applyScaledDeltaWithCarry(&player.Energy, &player.metabolismCarryEnergy, float64(penalty.Energy)*fraction, 0, 100)
	applyScaledDeltaWithCarry(&player.Hydration, &player.metabolismCarryHydration, float64(penalty.Hydration)*fraction, 0, 100)
	applyScaledDeltaWithCarry(&player.Morale, &player.metabolismCarryMorale, float64(penalty.Morale)*fraction, 0, 100)

	refreshEffectBars(player)
}

func consumeReserveWithCarry(value *int, carry *float64, amount float64, min int, max int) {
	if value == nil || carry == nil || amount <= 0 {
		return
	}
	*carry += amount
	whole := extractWholeCarry(carry)
	if whole <= 0 {
		return
	}
	*value = clamp(*value-whole, min, max)
}

func applyScaledDeltaWithCarry(value *int, carry *float64, amount float64, min int, max int) {
	if value == nil || carry == nil || amount == 0 {
		return
	}
	*carry += amount
	whole := extractWholeCarry(carry)
	if whole == 0 {
		return
	}
	*value = clamp(*value+whole, min, max)
}

func extractWholeCarry(carry *float64) int {
	if carry == nil {
		return 0
	}
	if *carry >= 1 {
		whole := int(math.Floor(*carry))
		*carry -= float64(whole)
		return whole
	}
	if *carry <= -1 {
		whole := int(math.Ceil(*carry))
		*carry -= float64(whole)
		return whole
	}
	return 0
}

func refreshEffectBars(player *PlayerState) {
	if player == nil {
		return
	}
	needs := DailyNutritionNeedsForPlayer(*player)
	idealReserve := maxInt(needs.CaloriesKcal, 1)
	hunger := 0
	switch {
	case player.CaloriesReserveKcal >= idealReserve:
		surplus := player.CaloriesReserveKcal - idealReserve
		hunger = 25 - (surplus / 220)
	case player.CaloriesReserveKcal >= 0:
		missing := idealReserve - player.CaloriesReserveKcal
		hunger = 25 + (missing*35)/idealReserve
	default:
		deficit := -player.CaloriesReserveKcal
		hunger = 60 + (deficit*40)/idealReserve
	}
	if player.ProteinReserveG < 0 {
		hunger += (-player.ProteinReserveG) / 6
	}
	if player.FatReserveG < 0 {
		hunger += (-player.FatReserveG) / 8
	}
	if player.SugarReserveG < 0 {
		hunger += (-player.SugarReserveG) / 5
	}
	player.Hunger = clamp(hunger, 0, 100)

	thirst := clamp(100-player.Hydration, 0, 100)
	if player.Hydration < 25 {
		thirst = clamp(thirst+10, 0, 100)
	}
	player.Thirst = thirst

	fatigue := (100 - player.Energy) + player.Hunger/4 + player.Thirst/5 + len(player.Ailments)*4
	if player.Morale < 40 {
		fatigue += (40 - player.Morale) / 4
	}
	if player.SugarReserveG < 0 {
		fatigue += (-player.SugarReserveG) / 12
	}
	player.Fatigue = clamp(fatigue, 0, 100)
}

func effectBarPenalty(player PlayerState) statDelta {
	var out statDelta

	if player.Hunger > 35 {
		over := player.Hunger - 30
		out.Energy -= over / 14
		out.Morale -= over / 18
	}
	if player.Thirst > 30 {
		over := player.Thirst - 25
		out.Energy -= over / 12
		out.Morale -= over / 16
		if player.Thirst > 40 {
			out.Hydration -= (player.Thirst - 40) / 20
		}
	}
	if player.Fatigue > 40 {
		over := player.Fatigue - 35
		out.Energy -= over / 12
		out.Morale -= over / 18
	}

	if player.Hunger >= 90 {
		out.Morale -= 2
	}
	if player.Thirst >= 90 {
		out.Energy -= 2
		out.Morale -= 2
	}
	if player.Fatigue >= 90 {
		out.Energy -= 3
	}

	// Core player modifiers mitigate or amplify bar-derived penalties.
	if out.Energy < 0 {
		out.Energy += clamp(player.Endurance, -3, 3)
	}
	if out.Hydration < 0 {
		out.Hydration += clamp(player.Bushcraft, -3, 3)
	}
	if out.Morale < 0 {
		out.Morale += clamp(player.Mental, -3, 3)
	}

	return out
}
