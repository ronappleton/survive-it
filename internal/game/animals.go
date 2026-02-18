package game

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"
)

type AnimalDomain string

const (
	AnimalDomainLand  AnimalDomain = "land"
	AnimalDomainWater AnimalDomain = "water"
	AnimalDomainAir   AnimalDomain = "air"
)

type NutritionPer100g struct {
	CaloriesKcal int
	ProteinG     int
	FatG         int
}

type NutritionTotals struct {
	CaloriesKcal int `json:"calories_kcal"`
	ProteinG     int `json:"protein_g"`
	FatG         int `json:"fat_g"`
}

type DiseaseID string

type AilmentType string

const (
	AilmentVomiting      AilmentType = "vomiting"
	AilmentParasites     AilmentType = "parasites"
	AilmentFoodPoison    AilmentType = "food_poisoning"
	AilmentGIInfection   AilmentType = "gi_infection"
	AilmentDehydration   AilmentType = "dehydration"
	AilmentRespInfection AilmentType = "resp_infection"
)

type Ailment struct {
	Type             AilmentType `json:"type"`
	Name             string      `json:"name"`
	DaysRemaining    int         `json:"days_remaining"`
	EnergyPenalty    int         `json:"energy_penalty"`
	HydrationPenalty int         `json:"hydration_penalty"`
	MoralePenalty    int         `json:"morale_penalty"`
}

type AilmentTemplate struct {
	Type             AilmentType
	Name             string
	Days             int
	EnergyPenalty    int
	HydrationPenalty int
	MoralePenalty    int
}

type DiseaseRisk struct {
	ID          DiseaseID
	Name        string
	BaseChance  float64
	CarrierPart string // muscle, liver, blood, skin, respiratory, any
	Effect      AilmentTemplate
}

type AnimalSpec struct {
	ID               string
	Name             string
	Domain           AnimalDomain
	BiomeTags        []string
	WeightMinKg      float64
	WeightMaxKg      float64
	EdibleYieldRatio float64
	NutritionPer100g NutritionPer100g
	DiseaseRisks     []DiseaseRisk
}

type CatchResult struct {
	Animal      AnimalSpec
	WeightGrams int
	EdibleGrams int
}

func AnimalCatalog() []AnimalSpec {
	return []AnimalSpec{
		{
			ID:               "deer",
			Name:             "Deer",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "boreal", "coast", "lake", "mountain"},
			WeightMinKg:      35,
			WeightMaxKg:      180,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 158, ProteinG: 30, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "deer_gi", Name: "GI contamination", BaseChance: 0.06, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "rabbit",
			Name:             "Rabbit",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "mountain", "coast", "savanna", "badlands", "dry"},
			WeightMinKg:      1.0,
			WeightMaxKg:      2.8,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 173, ProteinG: 33, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "rabbit_liver_worms", Name: "Liver worms", BaseChance: 0.08, CarrierPart: "liver", Effect: AilmentTemplate{Type: AilmentVomiting, Name: "Vomiting", Days: 2, EnergyPenalty: 3, HydrationPenalty: 6, MoralePenalty: 4}},
				{ID: "rabbit_tularemia", Name: "Tularemia risk", BaseChance: 0.04, CarrierPart: "blood", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "Feverish GI Illness", Days: 3, EnergyPenalty: 4, HydrationPenalty: 4, MoralePenalty: 5}},
			},
		},
		{
			ID:               "boar",
			Name:             "Wild Boar",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "jungle", "wetlands", "swamp", "island"},
			WeightMinKg:      25,
			WeightMaxKg:      160,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 160, ProteinG: 27, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "boar_trichinella", Name: "Trichinella", BaseChance: 0.07, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 2, MoralePenalty: 4}},
			},
		},
		{
			ID:               "moose",
			Name:             "Moose",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"boreal", "subarctic", "lake", "delta"},
			WeightMinKg:      180,
			WeightMaxKg:      550,
			EdibleYieldRatio: 0.49,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 146, ProteinG: 29, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "moose_gi", Name: "Meat contamination", BaseChance: 0.03, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "mouse",
			Name:             "Mouse",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "boreal", "jungle", "desert", "savanna", "wetlands", "swamp", "mountain", "coast", "arctic"},
			WeightMinKg:      0.015,
			WeightMaxKg:      0.045,
			EdibleYieldRatio: 0.40,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 120, ProteinG: 20, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "mouse_hantavirus", Name: "Hantavirus risk", BaseChance: 0.05, CarrierPart: "respiratory", Effect: AilmentTemplate{Type: AilmentRespInfection, Name: "Respiratory Illness", Days: 3, EnergyPenalty: 3, HydrationPenalty: 2, MoralePenalty: 3}},
				{ID: "mouse_salmonella", Name: "Salmonella risk", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "trout",
			Name:             "Trout",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "mountain", "forest", "boreal"},
			WeightMinKg:      0.25,
			WeightMaxKg:      4.5,
			EdibleYieldRatio: 0.56,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 141, ProteinG: 20, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "trout_parasite", Name: "Fish parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Intestinal Parasites", Days: 3, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "salmon",
			Name:             "Salmon",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "river", "delta", "temperate_rainforest", "island"},
			WeightMinKg:      1.2,
			WeightMaxKg:      18,
			EdibleYieldRatio: 0.58,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 206, ProteinG: 22, FatG: 12},
			DiseaseRisks: []DiseaseRisk{
				{ID: "salmon_parasite", Name: "Marine parasite", BaseChance: 0.04, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "catfish",
			Name:             "Catfish",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"wetlands", "swamp", "jungle", "delta", "savanna"},
			WeightMinKg:      0.4,
			WeightMaxKg:      9,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 144, ProteinG: 18, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "catfish_contamination", Name: "Waterborne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Waterborne Food Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "duck",
			Name:             "Duck",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"lake", "delta", "wetlands", "swamp", "coast"},
			WeightMinKg:      0.7,
			WeightMaxKg:      2.1,
			EdibleYieldRatio: 0.49,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 337, ProteinG: 19, FatG: 28},
			DiseaseRisks: []DiseaseRisk{
				{ID: "duck_salmonella", Name: "Salmonella risk", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "grouse",
			Name:             "Grouse",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"boreal", "subarctic", "forest", "mountain"},
			WeightMinKg:      0.55,
			WeightMaxKg:      1.4,
			EdibleYieldRatio: 0.50,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 190, ProteinG: 25, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "grouse_gi", Name: "GI contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "pigeon",
			Name:             "Rock Pigeon",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"desert", "dry", "savanna", "badlands", "coast"},
			WeightMinKg:      0.25,
			WeightMaxKg:      0.42,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 213, ProteinG: 24, FatG: 13},
			DiseaseRisks: []DiseaseRisk{
				{ID: "pigeon_gi", Name: "Bird-borne GI contamination", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "macaw",
			Name:             "Macaw",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"jungle", "tropical", "island", "rainforest"},
			WeightMinKg:      0.8,
			WeightMaxKg:      1.8,
			EdibleYieldRatio: 0.43,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 205, ProteinG: 23, FatG: 11},
			DiseaseRisks: []DiseaseRisk{
				{ID: "macaw_gi", Name: "Bird-borne GI contamination", BaseChance: 0.07, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
	}
}

func AnimalDomains() []AnimalDomain {
	return []AnimalDomain{AnimalDomainLand, AnimalDomainWater, AnimalDomainAir}
}

func AnimalsForBiome(biome string, domain AnimalDomain) []AnimalSpec {
	norm := normalizeBiome(biome)
	all := AnimalCatalog()
	filtered := make([]AnimalSpec, 0, len(all))
	for _, animal := range all {
		if animal.Domain != domain {
			continue
		}
		if animalMatchesBiome(animal, norm) {
			filtered = append(filtered, animal)
		}
	}
	return filtered
}

func RandomCatch(seed int64, biome string, domain AnimalDomain, day, actorID int) (CatchResult, error) {
	pool := AnimalsForBiome(biome, domain)
	if len(pool) == 0 {
		return CatchResult{}, fmt.Errorf("no %s animals available in biome %s", domain, biome)
	}
	if day < 1 {
		day = 1
	}
	if actorID < 1 {
		actorID = 1
	}

	rng := seededRNG(seedFromLabel(seed, fmt.Sprintf("catch:%s:%s:%d:%d", normalizeBiome(biome), domain, day, actorID)))
	animal := pool[rng.IntN(len(pool))]

	minKg := animal.WeightMinKg
	maxKg := animal.WeightMaxKg
	if maxKg < minKg {
		maxKg = minKg
	}
	kg := minKg
	if maxKg > minKg {
		kg += rng.Float64() * (maxKg - minKg)
	}
	grams := int(math.Round(kg * 1000))
	if grams < 1 {
		grams = 1
	}

	yield := animal.EdibleYieldRatio
	if yield <= 0 || yield > 0.95 {
		yield = 0.5
	}
	edible := int(math.Round(float64(grams) * yield))
	if edible < 1 {
		edible = 1
	}

	return CatchResult{
		Animal:      animal,
		WeightGrams: grams,
		EdibleGrams: edible,
	}, nil
}

func (c CatchResult) NutritionForGrams(portionGrams int) NutritionTotals {
	if portionGrams <= 0 {
		return NutritionTotals{}
	}
	if portionGrams > c.EdibleGrams {
		portionGrams = c.EdibleGrams
	}

	per100 := c.Animal.NutritionPer100g
	return NutritionTotals{
		CaloriesKcal: int(math.Round(float64(per100.CaloriesKcal) * float64(portionGrams) / 100.0)),
		ProteinG:     int(math.Round(float64(per100.ProteinG) * float64(portionGrams) / 100.0)),
		FatG:         int(math.Round(float64(per100.FatG) * float64(portionGrams) / 100.0)),
	}
}

func animalMatchesBiome(animal AnimalSpec, normBiome string) bool {
	for _, tag := range animal.BiomeTags {
		t := strings.TrimSpace(strings.ToLower(tag))
		if t == "" {
			continue
		}
		if strings.Contains(normBiome, t) {
			return true
		}
	}
	if strings.Contains(normBiome, "subarctic") || strings.Contains(normBiome, "arctic") {
		return containsBiomeTag(animal.BiomeTags, "boreal")
	}
	return false
}

func containsBiomeTag(tags []string, needle string) bool {
	n := strings.ToLower(strings.TrimSpace(needle))
	for _, tag := range tags {
		if strings.ToLower(strings.TrimSpace(tag)) == n {
			return true
		}
	}
	return false
}

func seedFromLabel(seed int64, label string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s", seed, label)))
	return int64(h.Sum64() & 0x7fffffffffffffff)
}
