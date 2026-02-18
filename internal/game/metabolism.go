package game

type DailyNutritionNeeds struct {
	CaloriesKcal int `json:"calories_kcal"`
	ProteinG     int `json:"protein_g"`
	FatG         int `json:"fat_g"`
}

func DailyNutritionNeedsForPlayer(player PlayerState) DailyNutritionNeeds {
	weight := player.WeightKg
	if weight <= 0 {
		weight = 75
	}

	calories := 1750 + (weight * 9)
	protein := maxInt(50, int(float64(weight)*0.8))
	fat := maxInt(35, int(float64(weight)*0.5))

	switch player.BodyType {
	case BodyTypeMale:
		calories += 140
		protein += 6
		fat += 4
	case BodyTypeFemale:
		calories -= 40
		protein -= 2
		fat -= 2
	}

	if player.HeightFt >= 6 {
		calories += 60
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

	return DailyNutritionNeeds{
		CaloriesKcal: calories,
		ProteinG:     protein,
		FatG:         fat,
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
		player.Hunger == 0 &&
		player.Thirst == 0 &&
		player.Fatigue == 0

	if likelyLegacyLoad {
		player.CaloriesReserveKcal = needs.CaloriesKcal * 2
		player.ProteinReserveG = needs.ProteinG * 2
		player.FatReserveG = needs.FatG * 2
	}

	player.CaloriesReserveKcal = clamp(player.CaloriesReserveKcal, -5000, 12000)
	player.ProteinReserveG = clamp(player.ProteinReserveG, -250, 800)
	player.FatReserveG = clamp(player.FatReserveG, -250, 600)

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

	refreshEffectBars(player)
}

func applyDailyMetabolism(player *PlayerState) {
	if player == nil {
		return
	}
	initializeRuntimeBars(player)

	needs := DailyNutritionNeedsForPlayer(*player)
	player.CaloriesReserveKcal -= needs.CaloriesKcal
	player.ProteinReserveG -= needs.ProteinG
	player.FatReserveG -= needs.FatG

	player.CaloriesReserveKcal = clamp(player.CaloriesReserveKcal, -5000, 12000)
	player.ProteinReserveG = clamp(player.ProteinReserveG, -250, 800)
	player.FatReserveG = clamp(player.FatReserveG, -250, 600)

	refreshEffectBars(player)
	penalty := effectBarPenalty(*player)
	player.Energy += penalty.Energy
	player.Hydration += penalty.Hydration
	player.Morale += penalty.Morale
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

	return out
}
