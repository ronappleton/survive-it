package game

import "testing"

func TestCreatePlayersInitializesRuntimeMetabolism(t *testing.T) {
	players := CreatePlayers(RunConfig{
		PlayerCount: 1,
		Seed:        99,
	})
	if len(players) != 1 {
		t.Fatalf("expected one player")
	}

	p := players[0]
	if p.CaloriesReserveKcal <= 0 || p.ProteinReserveG <= 0 || p.FatReserveG <= 0 {
		t.Fatalf("expected positive initialized reserves, got kcal=%d protein=%d fat=%d", p.CaloriesReserveKcal, p.ProteinReserveG, p.FatReserveG)
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
