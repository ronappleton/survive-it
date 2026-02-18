package game

import (
	"testing"
	"time"
)

func TestCreatePlayersInitializesRuntimeMetabolism(t *testing.T) {
	players := CreatePlayers(RunConfig{
		PlayerCount: 1,
		Seed:        99,
	})
	if len(players) != 1 {
		t.Fatalf("expected one player")
	}

	p := players[0]
	if p.CaloriesReserveKcal <= 0 || p.ProteinReserveG <= 0 || p.FatReserveG <= 0 || p.SugarReserveG <= 0 {
		t.Fatalf("expected positive initialized reserves, got kcal=%d protein=%d fat=%d sugar=%d", p.CaloriesReserveKcal, p.ProteinReserveG, p.FatReserveG, p.SugarReserveG)
	}
	if p.Hunger < 0 || p.Hunger > 100 || p.Thirst < 0 || p.Thirst > 100 || p.Fatigue < 0 || p.Fatigue > 100 {
		t.Fatalf("expected initialized effect bars in range 0..100, got hunger=%d thirst=%d fatigue=%d", p.Hunger, p.Thirst, p.Fatigue)
	}
}

func TestConsumeCatchImprovesReservesAndHunger(t *testing.T) {
	player := PlayerState{
		ID:        1,
		BodyType:  BodyTypeNeutral,
		WeightKg:  75,
		Energy:    60,
		Hydration: 80,
		Morale:    70,
	}
	initializeRuntimeBars(&player)
	player.CaloriesReserveKcal = -500
	player.ProteinReserveG = -40
	player.FatReserveG = -20
	refreshEffectBars(&player)
	beforeHunger := player.Hunger

	catch := CatchResult{
		Animal: AnimalSpec{
			ID:               "test_fish",
			Name:             "Test Fish",
			EdibleYieldRatio: 0.5,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 240, ProteinG: 25, FatG: 10},
		},
		WeightGrams: 1000,
		EdibleGrams: 500,
	}

	_ = ConsumeCatch(1, 1, &player, catch, MealChoice{PortionGrams: 300, Cooked: true})
	if player.CaloriesReserveKcal <= -500 {
		t.Fatalf("expected calorie reserve to improve after meal")
	}
	if player.Hunger >= beforeHunger {
		t.Fatalf("expected hunger to reduce after meal, before=%d after=%d", beforeHunger, player.Hunger)
	}
}

func TestAdvanceDayAppliesMetabolismPenalty(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 10},
		Seed:        100,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	player := &run.Players[0]
	player.Energy = 70
	player.Hydration = 60
	player.Morale = 70
	player.CaloriesReserveKcal = -600
	player.ProteinReserveG = -40
	player.FatReserveG = -20
	refreshEffectBars(player)

	beforeEnergy := player.Energy
	beforeHunger := player.Hunger

	run.AdvanceDay()

	if player.Energy >= beforeEnergy {
		t.Fatalf("expected energy to drop from metabolism/effects, before=%d after=%d", beforeEnergy, player.Energy)
	}
	if player.Hunger <= beforeHunger {
		t.Fatalf("expected hunger to increase under deficit, before=%d after=%d", beforeHunger, player.Hunger)
	}
	if player.Thirst < 0 || player.Thirst > 100 || player.Fatigue < 0 || player.Fatigue > 100 {
		t.Fatalf("expected effect bars to stay bounded, thirst=%d fatigue=%d", player.Thirst, player.Fatigue)
	}
}

func TestEffectBarPenaltyUsesPlayerModifiers(t *testing.T) {
	base := PlayerState{
		Hunger:    85,
		Thirst:    85,
		Fatigue:   85,
		Energy:    60,
		Hydration: 60,
		Morale:    60,
	}

	strong := base
	strong.Endurance = 3
	strong.Bushcraft = 3
	strong.Mental = 3

	weak := base
	weak.Endurance = -3
	weak.Bushcraft = -3
	weak.Mental = -3

	strongPenalty := effectBarPenalty(strong)
	weakPenalty := effectBarPenalty(weak)

	if strongPenalty.Energy <= weakPenalty.Energy {
		t.Fatalf("expected endurance to improve energy penalty, strong=%d weak=%d", strongPenalty.Energy, weakPenalty.Energy)
	}
	if strongPenalty.Hydration <= weakPenalty.Hydration {
		t.Fatalf("expected bushcraft to improve hydration penalty, strong=%d weak=%d", strongPenalty.Hydration, weakPenalty.Hydration)
	}
	if strongPenalty.Morale <= weakPenalty.Morale {
		t.Fatalf("expected mental to improve morale penalty, strong=%d weak=%d", strongPenalty.Morale, weakPenalty.Morale)
	}
}

func TestApplyRealtimeMetabolismPartialDay(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 5},
		Seed:        212,
	})
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	player := &run.Players[0]
	beforeCalories := player.CaloriesReserveKcal
	beforeHydration := player.Hydration
	beforeProgress := run.MetabolismProgress

	run.ApplyRealtimeMetabolism(30*time.Minute, 2*time.Hour)

	if run.Day != 1 {
		t.Fatalf("expected day to remain 1, got %d", run.Day)
	}
	if run.MetabolismProgress <= beforeProgress || run.MetabolismProgress >= 1 {
		t.Fatalf("expected partial progress between 0 and 1, got %.2f", run.MetabolismProgress)
	}
	if player.CaloriesReserveKcal >= beforeCalories {
		t.Fatalf("expected calorie reserve to decrease during realtime metabolism, before=%d after=%d", beforeCalories, player.CaloriesReserveKcal)
	}
	if player.Hydration >= beforeHydration {
		t.Fatalf("expected hydration to decrease during realtime progression, before=%d after=%d", beforeHydration, player.Hydration)
	}
}

func TestDailyDeficiencyEffectsApplyMalnutritionAndDehydration(t *testing.T) {
	player := PlayerState{
		ID:                  1,
		BodyType:            BodyTypeNeutral,
		WeightKg:            75,
		Energy:              70,
		Hydration:           25,
		Morale:              70,
		CaloriesReserveKcal: -1800,
		ProteinReserveG:     -120,
		FatReserveG:         -90,
		SugarReserveG:       -80,
	}
	refreshEffectBars(&player)
	player.NutritionDeficitDays = 4
	player.DehydrationDays = 2

	applyDailyDeficiencyEffects(&player)

	if player.NutritionDeficitDays < 5 {
		t.Fatalf("expected nutrition deficit streak to increase, got %d", player.NutritionDeficitDays)
	}
	if player.DehydrationDays < 3 {
		t.Fatalf("expected dehydration streak to increase, got %d", player.DehydrationDays)
	}
	if player.Energy >= 70 {
		t.Fatalf("expected deficiency effects to reduce energy, got %d", player.Energy)
	}

	foundMalnutrition := false
	foundDehydration := false
	for _, ailment := range player.Ailments {
		if ailment.Type == AilmentMalnutrition {
			foundMalnutrition = true
		}
		if ailment.Type == AilmentDehydration {
			foundDehydration = true
		}
	}
	if !foundMalnutrition {
		t.Fatalf("expected malnutrition ailment under sustained deficit")
	}
	if !foundDehydration {
		t.Fatalf("expected dehydration ailment under sustained low hydration")
	}
}
