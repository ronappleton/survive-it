package game

func applyPhysiologyFraction(player *PlayerState, fraction float64) {
	if player == nil || fraction <= 0 {
		return
	}
	if fraction > 1 {
		fraction = 1
	}

	profile := PhysiologyFor(player.BodyType)
	applyScaledDeltaWithCarry(&player.Energy, &player.physiologyCarryEnergy, -float64(profile.EnergyDrainPerDay)*fraction, 0, 100)
	applyScaledDeltaWithCarry(&player.Hydration, &player.physiologyCarryHydration, -float64(profile.HydrationDrainPerDay)*fraction, 0, 100)
	applyScaledDeltaWithCarry(&player.Morale, &player.physiologyCarryMorale, -float64(profile.MoraleDrainPerDay)*fraction, 0, 100)
	refreshEffectBars(player)
}

func applyDailyDeficiencyEffects(player *PlayerState) {
	if player == nil {
		return
	}
	initializeRuntimeBars(player)

	deficitScore := nutritionDeficitScore(*player)
	if deficitScore > 0 {
		player.NutritionDeficitDays++
	} else if player.NutritionDeficitDays > 0 {
		player.NutritionDeficitDays--
	}

	if player.NutritionDeficitDays >= 2 && deficitScore > 0 {
		player.Energy = clamp(player.Energy-(1+deficitScore/2), 0, 100)
		player.Morale = clamp(player.Morale-(1+deficitScore/3), 0, 100)
	}
	if player.NutritionDeficitDays >= 4 && deficitScore >= 3 {
		player.applyAilment(Ailment{
			Type:             AilmentMalnutrition,
			Name:             "Malnutrition",
			DaysRemaining:    2,
			EnergyPenalty:    clamp(1+deficitScore/2, 1, 6),
			HydrationPenalty: clamp(deficitScore/4, 0, 2),
			MoralePenalty:    clamp(1+deficitScore/3, 1, 5),
		})
	}

	dehydrationScore := dehydrationSeverity(*player)
	if dehydrationScore > 0 {
		player.DehydrationDays++
	} else if player.DehydrationDays > 0 {
		player.DehydrationDays--
	}

	if dehydrationScore > 0 {
		player.Energy = clamp(player.Energy-(dehydrationScore), 0, 100)
		player.Morale = clamp(player.Morale-(dehydrationScore/2), 0, 100)
	}
	if dehydrationScore >= 2 {
		player.Hydration = clamp(player.Hydration-1, 0, 100)
	}
	if player.DehydrationDays >= 2 && dehydrationScore >= 2 {
		player.applyAilment(Ailment{
			Type:             AilmentDehydration,
			Name:             "Dehydration",
			DaysRemaining:    2,
			EnergyPenalty:    clamp(dehydrationScore, 2, 6),
			HydrationPenalty: clamp(1+dehydrationScore, 2, 7),
			MoralePenalty:    clamp(dehydrationScore/2, 1, 4),
		})
	}

	refreshEffectBars(player)
}

func nutritionDeficitScore(player PlayerState) int {
	score := 0
	if player.CaloriesReserveKcal < 0 {
		score += clamp((-player.CaloriesReserveKcal)/400, 1, 6)
	}
	if player.ProteinReserveG < 0 {
		score += clamp((-player.ProteinReserveG)/18, 1, 4)
	}
	if player.FatReserveG < 0 {
		score += clamp((-player.FatReserveG)/16, 1, 4)
	}
	if player.SugarReserveG < 0 {
		score += clamp((-player.SugarReserveG)/24, 1, 3)
	}
	if player.Hunger >= 70 {
		score += clamp((player.Hunger-60)/10, 1, 4)
	}
	return score
}

func dehydrationSeverity(player PlayerState) int {
	score := 0
	if player.Hydration <= 65 {
		score += clamp((70-player.Hydration)/15, 1, 4)
	}
	if player.Thirst >= 65 {
		score += clamp((player.Thirst-60)/12, 1, 4)
	}
	if player.Hydration <= 35 {
		score++
	}
	return score
}
