package game

import "testing"

func TestAnimalsForBiomeByDomain(t *testing.T) {
	land := AnimalsForBiome("tropical_jungle", AnimalDomainLand)
	water := AnimalsForBiome("tropical_jungle", AnimalDomainWater)
	air := AnimalsForBiome("tropical_jungle", AnimalDomainAir)

	if len(land) == 0 || len(water) == 0 || len(air) == 0 {
		t.Fatalf("expected all domains to have options in jungle, got land=%d water=%d air=%d", len(land), len(water), len(air))
	}

	foundMouse := false
	for _, animal := range land {
		if animal.ID == "mouse" {
			foundMouse = true
			break
		}
	}
	if !foundMouse {
		t.Fatalf("expected mouse to be available as a land animal")
	}
}

func TestRandomCatchRespectsWeightAndYield(t *testing.T) {
	catch, err := RandomCatch(99, "temperate_rainforest", AnimalDomainWater, 3, 1)
	if err != nil {
		t.Fatalf("random catch: %v", err)
	}
	if catch.WeightGrams <= 0 {
		t.Fatalf("expected positive catch weight")
	}
	if catch.EdibleGrams <= 0 || catch.EdibleGrams > catch.WeightGrams {
		t.Fatalf("invalid edible grams: %d from weight %d", catch.EdibleGrams, catch.WeightGrams)
	}

	min := int(catch.Animal.WeightMinKg * 1000)
	max := int(catch.Animal.WeightMaxKg*1000) + 1
	if catch.WeightGrams < min || catch.WeightGrams > max {
		t.Fatalf("catch weight %d outside species range [%d,%d]", catch.WeightGrams, min, max)
	}
}

func TestConsumeCatchAppliesNutritionAndEffects(t *testing.T) {
	player := &PlayerState{ID: 1, Energy: 40, Hydration: 40, Morale: 40}
	catch := CatchResult{
		Animal: AnimalSpec{
			ID:               "test_fish",
			Name:             "Test Fish",
			EdibleYieldRatio: 0.5,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 200, ProteinG: 20, FatG: 8},
		},
		WeightGrams: 1000,
		EdibleGrams: 500,
	}

	outcome := ConsumeCatch(42, 1, player, catch, MealChoice{PortionGrams: 300, Cooked: true})
	if outcome.Nutrition.CaloriesKcal <= 0 {
		t.Fatalf("expected positive calories outcome")
	}
	if player.Energy <= 40 || player.Morale <= 40 {
		t.Fatalf("expected player stats to improve after meal, got E=%d H=%d M=%d", player.Energy, player.Hydration, player.Morale)
	}
	if player.Nutrition.CaloriesKcal != outcome.Nutrition.CaloriesKcal {
		t.Fatalf("expected nutrition totals to be tracked on player")
	}
}

func TestConsumeCatchCanTriggerLiverDisease(t *testing.T) {
	player := &PlayerState{ID: 1, Energy: 90, Hydration: 90, Morale: 90}
	catch := CatchResult{
		Animal: AnimalSpec{
			ID:               "test_rabbit",
			Name:             "Test Rabbit",
			EdibleYieldRatio: 0.5,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 170, ProteinG: 30, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{
					ID:          "test_liver",
					Name:        "Liver Worms",
					BaseChance:  1,
					CarrierPart: "liver",
					Effect: AilmentTemplate{
						Type:             AilmentVomiting,
						Name:             "Vomiting",
						Days:             2,
						EnergyPenalty:    3,
						HydrationPenalty: 5,
						MoralePenalty:    4,
					},
				},
			},
		},
		WeightGrams: 2000,
		EdibleGrams: 1000,
	}

	outcome := ConsumeCatch(100, 2, player, catch, MealChoice{PortionGrams: 200, Cooked: false, EatLiver: true})
	if len(outcome.DiseaseEvents) == 0 {
		t.Fatalf("expected guaranteed liver disease event to trigger")
	}
	if len(player.Ailments) == 0 {
		t.Fatalf("expected ailment added to player")
	}
}

func TestAdvanceDayAppliesAilmentPenalties(t *testing.T) {
	run := RunState{
		Config: RunConfig{
			Mode:        ModeAlone,
			ScenarioID:  ScenarioVancouverIslandID,
			PlayerCount: 1,
			RunLength:   RunLength{Days: 7},
			Seed:        1000,
		},
		Scenario: BuiltInScenarios()[0],
		Day:      1,
		Players: []PlayerState{
			{
				ID:        1,
				Name:      "Test",
				BodyType:  BodyTypeNeutral,
				Energy:    100,
				Hydration: 100,
				Morale:    100,
				Ailments: []Ailment{
					{Type: AilmentVomiting, Name: "Vomiting", DaysRemaining: 2, EnergyPenalty: 1, HydrationPenalty: 2, MoralePenalty: 1},
				},
			},
		},
	}
	run.SeasonSetID = run.Scenario.DefaultSeasonSetID
	run.EnsureWeather()

	run.AdvanceDay()
	if len(run.Players[0].Ailments) != 1 {
		t.Fatalf("expected ailment to still be active after first day")
	}

	run.AdvanceDay()
	if len(run.Players[0].Ailments) != 0 {
		t.Fatalf("expected ailment to expire after second day")
	}
}

func TestRunCatchAndConsumeFlow(t *testing.T) {
	cfg := RunConfig{
		Mode:        ModeNakedAndAfraid,
		ScenarioID:  ScenarioJungleID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 21},
		Seed:        222,
	}
	run, err := NewRunState(cfg)
	if err != nil {
		t.Fatalf("new run: %v", err)
	}

	catch, outcome, err := run.CatchAndConsume(1, AnimalDomainLand, MealChoice{Cooked: true, PortionGrams: 150})
	if err != nil {
		t.Fatalf("catch and consume: %v", err)
	}
	if catch.WeightGrams <= 0 || outcome.Nutrition.CaloriesKcal <= 0 {
		t.Fatalf("expected positive catch and nutrition values")
	}
}
