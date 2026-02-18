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
	SugarG       int
}

type NutritionTotals struct {
	CaloriesKcal int `json:"calories_kcal"`
	ProteinG     int `json:"protein_g"`
	FatG         int `json:"fat_g"`
	SugarG       int `json:"sugar_g"`
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
	AilmentEnvenomation  AilmentType = "envenomation"
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
	VomitChance float64
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
			ID:               "elk",
			Name:             "Elk",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "mountain", "boreal", "coast"},
			WeightMinKg:      110,
			WeightMaxKg:      480,
			EdibleYieldRatio: 0.48,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 150, ProteinG: 30, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "elk_gi", Name: "Field contamination", BaseChance: 0.04, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "caribou",
			Name:             "Caribou",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"arctic", "tundra", "subarctic", "boreal"},
			WeightMinKg:      60,
			WeightMaxKg:      320,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 143, ProteinG: 29, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "caribou_parasite", Name: "Caribou parasites", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 3, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
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
			ID:               "muskrat",
			Name:             "Muskrat",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"wetlands", "swamp", "delta", "lake", "river"},
			WeightMinKg:      0.6,
			WeightMaxKg:      2.3,
			EdibleYieldRatio: 0.44,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 138, ProteinG: 21, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "muskrat_lepto", Name: "Leptospirosis risk", BaseChance: 0.09, CarrierPart: "blood", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "Feverish Infection", Days: 3, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "beaver",
			Name:             "Beaver",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"lake", "delta", "river", "boreal", "forest"},
			WeightMinKg:      8,
			WeightMaxKg:      32,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 146, ProteinG: 24, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "beaver_gi", Name: "Waterborne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "hyena",
			Name:             "Hyena",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry"},
			WeightMinKg:      35,
			WeightMaxKg:      85,
			EdibleYieldRatio: 0.43,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 148, ProteinG: 27, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "hyena_gi", Name: "Scavenger pathogen load", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Severe GI Illness", Days: 3, EnergyPenalty: 4, HydrationPenalty: 5, MoralePenalty: 4}},
			},
		},
		{
			ID:               "lion",
			Name:             "Lion",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry"},
			WeightMinKg:      110,
			WeightMaxKg:      250,
			EdibleYieldRatio: 0.42,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 170, ProteinG: 28, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "lion_trichinella", Name: "Predator parasite load", BaseChance: 0.12, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 4, HydrationPenalty: 3, MoralePenalty: 5}},
			},
		},
		{
			ID:               "tiger",
			Name:             "Tiger",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "swamp", "wetlands", "tropical"},
			WeightMinKg:      80,
			WeightMaxKg:      310,
			EdibleYieldRatio: 0.42,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 168, ProteinG: 28, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "tiger_trichinella", Name: "Predator parasite load", BaseChance: 0.12, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 4, HydrationPenalty: 3, MoralePenalty: 5}},
			},
		},
		{
			ID:               "leopard",
			Name:             "Leopard",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "jungle", "badlands", "dry", "mountain"},
			WeightMinKg:      25,
			WeightMaxKg:      90,
			EdibleYieldRatio: 0.41,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 167, ProteinG: 28, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "leopard_parasites", Name: "Predator parasite load", BaseChance: 0.11, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "jaguar",
			Name:             "Jaguar",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "delta"},
			WeightMinKg:      45,
			WeightMaxKg:      120,
			EdibleYieldRatio: 0.41,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 167, ProteinG: 28, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "jaguar_parasites", Name: "Predator parasite load", BaseChance: 0.11, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "antelope",
			Name:             "Antelope",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry"},
			WeightMinKg:      22,
			WeightMaxKg:      140,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 152, ProteinG: 29, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "antelope_gi", Name: "Meat contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "warthog",
			Name:             "Warthog",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry"},
			WeightMinKg:      35,
			WeightMaxKg:      150,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 163, ProteinG: 26, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "warthog_trichinella", Name: "Trichinella", BaseChance: 0.08, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 2, MoralePenalty: 4}},
			},
		},
		{
			ID:               "mountain_goat",
			Name:             "Mountain Goat",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"mountain", "highland", "montane"},
			WeightMinKg:      50,
			WeightMaxKg:      120,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 142, ProteinG: 27, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "goat_gi", Name: "GI contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "capybara",
			Name:             "Capybara",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"wetlands", "swamp", "jungle", "delta"},
			WeightMinKg:      20,
			WeightMaxKg:      75,
			EdibleYieldRatio: 0.48,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 172, ProteinG: 24, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "capybara_gi", Name: "Waterborne contamination", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "iguana",
			Name:             "Iguana",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "tropical", "island", "dry_forest"},
			WeightMinKg:      1.2,
			WeightMaxKg:      6,
			EdibleYieldRatio: 0.42,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 119, ProteinG: 20, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "iguana_salmonella", Name: "Salmonella risk", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "monitor_lizard",
			Name:             "Monitor Lizard",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "jungle", "wetlands", "swamp", "island"},
			WeightMinKg:      3,
			WeightMaxKg:      65,
			EdibleYieldRatio: 0.41,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 126, ProteinG: 22, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "monitor_salmonella", Name: "Reptile salmonella", BaseChance: 0.13, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "rattlesnake",
			Name:             "Rattlesnake",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"desert", "dry", "savanna", "badlands"},
			WeightMinKg:      0.3,
			WeightMaxKg:      2.8,
			EdibleYieldRatio: 0.36,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 93, ProteinG: 20, FatG: 1},
			DiseaseRisks: []DiseaseRisk{
				{ID: "rattlesnake_venom", Name: "Venom contamination", BaseChance: 0.18, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Envenomation", Days: 2, EnergyPenalty: 4, HydrationPenalty: 4, MoralePenalty: 5}},
			},
		},
		{
			ID:               "cobra",
			Name:             "Cobra",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "savanna", "badlands", "tropical", "wetlands"},
			WeightMinKg:      0.7,
			WeightMaxKg:      6.8,
			EdibleYieldRatio: 0.35,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 95, ProteinG: 20, FatG: 1},
			DiseaseRisks: []DiseaseRisk{
				{ID: "cobra_venom", Name: "Neurotoxic venom risk", BaseChance: 0.22, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Severe Envenomation", Days: 3, EnergyPenalty: 5, HydrationPenalty: 5, MoralePenalty: 6}},
			},
		},
		{
			ID:               "python",
			Name:             "Python",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "swamp", "wetlands", "tropical", "island"},
			WeightMinKg:      8,
			WeightMaxKg:      95,
			EdibleYieldRatio: 0.40,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 96, ProteinG: 20, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "python_salmonella", Name: "Reptile salmonella", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "boa_constrictor",
			Name:             "Boa Constrictor",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "tropical", "river"},
			WeightMinKg:      4,
			WeightMaxKg:      35,
			EdibleYieldRatio: 0.39,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 96, ProteinG: 20, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "boa_salmonella", Name: "Reptile salmonella", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "sea_snake",
			Name:             "Sea Snake",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"coast", "island", "delta", "tropical"},
			WeightMinKg:      0.4,
			WeightMaxKg:      2.5,
			EdibleYieldRatio: 0.34,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 94, ProteinG: 19, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "sea_snake_venom", Name: "Venom contamination", BaseChance: 0.24, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Severe Envenomation", Days: 3, EnergyPenalty: 5, HydrationPenalty: 5, MoralePenalty: 6}},
			},
		},
		{
			ID:               "tarantula",
			Name:             "Tarantula",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "savanna", "badlands", "desert", "dry"},
			WeightMinKg:      0.02,
			WeightMaxKg:      0.18,
			EdibleYieldRatio: 0.32,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 135, ProteinG: 25, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "tarantula_venom", Name: "Venom contamination", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Envenomation", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 3}},
			},
		},
		{
			ID:               "scorpion",
			Name:             "Scorpion",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"desert", "dry", "savanna", "badlands", "tropical_dry_forest"},
			WeightMinKg:      0.01,
			WeightMaxKg:      0.09,
			EdibleYieldRatio: 0.30,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 121, ProteinG: 22, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "scorpion_venom", Name: "Venom contamination", BaseChance: 0.15, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Envenomation", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 4}},
			},
		},
		{
			ID:               "wolf_spider",
			Name:             "Wolf Spider",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "jungle", "savanna", "badlands", "swamp"},
			WeightMinKg:      0.005,
			WeightMaxKg:      0.03,
			EdibleYieldRatio: 0.28,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 128, ProteinG: 24, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "spider_venom", Name: "Venom contamination", BaseChance: 0.10, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentEnvenomation, Name: "Mild Envenomation", Days: 1, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "alligator",
			Name:             "Alligator",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"swamp", "wetlands", "delta", "coast"},
			WeightMinKg:      35,
			WeightMaxKg:      450,
			EdibleYieldRatio: 0.44,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 143, ProteinG: 29, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "alligator_salmonella", Name: "Salmonella risk", BaseChance: 0.10, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "caiman",
			Name:             "Caiman",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "delta"},
			WeightMinKg:      8,
			WeightMaxKg:      250,
			EdibleYieldRatio: 0.43,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 140, ProteinG: 28, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "caiman_salmonella", Name: "Salmonella risk", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "crocodile",
			Name:             "Crocodile",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "delta", "savanna"},
			WeightMinKg:      90,
			WeightMaxKg:      900,
			EdibleYieldRatio: 0.44,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 146, ProteinG: 29, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "crocodile_salmonella", Name: "Salmonella risk", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
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
			ID:               "northern_pike",
			Name:             "Northern Pike",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "boreal", "subarctic", "delta"},
			WeightMinKg:      0.8,
			WeightMaxKg:      21,
			EdibleYieldRatio: 0.55,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 113, ProteinG: 20, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "pike_parasite", Name: "Freshwater parasite", BaseChance: 0.06, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Intestinal Parasites", Days: 3, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "perch",
			Name:             "Perch",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "delta", "forest", "boreal"},
			WeightMinKg:      0.15,
			WeightMaxKg:      2.2,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 117, ProteinG: 24, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "perch_parasite", Name: "Fish parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "walleye",
			Name:             "Walleye",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "delta", "boreal", "subarctic"},
			WeightMinKg:      0.4,
			WeightMaxKg:      9,
			EdibleYieldRatio: 0.56,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 111, ProteinG: 21, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "walleye_parasite", Name: "Freshwater parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "char",
			Name:             "Arctic Char",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"arctic", "tundra", "subarctic", "lake", "delta"},
			WeightMinKg:      0.3,
			WeightMaxKg:      8,
			EdibleYieldRatio: 0.57,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 172, ProteinG: 21, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "char_parasite", Name: "Cold-water parasite", BaseChance: 0.04, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "grayling",
			Name:             "Grayling",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"river", "subarctic", "boreal", "mountain", "lake"},
			WeightMinKg:      0.15,
			WeightMaxKg:      2.6,
			EdibleYieldRatio: 0.54,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 122, ProteinG: 20, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "grayling_parasite", Name: "Fish parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Intestinal Parasites", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "carp",
			Name:             "Carp",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "delta", "river", "wetlands", "savanna"},
			WeightMinKg:      0.9,
			WeightMaxKg:      35,
			EdibleYieldRatio: 0.54,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 127, ProteinG: 18, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "carp_contamination", Name: "Water contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Waterborne GI Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "largemouth_bass",
			Name:             "Largemouth Bass",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "wetlands", "swamp", "delta"},
			WeightMinKg:      0.4,
			WeightMaxKg:      10,
			EdibleYieldRatio: 0.55,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 130, ProteinG: 20, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "bass_parasite", Name: "Fish parasite", BaseChance: 0.06, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "sturgeon",
			Name:             "Sturgeon",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"delta", "river", "coast", "lake"},
			WeightMinKg:      3,
			WeightMaxKg:      250,
			EdibleYieldRatio: 0.60,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 135, ProteinG: 19, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "sturgeon_parasite", Name: "Fish parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "tilapia",
			Name:             "Tilapia",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "savanna", "delta"},
			WeightMinKg:      0.2,
			WeightMaxKg:      4.5,
			EdibleYieldRatio: 0.54,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 129, ProteinG: 20, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "tilapia_contamination", Name: "Water contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Waterborne Food Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "piranha",
			Name:             "Piranha",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "delta", "river"},
			WeightMinKg:      0.2,
			WeightMaxKg:      2,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 119, ProteinG: 21, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "piranha_parasite", Name: "Fish parasite", BaseChance: 0.07, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "arapaima",
			Name:             "Arapaima",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"jungle", "wetlands", "swamp", "delta", "river"},
			WeightMinKg:      8,
			WeightMaxKg:      200,
			EdibleYieldRatio: 0.61,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 121, ProteinG: 22, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "arapaima_parasite", Name: "Fish parasite", BaseChance: 0.06, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "eel",
			Name:             "Eel",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "swamp", "river", "lake"},
			WeightMinKg:      0.2,
			WeightMaxKg:      8,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 184, ProteinG: 18, FatG: 12},
			DiseaseRisks: []DiseaseRisk{
				{ID: "eel_parasite", Name: "Fish parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "cod",
			Name:             "Cod",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "island"},
			WeightMinKg:      0.8,
			WeightMaxKg:      40,
			EdibleYieldRatio: 0.59,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 105, ProteinG: 23, FatG: 1},
			DiseaseRisks: []DiseaseRisk{
				{ID: "cod_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "mackerel",
			Name:             "Mackerel",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "delta"},
			WeightMinKg:      0.15,
			WeightMaxKg:      4,
			EdibleYieldRatio: 0.58,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 205, ProteinG: 19, FatG: 13},
			DiseaseRisks: []DiseaseRisk{
				{ID: "mackerel_contamination", Name: "Scombroid contamination", BaseChance: 0.06, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "sardine",
			Name:             "Sardine",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "delta"},
			WeightMinKg:      0.02,
			WeightMaxKg:      0.18,
			EdibleYieldRatio: 0.61,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 208, ProteinG: 25, FatG: 11},
			DiseaseRisks: []DiseaseRisk{
				{ID: "sardine_contamination", Name: "Marine contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 1, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "tuna",
			Name:             "Tuna",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island"},
			WeightMinKg:      4,
			WeightMaxKg:      250,
			EdibleYieldRatio: 0.63,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 132, ProteinG: 29, FatG: 1},
			DiseaseRisks: []DiseaseRisk{
				{ID: "tuna_histamine", Name: "Histamine contamination", BaseChance: 0.04, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "red_snapper",
			Name:             "Red Snapper",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "delta", "tropical"},
			WeightMinKg:      0.6,
			WeightMaxKg:      18,
			EdibleYieldRatio: 0.56,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 128, ProteinG: 26, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "snapper_ciguatera", Name: "Ciguatera risk", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentVomiting, Name: "Ciguatera-like Illness", Days: 3, EnergyPenalty: 4, HydrationPenalty: 5, MoralePenalty: 4}},
			},
		},
		{
			ID:               "grouper_fish",
			Name:             "Grouper",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "tropical"},
			WeightMinKg:      1.2,
			WeightMaxKg:      110,
			EdibleYieldRatio: 0.59,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 118, ProteinG: 24, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "grouper_ciguatera", Name: "Ciguatera risk", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentVomiting, Name: "Ciguatera-like Illness", Days: 3, EnergyPenalty: 4, HydrationPenalty: 5, MoralePenalty: 4}},
			},
		},
		{
			ID:               "barracuda",
			Name:             "Barracuda",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "tropical"},
			WeightMinKg:      1.5,
			WeightMaxKg:      45,
			EdibleYieldRatio: 0.57,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 124, ProteinG: 21, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "barracuda_ciguatera", Name: "Ciguatera risk", BaseChance: 0.08, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentVomiting, Name: "Ciguatera-like Illness", Days: 3, EnergyPenalty: 4, HydrationPenalty: 5, MoralePenalty: 4}},
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
		{
			ID:               "goose",
			Name:             "Goose",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"lake", "delta", "wetlands", "coast", "tundra"},
			WeightMinKg:      2,
			WeightMaxKg:      7,
			EdibleYieldRatio: 0.51,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 305, ProteinG: 25, FatG: 21},
			DiseaseRisks: []DiseaseRisk{
				{ID: "goose_salmonella", Name: "Salmonella risk", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "ptarmigan",
			Name:             "Ptarmigan",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"arctic", "tundra", "subarctic", "boreal"},
			WeightMinKg:      0.35,
			WeightMaxKg:      0.95,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 183, ProteinG: 24, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "ptarmigan_gi", Name: "GI contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "quail",
			Name:             "Quail",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"savanna", "badlands", "dry", "forest", "coast"},
			WeightMinKg:      0.12,
			WeightMaxKg:      0.35,
			EdibleYieldRatio: 0.44,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 134, ProteinG: 21, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "quail_salmonella", Name: "Salmonella risk", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "pheasant",
			Name:             "Pheasant",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"forest", "savanna", "badlands", "mountain", "coast"},
			WeightMinKg:      0.7,
			WeightMaxKg:      2,
			EdibleYieldRatio: 0.48,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 181, ProteinG: 25, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "pheasant_gi", Name: "GI contamination", BaseChance: 0.06, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "wild_turkey",
			Name:             "Wild Turkey",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"forest", "mountain", "wetlands", "savanna"},
			WeightMinKg:      2.5,
			WeightMaxKg:      11,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 189, ProteinG: 29, FatG: 7},
			DiseaseRisks: []DiseaseRisk{
				{ID: "turkey_salmonella", Name: "Salmonella risk", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "seagull",
			Name:             "Seagull",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"coast", "island", "delta", "lake"},
			WeightMinKg:      0.35,
			WeightMaxKg:      1.8,
			EdibleYieldRatio: 0.42,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 202, ProteinG: 22, FatG: 11},
			DiseaseRisks: []DiseaseRisk{
				{ID: "seagull_gi", Name: "Scavenger contamination", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "black_bear",
			Name:             "Black Bear",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "boreal", "mountain", "coast", "lake"},
			WeightMinKg:      60,
			WeightMaxKg:      300,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 155, ProteinG: 27, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "black_bear_trichinella", Name: "Trichinella risk", BaseChance: 0.14, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 4, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "brown_bear",
			Name:             "Brown Bear",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"boreal", "subarctic", "arctic", "mountain", "delta"},
			WeightMinKg:      110,
			WeightMaxKg:      650,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 158, ProteinG: 27, FatG: 6},
			DiseaseRisks: []DiseaseRisk{
				{ID: "brown_bear_trichinella", Name: "Trichinella risk", BaseChance: 0.15, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 4, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "cougar",
			Name:             "Cougar",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"mountain", "forest", "coast", "badlands"},
			WeightMinKg:      28,
			WeightMaxKg:      100,
			EdibleYieldRatio: 0.40,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 164, ProteinG: 28, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "cougar_parasites", Name: "Predator parasite load", BaseChance: 0.11, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "wolf",
			Name:             "Wolf",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"boreal", "subarctic", "forest", "mountain", "tundra"},
			WeightMinKg:      25,
			WeightMaxKg:      75,
			EdibleYieldRatio: 0.41,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 162, ProteinG: 27, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "wolf_parasites", Name: "Predator parasite load", BaseChance: 0.10, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 4, EnergyPenalty: 3, HydrationPenalty: 3, MoralePenalty: 4}},
			},
		},
		{
			ID:               "coyote",
			Name:             "Coyote",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"desert", "dry", "badlands", "savanna", "mountain"},
			WeightMinKg:      8,
			WeightMaxKg:      25,
			EdibleYieldRatio: 0.39,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 160, ProteinG: 27, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "coyote_gi", Name: "Scavenger pathogen load", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "fox",
			Name:             "Fox",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"forest", "boreal", "desert", "dry", "mountain"},
			WeightMinKg:      2.5,
			WeightMaxKg:      14,
			EdibleYieldRatio: 0.38,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 156, ProteinG: 26, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "fox_parasites", Name: "Parasite load", BaseChance: 0.10, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 3, EnergyPenalty: 3, HydrationPenalty: 2, MoralePenalty: 3}},
			},
		},
		{
			ID:               "jackal",
			Name:             "Jackal",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry", "desert"},
			WeightMinKg:      6,
			WeightMaxKg:      16,
			EdibleYieldRatio: 0.38,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 157, ProteinG: 26, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "jackal_gi", Name: "Scavenger pathogen load", BaseChance: 0.11, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "dingo",
			Name:             "Dingo",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry", "desert", "island"},
			WeightMinKg:      11,
			WeightMaxKg:      24,
			EdibleYieldRatio: 0.39,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 158, ProteinG: 26, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "dingo_gi", Name: "Pathogen load", BaseChance: 0.10, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "bison",
			Name:             "Bison",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry", "mountain"},
			WeightMinKg:      320,
			WeightMaxKg:      950,
			EdibleYieldRatio: 0.50,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 143, ProteinG: 28, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "bison_gi", Name: "Field contamination", BaseChance: 0.04, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "camel",
			Name:             "Camel",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"desert", "dry", "savanna", "badlands"},
			WeightMinKg:      320,
			WeightMaxKg:      700,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 148, ProteinG: 27, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "camel_gi", Name: "GI contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "kangaroo",
			Name:             "Kangaroo",
			Domain:           AnimalDomainLand,
			BiomeTags:        []string{"savanna", "badlands", "dry", "desert"},
			WeightMinKg:      18,
			WeightMaxKg:      90,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 98, ProteinG: 22, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "kangaroo_gi", Name: "Field contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentGIInfection, Name: "GI Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "bluegill",
			Name:             "Bluegill",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "wetlands", "swamp", "delta"},
			WeightMinKg:      0.08,
			WeightMaxKg:      0.9,
			EdibleYieldRatio: 0.50,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 120, ProteinG: 21, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "bluegill_parasite", Name: "Fish parasite", BaseChance: 0.06, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "crappie",
			Name:             "Crappie",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"lake", "river", "wetlands", "delta"},
			WeightMinKg:      0.1,
			WeightMaxKg:      1.5,
			EdibleYieldRatio: 0.52,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 118, ProteinG: 20, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "crappie_parasite", Name: "Fish parasite", BaseChance: 0.06, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "halibut",
			Name:             "Halibut",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "delta"},
			WeightMinKg:      2,
			WeightMaxKg:      220,
			EdibleYieldRatio: 0.63,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 111, ProteinG: 23, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "halibut_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "flounder",
			Name:             "Flounder",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "island"},
			WeightMinKg:      0.2,
			WeightMaxKg:      8,
			EdibleYieldRatio: 0.56,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 117, ProteinG: 24, FatG: 2},
			DiseaseRisks: []DiseaseRisk{
				{ID: "flounder_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic GI Upset", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "herring",
			Name:             "Herring",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "island"},
			WeightMinKg:      0.04,
			WeightMaxKg:      0.7,
			EdibleYieldRatio: 0.57,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 158, ProteinG: 18, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "herring_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "anchovy",
			Name:             "Anchovy",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "island"},
			WeightMinKg:      0.005,
			WeightMaxKg:      0.08,
			EdibleYieldRatio: 0.58,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 131, ProteinG: 20, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "anchovy_contamination", Name: "Marine contamination", BaseChance: 0.04, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 1, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "mahi_mahi",
			Name:             "Mahi-Mahi",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island", "tropical"},
			WeightMinKg:      1.5,
			WeightMaxKg:      38,
			EdibleYieldRatio: 0.61,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 134, ProteinG: 23, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "mahi_histamine", Name: "Histamine contamination", BaseChance: 0.05, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 2, HydrationPenalty: 3, MoralePenalty: 2}},
			},
		},
		{
			ID:               "tarpon",
			Name:             "Tarpon",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "delta", "island", "wetlands"},
			WeightMinKg:      3,
			WeightMaxKg:      130,
			EdibleYieldRatio: 0.56,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 126, ProteinG: 22, FatG: 3},
			DiseaseRisks: []DiseaseRisk{
				{ID: "tarpon_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "swordfish",
			Name:             "Swordfish",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"coast", "island"},
			WeightMinKg:      20,
			WeightMaxKg:      540,
			EdibleYieldRatio: 0.62,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 172, ProteinG: 20, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "swordfish_parasite", Name: "Marine parasite", BaseChance: 0.05, CarrierPart: "muscle", Effect: AilmentTemplate{Type: AilmentParasites, Name: "Parasitic Infection", Days: 2, EnergyPenalty: 2, HydrationPenalty: 2, MoralePenalty: 2}},
			},
		},
		{
			ID:               "catla",
			Name:             "Catla",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"river", "delta", "wetlands", "jungle", "tropical"},
			WeightMinKg:      1,
			WeightMaxKg:      45,
			EdibleYieldRatio: 0.55,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 127, ProteinG: 19, FatG: 5},
			DiseaseRisks: []DiseaseRisk{
				{ID: "catla_contamination", Name: "Water contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Waterborne GI Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "rohu",
			Name:             "Rohu",
			Domain:           AnimalDomainWater,
			BiomeTags:        []string{"river", "delta", "wetlands", "jungle", "tropical"},
			WeightMinKg:      0.8,
			WeightMaxKg:      18,
			EdibleYieldRatio: 0.55,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 123, ProteinG: 20, FatG: 4},
			DiseaseRisks: []DiseaseRisk{
				{ID: "rohu_contamination", Name: "Water contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Waterborne GI Illness", Days: 2, EnergyPenalty: 2, HydrationPenalty: 4, MoralePenalty: 2}},
			},
		},
		{
			ID:               "crow",
			Name:             "Crow",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"forest", "savanna", "badlands", "coast", "wetlands"},
			WeightMinKg:      0.25,
			WeightMaxKg:      0.65,
			EdibleYieldRatio: 0.40,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 186, ProteinG: 23, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "crow_gi", Name: "Scavenger contamination", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "raven",
			Name:             "Raven",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"boreal", "subarctic", "mountain", "coast", "tundra"},
			WeightMinKg:      0.4,
			WeightMaxKg:      1.6,
			EdibleYieldRatio: 0.42,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 188, ProteinG: 23, FatG: 9},
			DiseaseRisks: []DiseaseRisk{
				{ID: "raven_gi", Name: "Scavenger contamination", BaseChance: 0.12, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "GI Food Illness", Days: 2, EnergyPenalty: 3, HydrationPenalty: 4, MoralePenalty: 3}},
			},
		},
		{
			ID:               "heron",
			Name:             "Heron",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"wetlands", "swamp", "delta", "lake", "coast"},
			WeightMinKg:      1.0,
			WeightMaxKg:      3.8,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 170, ProteinG: 24, FatG: 7},
			DiseaseRisks: []DiseaseRisk{
				{ID: "heron_salmonella", Name: "Bird-borne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "egret",
			Name:             "Egret",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"wetlands", "swamp", "delta", "coast", "jungle"},
			WeightMinKg:      0.7,
			WeightMaxKg:      1.8,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 166, ProteinG: 23, FatG: 7},
			DiseaseRisks: []DiseaseRisk{
				{ID: "egret_salmonella", Name: "Bird-borne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "pelican",
			Name:             "Pelican",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"coast", "island", "delta", "wetlands"},
			WeightMinKg:      2.5,
			WeightMaxKg:      14,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 176, ProteinG: 23, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "pelican_gi", Name: "Bird-borne contamination", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "cormorant",
			Name:             "Cormorant",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"coast", "island", "delta", "lake"},
			WeightMinKg:      1.2,
			WeightMaxKg:      4,
			EdibleYieldRatio: 0.45,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 174, ProteinG: 24, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "cormorant_gi", Name: "Bird-borne contamination", BaseChance: 0.09, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "kingfisher",
			Name:             "Kingfisher",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"river", "lake", "delta", "wetlands", "jungle"},
			WeightMinKg:      0.03,
			WeightMaxKg:      0.25,
			EdibleYieldRatio: 0.38,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 171, ProteinG: 23, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "kingfisher_gi", Name: "Bird-borne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "partridge",
			Name:             "Partridge",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"forest", "mountain", "savanna", "badlands", "dry"},
			WeightMinKg:      0.25,
			WeightMaxKg:      0.9,
			EdibleYieldRatio: 0.46,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 180, ProteinG: 24, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "partridge_gi", Name: "Bird-borne contamination", BaseChance: 0.07, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "dove",
			Name:             "Dove",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"forest", "savanna", "badlands", "dry", "coast"},
			WeightMinKg:      0.09,
			WeightMaxKg:      0.35,
			EdibleYieldRatio: 0.40,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 172, ProteinG: 23, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "dove_gi", Name: "Bird-borne contamination", BaseChance: 0.07, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
			},
		},
		{
			ID:               "albatross",
			Name:             "Albatross",
			Domain:           AnimalDomainAir,
			BiomeTags:        []string{"coast", "island"},
			WeightMinKg:      2.5,
			WeightMaxKg:      12,
			EdibleYieldRatio: 0.47,
			NutritionPer100g: NutritionPer100g{CaloriesKcal: 178, ProteinG: 23, FatG: 8},
			DiseaseRisks: []DiseaseRisk{
				{ID: "albatross_gi", Name: "Bird-borne contamination", BaseChance: 0.08, CarrierPart: "any", Effect: AilmentTemplate{Type: AilmentFoodPoison, Name: "Food Poisoning", Days: 2, EnergyPenalty: 3, HydrationPenalty: 5, MoralePenalty: 3}},
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
		SugarG:       int(math.Round(float64(per100.SugarG) * float64(portionGrams) / 100.0)),
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
