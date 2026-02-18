package game

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

type PlantCategory string

const (
	PlantCategoryAny       PlantCategory = "any"
	PlantCategoryRoots     PlantCategory = "roots"
	PlantCategoryBerries   PlantCategory = "berries"
	PlantCategoryFruits    PlantCategory = "fruits"
	PlantCategoryVegetable PlantCategory = "vegetables"
)

type PlantSpec struct {
	ID               string
	Name             string
	Category         PlantCategory
	BiomeTags        []string
	YieldMinG        int
	YieldMaxG        int
	NutritionPer100g NutritionPer100g
}

type ForageResult struct {
	Plant        PlantSpec
	HarvestGrams int
	Nutrition    NutritionTotals
}

func PlantCatalog() []PlantSpec {
	return []PlantSpec{
		{ID: "burdock_root", Name: "Burdock Root", Category: PlantCategoryRoots, BiomeTags: []string{"forest", "temperate", "mountain"}, YieldMinG: 120, YieldMaxG: 600, NutritionPer100g: NutritionPer100g{CaloriesKcal: 72, ProteinG: 1, FatG: 0, SugarG: 2}},
		{ID: "cattail_root", Name: "Cattail Rhizome", Category: PlantCategoryRoots, BiomeTags: []string{"wetlands", "swamp", "delta", "lake"}, YieldMinG: 200, YieldMaxG: 900, NutritionPer100g: NutritionPer100g{CaloriesKcal: 80, ProteinG: 2, FatG: 0, SugarG: 3}},
		{ID: "yuca_root", Name: "Yuca Root", Category: PlantCategoryRoots, BiomeTags: []string{"jungle", "savanna", "tropical"}, YieldMinG: 180, YieldMaxG: 1100, NutritionPer100g: NutritionPer100g{CaloriesKcal: 160, ProteinG: 1, FatG: 0, SugarG: 2}},
		{ID: "wild_turnip", Name: "Wild Turnip", Category: PlantCategoryRoots, BiomeTags: []string{"forest", "boreal", "mountain"}, YieldMinG: 100, YieldMaxG: 500, NutritionPer100g: NutritionPer100g{CaloriesKcal: 38, ProteinG: 1, FatG: 0, SugarG: 4}},
		{ID: "desert_tuber", Name: "Desert Tuber", Category: PlantCategoryRoots, BiomeTags: []string{"desert", "dry", "badlands"}, YieldMinG: 90, YieldMaxG: 320, NutritionPer100g: NutritionPer100g{CaloriesKcal: 93, ProteinG: 2, FatG: 0, SugarG: 3}},

		{ID: "blueberry", Name: "Blueberry", Category: PlantCategoryBerries, BiomeTags: []string{"forest", "boreal", "mountain", "lake"}, YieldMinG: 80, YieldMaxG: 450, NutritionPer100g: NutritionPer100g{CaloriesKcal: 57, ProteinG: 1, FatG: 0, SugarG: 10}},
		{ID: "salmonberry", Name: "Salmonberry", Category: PlantCategoryBerries, BiomeTags: []string{"coast", "temperate_rainforest", "vancouver"}, YieldMinG: 70, YieldMaxG: 380, NutritionPer100g: NutritionPer100g{CaloriesKcal: 52, ProteinG: 1, FatG: 0, SugarG: 9}},
		{ID: "crowberry", Name: "Crowberry", Category: PlantCategoryBerries, BiomeTags: []string{"arctic", "tundra", "subarctic", "boreal"}, YieldMinG: 50, YieldMaxG: 260, NutritionPer100g: NutritionPer100g{CaloriesKcal: 48, ProteinG: 1, FatG: 0, SugarG: 6}},
		{ID: "blackberry", Name: "Blackberry", Category: PlantCategoryBerries, BiomeTags: []string{"forest", "coast", "mountain", "river"}, YieldMinG: 80, YieldMaxG: 500, NutritionPer100g: NutritionPer100g{CaloriesKcal: 43, ProteinG: 1, FatG: 0, SugarG: 5}},
		{ID: "desert_berry", Name: "Desert Wolfberry", Category: PlantCategoryBerries, BiomeTags: []string{"desert", "dry", "savanna"}, YieldMinG: 45, YieldMaxG: 180, NutritionPer100g: NutritionPer100g{CaloriesKcal: 70, ProteinG: 2, FatG: 0, SugarG: 12}},

		{ID: "wild_apple", Name: "Wild Apple", Category: PlantCategoryFruits, BiomeTags: []string{"forest", "mountain", "temperate"}, YieldMinG: 150, YieldMaxG: 900, NutritionPer100g: NutritionPer100g{CaloriesKcal: 52, ProteinG: 0, FatG: 0, SugarG: 10}},
		{ID: "coconut", Name: "Coconut", Category: PlantCategoryFruits, BiomeTags: []string{"island", "coast", "tropical"}, YieldMinG: 180, YieldMaxG: 1500, NutritionPer100g: NutritionPer100g{CaloriesKcal: 354, ProteinG: 3, FatG: 33, SugarG: 6}},
		{ID: "plantain", Name: "Plantain", Category: PlantCategoryFruits, BiomeTags: []string{"jungle", "wetlands", "tropical"}, YieldMinG: 180, YieldMaxG: 1200, NutritionPer100g: NutritionPer100g{CaloriesKcal: 122, ProteinG: 1, FatG: 0, SugarG: 15}},
		{ID: "baobab_fruit", Name: "Baobab Fruit", Category: PlantCategoryFruits, BiomeTags: []string{"savanna", "badlands", "dry"}, YieldMinG: 60, YieldMaxG: 520, NutritionPer100g: NutritionPer100g{CaloriesKcal: 230, ProteinG: 2, FatG: 1, SugarG: 26}},
		{ID: "sea_grape", Name: "Sea Grape", Category: PlantCategoryFruits, BiomeTags: []string{"coast", "island", "delta"}, YieldMinG: 100, YieldMaxG: 420, NutritionPer100g: NutritionPer100g{CaloriesKcal: 67, ProteinG: 1, FatG: 0, SugarG: 15}},

		{ID: "wild_spinach", Name: "Wild Spinach", Category: PlantCategoryVegetable, BiomeTags: []string{"forest", "river", "lake", "mountain"}, YieldMinG: 100, YieldMaxG: 500, NutritionPer100g: NutritionPer100g{CaloriesKcal: 23, ProteinG: 3, FatG: 0, SugarG: 0}},
		{ID: "watercress", Name: "Watercress", Category: PlantCategoryVegetable, BiomeTags: []string{"wetlands", "swamp", "delta", "river", "lake"}, YieldMinG: 120, YieldMaxG: 650, NutritionPer100g: NutritionPer100g{CaloriesKcal: 11, ProteinG: 2, FatG: 0, SugarG: 0}},
		{ID: "wild_sorrel", Name: "Wild Sorrel", Category: PlantCategoryVegetable, BiomeTags: []string{"boreal", "subarctic", "forest", "arctic"}, YieldMinG: 80, YieldMaxG: 340, NutritionPer100g: NutritionPer100g{CaloriesKcal: 22, ProteinG: 2, FatG: 0, SugarG: 2}},
		{ID: "prickly_pear_pad", Name: "Prickly Pear Pad", Category: PlantCategoryVegetable, BiomeTags: []string{"desert", "dry", "badlands"}, YieldMinG: 120, YieldMaxG: 550, NutritionPer100g: NutritionPer100g{CaloriesKcal: 16, ProteinG: 1, FatG: 0, SugarG: 1}},
		{ID: "bamboo_shoot", Name: "Bamboo Shoot", Category: PlantCategoryVegetable, BiomeTags: []string{"jungle", "tropical", "wetlands"}, YieldMinG: 200, YieldMaxG: 900, NutritionPer100g: NutritionPer100g{CaloriesKcal: 27, ProteinG: 3, FatG: 0, SugarG: 3}},
	}
}

func PlantCategories() []PlantCategory {
	return []PlantCategory{
		PlantCategoryAny,
		PlantCategoryRoots,
		PlantCategoryBerries,
		PlantCategoryFruits,
		PlantCategoryVegetable,
	}
}

func ParsePlantCategory(raw string) PlantCategory {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "", "any", "all":
		return PlantCategoryAny
	case "roots", "root", "tubers", "tuber":
		return PlantCategoryRoots
	case "berries", "berry":
		return PlantCategoryBerries
	case "fruits", "fruit":
		return PlantCategoryFruits
	case "vegetables", "vegetable", "veg", "greens":
		return PlantCategoryVegetable
	default:
		return PlantCategoryAny
	}
}

func PlantsForBiome(biome string, category PlantCategory) []PlantSpec {
	norm := normalizeBiome(biome)
	catalog := PlantCatalog()
	filtered := make([]PlantSpec, 0, len(catalog))
	for _, plant := range catalog {
		if category != PlantCategoryAny && plant.Category != category {
			continue
		}
		for _, tag := range plant.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				filtered = append(filtered, plant)
				break
			}
		}
	}

	if len(filtered) == 0 && category != PlantCategoryAny {
		return PlantsForBiome(biome, PlantCategoryAny)
	}
	return filtered
}

func RandomForage(seed int64, biome string, category PlantCategory, day, actorID int) (ForageResult, error) {
	pool := PlantsForBiome(biome, category)
	if len(pool) == 0 {
		return ForageResult{}, fmt.Errorf("no plants available in biome %s", biome)
	}
	if day < 1 {
		day = 1
	}
	if actorID < 1 {
		actorID = 1
	}

	rng := seededRNG(seedFromLabel(seed, fmt.Sprintf("forage:%s:%s:%d:%d", normalizeBiome(biome), category, day, actorID)))
	plant := pool[rng.IntN(len(pool))]
	grams := plant.YieldMinG
	if plant.YieldMaxG > plant.YieldMinG {
		grams = plant.YieldMinG + rng.IntN(plant.YieldMaxG-plant.YieldMinG+1)
	}
	if grams < 1 {
		grams = 1
	}
	return ForageResult{
		Plant:        plant,
		HarvestGrams: grams,
		Nutrition:    nutritionFromPer100g(plant.NutritionPer100g, grams),
	}, nil
}

func (s *RunState) ForageAndConsume(playerID int, category PlantCategory, grams int) (ForageResult, error) {
	if s == nil {
		return ForageResult{}, fmt.Errorf("run state is nil")
	}
	s.EnsurePlayerRuntimeStats()

	player, ok := s.playerByID(playerID)
	if !ok {
		return ForageResult{}, fmt.Errorf("player %d not found", playerID)
	}

	forage, err := RandomForage(s.Config.Seed, s.Scenario.Biome, category, s.Day, playerID)
	if err != nil {
		return ForageResult{}, err
	}
	if grams <= 0 || grams > forage.HarvestGrams {
		grams = forage.HarvestGrams
	}
	forage.HarvestGrams = grams
	forage.Nutrition = nutritionFromPer100g(forage.Plant.NutritionPer100g, grams)

	applyMealNutritionReserves(player, forage.Nutrition)
	player.Nutrition = player.Nutrition.add(forage.Nutrition)
	energyGain, hydrationGain, moraleGain := nutritionToPlayerEffects(forage.Nutrition)
	player.Energy = clamp(player.Energy+energyGain, 0, 100)
	player.Hydration = clamp(player.Hydration+hydrationGain, 0, 100)
	player.Morale = clamp(player.Morale+moraleGain, 0, 100)
	refreshEffectBars(player)

	return forage, nil
}

func nutritionFromPer100g(per100 NutritionPer100g, grams int) NutritionTotals {
	if grams <= 0 {
		return NutritionTotals{}
	}
	return NutritionTotals{
		CaloriesKcal: int(math.Round(float64(per100.CaloriesKcal) * float64(grams) / 100.0)),
		ProteinG:     int(math.Round(float64(per100.ProteinG) * float64(grams) / 100.0)),
		FatG:         int(math.Round(float64(per100.FatG) * float64(grams) / 100.0)),
		SugarG:       int(math.Round(float64(per100.SugarG) * float64(grams) / 100.0)),
	}
}

type WoodType string

const (
	WoodTypeHardwood  WoodType = "hardwood"
	WoodTypeSoftwood  WoodType = "softwood"
	WoodTypeResinous  WoodType = "resinous"
	WoodTypeBamboo    WoodType = "bamboo"
	WoodTypeDriftwood WoodType = "driftwood"
)

type TreeSpec struct {
	ID          string
	Name        string
	BiomeTags   []string
	WoodType    WoodType
	GatherMinKg float64
	GatherMaxKg float64
	HeatFactor  float64
	BurnFactor  float64
	SparkEase   int
}

func TreeCatalog() []TreeSpec {
	return []TreeSpec{
		{ID: "cedar", Name: "Cedar", BiomeTags: []string{"coast", "temperate_rainforest", "vancouver", "forest"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.8, GatherMaxKg: 4.2, HeatFactor: 0.85, BurnFactor: 0.8, SparkEase: 3},
		{ID: "spruce", Name: "Spruce", BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, WoodType: WoodTypeResinous, GatherMinKg: 0.9, GatherMaxKg: 4.5, HeatFactor: 0.92, BurnFactor: 0.85, SparkEase: 4},
		{ID: "pine", Name: "Pine", BiomeTags: []string{"mountain", "boreal", "forest", "dry"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.1, HeatFactor: 0.9, BurnFactor: 0.82, SparkEase: 4},
		{ID: "fir", Name: "Fir", BiomeTags: []string{"forest", "mountain", "boreal"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.8, GatherMaxKg: 4.0, HeatFactor: 0.86, BurnFactor: 0.82, SparkEase: 3},
		{ID: "birch", Name: "Birch", BiomeTags: []string{"boreal", "forest", "lake", "subarctic"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.8, HeatFactor: 1.0, BurnFactor: 1.05, SparkEase: 3},
		{ID: "oak", Name: "Oak", BiomeTags: []string{"forest", "temperate", "coast", "river"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.9, GatherMaxKg: 5.0, HeatFactor: 1.15, BurnFactor: 1.18, SparkEase: 2},
		{ID: "maple", Name: "Maple", BiomeTags: []string{"forest", "mountain", "temperate"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.4, HeatFactor: 1.1, BurnFactor: 1.1, SparkEase: 2},
		{ID: "willow", Name: "Willow", BiomeTags: []string{"wetlands", "swamp", "delta", "river", "lake"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 3.0, HeatFactor: 0.74, BurnFactor: 0.7, SparkEase: 2},
		{ID: "mangrove", Name: "Mangrove", BiomeTags: []string{"delta", "coast", "swamp", "island"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 3.8, HeatFactor: 1.12, BurnFactor: 1.06, SparkEase: 2},
		{ID: "bamboo", Name: "Bamboo", BiomeTags: []string{"jungle", "tropical", "wetlands"}, WoodType: WoodTypeBamboo, GatherMinKg: 0.9, GatherMaxKg: 5.5, HeatFactor: 0.78, BurnFactor: 0.65, SparkEase: 4},
		{ID: "acacia", Name: "Acacia", BiomeTags: []string{"savanna", "badlands", "dry"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.5, HeatFactor: 1.08, BurnFactor: 1.02, SparkEase: 2},
		{ID: "baobab", Name: "Baobab", BiomeTags: []string{"savanna", "dry", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 1.0, GatherMaxKg: 5.8, HeatFactor: 1.18, BurnFactor: 1.16, SparkEase: 2},
		{ID: "palm", Name: "Palm", BiomeTags: []string{"island", "coast", "tropical", "delta"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 2.6, HeatFactor: 0.72, BurnFactor: 0.68, SparkEase: 3},
		{ID: "mesquite", Name: "Mesquite", BiomeTags: []string{"desert", "dry", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.4, HeatFactor: 1.2, BurnFactor: 1.2, SparkEase: 2},
		{ID: "driftwood", Name: "Driftwood", BiomeTags: []string{"coast", "delta", "island", "lake"}, WoodType: WoodTypeDriftwood, GatherMinKg: 0.4, GatherMaxKg: 2.4, HeatFactor: 0.62, BurnFactor: 0.54, SparkEase: 3},
	}
}

func TreesForBiome(biome string) []TreeSpec {
	norm := normalizeBiome(biome)
	catalog := TreeCatalog()
	out := make([]TreeSpec, 0, len(catalog))
	for _, tree := range catalog {
		for _, tag := range tree.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				out = append(out, tree)
				break
			}
		}
	}
	if len(out) == 0 {
		for _, tree := range catalog {
			if tree.WoodType == WoodTypeSoftwood || tree.WoodType == WoodTypeHardwood {
				out = append(out, tree)
			}
		}
	}
	return out
}

type WoodStock struct {
	Type WoodType `json:"type"`
	Kg   float64  `json:"kg"`
}

func (s *RunState) addWoodStock(woodType WoodType, kg float64) {
	if s == nil || kg <= 0 {
		return
	}
	for i := range s.WoodStock {
		if s.WoodStock[i].Type == woodType {
			s.WoodStock[i].Kg += kg
			return
		}
	}
	s.WoodStock = append(s.WoodStock, WoodStock{Type: woodType, Kg: kg})
}

func (s *RunState) consumeWoodStock(woodType WoodType, kg float64) bool {
	if s == nil || kg <= 0 {
		return false
	}
	for i := range s.WoodStock {
		if s.WoodStock[i].Type != woodType || s.WoodStock[i].Kg+1e-9 < kg {
			continue
		}
		s.WoodStock[i].Kg -= kg
		if s.WoodStock[i].Kg <= 0.001 {
			s.WoodStock = append(s.WoodStock[:i], s.WoodStock[i+1:]...)
		}
		return true
	}
	return false
}

func (s *RunState) totalWoodKg() float64 {
	if s == nil {
		return 0
	}
	total := 0.0
	for _, stock := range s.WoodStock {
		total += stock.Kg
	}
	return total
}

func (s *RunState) GatherWood(playerID int, requestedKg float64) (TreeSpec, float64, error) {
	if s == nil {
		return TreeSpec{}, 0, fmt.Errorf("run state is nil")
	}
	if _, ok := s.playerByID(playerID); !ok {
		return TreeSpec{}, 0, fmt.Errorf("player %d not found", playerID)
	}
	pool := TreesForBiome(s.Scenario.Biome)
	if len(pool) == 0 {
		return TreeSpec{}, 0, fmt.Errorf("no tree resources available")
	}

	rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("wood:%s:%d:%d", normalizeBiome(s.Scenario.Biome), s.Day, playerID)))
	tree := pool[rng.IntN(len(pool))]
	kg := requestedKg
	if kg <= 0 {
		kg = tree.GatherMinKg
		if tree.GatherMaxKg > tree.GatherMinKg {
			kg += rng.Float64() * (tree.GatherMaxKg - tree.GatherMinKg)
		}
	}
	if kg < 0.2 {
		kg = 0.2
	}
	s.addWoodStock(tree.WoodType, kg)
	return tree, kg, nil
}

type ShelterType string

const (
	ShelterLeanTo        ShelterType = "lean_to"
	ShelterDebrisHut     ShelterType = "debris_hut"
	ShelterTarpAFrame    ShelterType = "tarp_a_frame"
	ShelterSnowCave      ShelterType = "snow_cave"
	ShelterDesertShade   ShelterType = "desert_shade"
	ShelterSwampPlatform ShelterType = "swamp_platform"
	ShelterBambooHut     ShelterType = "bamboo_hut"
	ShelterRockOverhang  ShelterType = "rock_overhang"
)

type ShelterSpec struct {
	ID                 ShelterType
	Name               string
	BiomeTags          []string
	Insulation         int
	RainProtection     int
	WindProtection     int
	InsectProtection   int
	DurabilityPerDay   int
	BuildMoraleBonus   int
	BuildEnergyCost    int
	BuildHydrationCost int
}

func ShelterCatalog() []ShelterSpec {
	return []ShelterSpec{
		{ID: ShelterLeanTo, Name: "Lean-to", BiomeTags: []string{"forest", "mountain", "boreal", "coast"}, Insulation: 2, RainProtection: 2, WindProtection: 2, InsectProtection: 1, DurabilityPerDay: 6, BuildMoraleBonus: 2, BuildEnergyCost: 4, BuildHydrationCost: 2},
		{ID: ShelterDebrisHut, Name: "Debris Hut", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain"}, Insulation: 4, RainProtection: 3, WindProtection: 3, InsectProtection: 1, DurabilityPerDay: 5, BuildMoraleBonus: 3, BuildEnergyCost: 6, BuildHydrationCost: 3},
		{ID: ShelterTarpAFrame, Name: "Tarp A-Frame", BiomeTags: []string{"forest", "coast", "wetlands", "jungle", "mountain"}, Insulation: 3, RainProtection: 4, WindProtection: 3, InsectProtection: 2, DurabilityPerDay: 4, BuildMoraleBonus: 3, BuildEnergyCost: 3, BuildHydrationCost: 2},
		{ID: ShelterSnowCave, Name: "Snow Cave", BiomeTags: []string{"arctic", "subarctic", "tundra", "winter"}, Insulation: 6, RainProtection: 3, WindProtection: 6, InsectProtection: 3, DurabilityPerDay: 7, BuildMoraleBonus: 2, BuildEnergyCost: 8, BuildHydrationCost: 3},
		{ID: ShelterDesertShade, Name: "Desert Shade", BiomeTags: []string{"desert", "dry", "savanna", "badlands"}, Insulation: 1, RainProtection: 1, WindProtection: 2, InsectProtection: 2, DurabilityPerDay: 4, BuildMoraleBonus: 2, BuildEnergyCost: 3, BuildHydrationCost: 2},
		{ID: ShelterSwampPlatform, Name: "Swamp Platform", BiomeTags: []string{"swamp", "wetlands", "delta", "jungle"}, Insulation: 2, RainProtection: 2, WindProtection: 2, InsectProtection: 5, DurabilityPerDay: 5, BuildMoraleBonus: 3, BuildEnergyCost: 6, BuildHydrationCost: 3},
		{ID: ShelterBambooHut, Name: "Bamboo Hut", BiomeTags: []string{"jungle", "tropical", "wetlands", "island"}, Insulation: 3, RainProtection: 4, WindProtection: 3, InsectProtection: 4, DurabilityPerDay: 4, BuildMoraleBonus: 4, BuildEnergyCost: 6, BuildHydrationCost: 3},
		{ID: ShelterRockOverhang, Name: "Rock Overhang", BiomeTags: []string{"mountain", "badlands", "desert", "coast"}, Insulation: 3, RainProtection: 2, WindProtection: 4, InsectProtection: 1, DurabilityPerDay: 3, BuildMoraleBonus: 2, BuildEnergyCost: 2, BuildHydrationCost: 1},
	}
}

type ShelterState struct {
	Type       ShelterType `json:"type"`
	Durability int         `json:"durability"`
	BuiltDay   int         `json:"built_day"`
}

func SheltersForBiome(biome string) []ShelterSpec {
	norm := normalizeBiome(biome)
	catalog := ShelterCatalog()
	out := make([]ShelterSpec, 0, len(catalog))
	for _, shelter := range catalog {
		for _, tag := range shelter.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				out = append(out, shelter)
				break
			}
		}
	}
	if len(out) == 0 {
		out = append(out, catalog[0], catalog[1], catalog[2])
	}
	return out
}

func (s *RunState) BuildShelter(playerID int, shelterID string) (ShelterSpec, error) {
	if s == nil {
		return ShelterSpec{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return ShelterSpec{}, fmt.Errorf("player %d not found", playerID)
	}
	options := SheltersForBiome(s.Scenario.Biome)
	if len(options) == 0 {
		return ShelterSpec{}, fmt.Errorf("no shelters available")
	}

	id := ShelterType(strings.ToLower(strings.TrimSpace(shelterID)))
	if id == "" {
		id = options[0].ID
	}

	chosen := options[0]
	found := false
	for _, option := range options {
		if option.ID == id {
			chosen = option
			found = true
			break
		}
	}
	if !found {
		return ShelterSpec{}, fmt.Errorf("shelter not available in biome: %s", shelterID)
	}

	s.Shelter = ShelterState{
		Type:       chosen.ID,
		Durability: 100,
		BuiltDay:   s.Day,
	}
	player.Energy = clamp(player.Energy-chosen.BuildEnergyCost, 0, 100)
	player.Hydration = clamp(player.Hydration-chosen.BuildHydrationCost, 0, 100)
	player.Morale = clamp(player.Morale+chosen.BuildMoraleBonus, 0, 100)
	refreshEffectBars(player)

	return chosen, nil
}

func shelterByID(id ShelterType) (ShelterSpec, bool) {
	for _, shelter := range ShelterCatalog() {
		if shelter.ID == id {
			return shelter, true
		}
	}
	return ShelterSpec{}, false
}

type FireState struct {
	Lit           bool     `json:"lit"`
	WoodType      WoodType `json:"wood_type"`
	Intensity     int      `json:"intensity"`
	HeatC         int      `json:"heat_c"`
	FuelKg        float64  `json:"fuel_kg"`
	LastTendedDay int      `json:"last_tended_day"`
}

func (s *RunState) StartFire(playerID int, woodType WoodType, kg float64) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return fmt.Errorf("player %d not found", playerID)
	}
	if kg <= 0 {
		kg = 1.0
	}
	if !s.consumeWoodStock(woodType, kg) {
		return fmt.Errorf("not enough %s wood (need %.1fkg)", woodType, kg)
	}

	intensity, heat := fireMetrics(woodType, kg)
	s.Fire = FireState{
		Lit:           true,
		WoodType:      woodType,
		Intensity:     intensity,
		HeatC:         heat,
		FuelKg:        kg,
		LastTendedDay: s.Day,
	}
	player.Morale = clamp(player.Morale+3, 0, 100)
	player.Energy = clamp(player.Energy-1, 0, 100)
	refreshEffectBars(player)
	return nil
}

func (s *RunState) TendFire(playerID int, kg float64, woodType WoodType) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return fmt.Errorf("player %d not found", playerID)
	}
	if !s.Fire.Lit {
		return fmt.Errorf("no active fire")
	}
	if woodType == "" {
		woodType = s.Fire.WoodType
	}
	if kg <= 0 {
		kg = 0.8
	}
	if !s.consumeWoodStock(woodType, kg) {
		return fmt.Errorf("not enough %s wood (need %.1fkg)", woodType, kg)
	}

	s.Fire.FuelKg += kg
	intensityGain, heatGain := fireMetrics(woodType, kg)
	s.Fire.Intensity = clamp(s.Fire.Intensity+intensityGain/2, 10, 100)
	s.Fire.HeatC = clamp(s.Fire.HeatC+heatGain/3, 0, 120)
	s.Fire.WoodType = woodType
	s.Fire.LastTendedDay = s.Day
	player.Morale = clamp(player.Morale+2, 0, 100)
	refreshEffectBars(player)
	return nil
}

func (s *RunState) ExtinguishFire() {
	if s == nil {
		return
	}
	s.Fire = FireState{}
}

func fireMetrics(woodType WoodType, kg float64) (intensity int, heatC int) {
	if kg <= 0 {
		kg = 0.1
	}
	heatFactor := 1.0
	intensityFactor := 1.0
	switch woodType {
	case WoodTypeHardwood:
		heatFactor = 1.25
		intensityFactor = 1.1
	case WoodTypeResinous:
		heatFactor = 1.1
		intensityFactor = 1.2
	case WoodTypeBamboo:
		heatFactor = 0.85
		intensityFactor = 1.05
	case WoodTypeDriftwood:
		heatFactor = 0.7
		intensityFactor = 0.9
	case WoodTypeSoftwood:
		heatFactor = 0.95
		intensityFactor = 1.0
	}
	intensity = clamp(int(math.Round(22.0*kg*intensityFactor)), 8, 100)
	heatC = clamp(int(math.Round(18.0+40.0*kg*heatFactor)), 0, 120)
	return intensity, heatC
}

func (s *RunState) progressCampState() {
	if s == nil {
		return
	}

	if s.Shelter.Type != "" && s.Shelter.Durability > 0 {
		if shelter, ok := shelterByID(s.Shelter.Type); ok {
			loss := shelter.DurabilityPerDay
			if isRainyWeather(s.Weather.Type) {
				loss += 2
			}
			if isSevereWeather(s.Weather.Type) {
				loss += 3
			}
			s.Shelter.Durability = clamp(s.Shelter.Durability-loss, 0, 100)
		}
		if s.Shelter.Durability == 0 {
			s.Shelter = ShelterState{}
		}
	}

	if !s.Fire.Lit {
		return
	}
	burn := 0.7 + float64(s.Fire.Intensity)/120.0
	switch s.Fire.WoodType {
	case WoodTypeHardwood:
		burn *= 0.82
	case WoodTypeResinous:
		burn *= 1.05
	case WoodTypeBamboo:
		burn *= 1.08
	case WoodTypeDriftwood:
		burn *= 1.15
	}
	s.Fire.FuelKg -= burn
	if s.Fire.FuelKg <= 0.05 {
		s.ExtinguishFire()
		return
	}
	s.Fire.Intensity = clamp(int(float64(s.Fire.Intensity)*0.86), 8, 100)
	s.Fire.HeatC = clamp(int(float64(s.Fire.HeatC)*0.84), 0, 120)
}

func (s *RunState) campImpactForDay() statDelta {
	if s == nil {
		return statDelta{}
	}
	impact := statDelta{}

	if s.Shelter.Type != "" && s.Shelter.Durability > 0 {
		if shelter, ok := shelterByID(s.Shelter.Type); ok {
			// Shelter blunts weather stress and bug pressure.
			impact.Energy += shelter.Insulation / 2
			impact.Morale += shelter.RainProtection / 2
			if isRainyWeather(s.Weather.Type) {
				impact.Energy += shelter.RainProtection/2 + shelter.WindProtection/3
				impact.Morale += shelter.RainProtection / 2
			}
			if biomeIsTropicalWet(s.Scenario.Biome) {
				impact.Morale += shelter.InsectProtection / 2
			}
		}
	} else if isRainyWeather(s.Weather.Type) || isSevereWeather(s.Weather.Type) {
		impact.Energy -= 2
		impact.Morale -= 2
	}

	if s.Fire.Lit {
		if s.Weather.TemperatureC <= 4 {
			impact.Energy += clamp(s.Fire.HeatC/18, 1, 5)
			impact.Morale += clamp(s.Fire.Intensity/30, 1, 3)
		}
		if s.Weather.TemperatureC >= 32 {
			impact.Hydration -= clamp(s.Fire.HeatC/26, 1, 3)
		}
	} else if s.Weather.TemperatureC <= 0 {
		impact.Energy -= 2
		impact.Morale -= 1
	}

	return impact
}

type CraftableSpec struct {
	ID              string
	Name            string
	BiomeTags       []string
	Description     string
	MinBushcraft    int
	RequiresFire    bool
	RequiresShelter bool
	WoodKg          float64
	Effects         statDelta
}

func CraftableCatalog() []CraftableSpec {
	return []CraftableSpec{
		{ID: "tripod", Name: "Cooking Tripod", BiomeTags: []string{"forest", "coast", "mountain", "jungle"}, Description: "Stabilizes cooking over open flame.", MinBushcraft: 0, RequiresFire: true, WoodKg: 0.8, Effects: statDelta{Morale: 1}},
		{ID: "rain_catcher", Name: "Rain Catcher", BiomeTags: []string{"jungle", "wetlands", "coast", "island", "forest"}, Description: "Collects safer water during rain events.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 0.6, Effects: statDelta{Hydration: 2, Morale: 1}},
		{ID: "smoke_rack", Name: "Smoke Rack", BiomeTags: []string{"forest", "coast", "savanna", "jungle"}, Description: "Improves food preservation and morale.", MinBushcraft: 1, RequiresFire: true, WoodKg: 1.2, Effects: statDelta{Energy: 1, Morale: 2}},
		{ID: "fish_trap", Name: "Fish Trap", BiomeTags: []string{"delta", "river", "lake", "swamp", "coast"}, Description: "Passive fish capture near water channels.", MinBushcraft: 1, WoodKg: 0.7, Effects: statDelta{Energy: 1}},
		{ID: "windbreak", Name: "Windbreak Wall", BiomeTags: []string{"arctic", "subarctic", "mountain", "badlands", "coast"}, Description: "Reduces nightly wind chill exposure.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.0, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "charcoal_bed", Name: "Charcoal Bed", BiomeTags: []string{"desert", "dry", "savanna", "badlands"}, Description: "Retains heat after sunset in dry zones.", MinBushcraft: 2, RequiresFire: true, WoodKg: 1.4, Effects: statDelta{Energy: 2}},
		{ID: "bug_smudge", Name: "Bug Smudge Fire", BiomeTags: []string{"jungle", "swamp", "wetlands", "tropical", "island"}, Description: "Smoky fire that repels insects around camp.", MinBushcraft: 1, RequiresFire: true, WoodKg: 0.5, Effects: statDelta{Morale: 2}},
		{ID: "snow_melt_station", Name: "Snow Melt Station", BiomeTags: []string{"arctic", "subarctic", "tundra", "winter"}, Description: "Converts snow to usable water efficiently.", MinBushcraft: 1, RequiresFire: true, WoodKg: 0.9, Effects: statDelta{Hydration: 2}},
		{ID: "raised_bed", Name: "Raised Bed", BiomeTags: []string{"swamp", "wetlands", "jungle", "forest"}, Description: "Improves overnight rest by lifting off damp ground.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 1.1, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "signal_beacon", Name: "Signal Beacon", BiomeTags: []string{"coast", "island", "mountain", "badlands", "savanna"}, Description: "High-visibility signal structure.", MinBushcraft: 0, RequiresFire: true, WoodKg: 1.3, Effects: statDelta{Morale: 2}},
	}
}

func CraftablesForBiome(biome string) []CraftableSpec {
	norm := normalizeBiome(biome)
	catalog := CraftableCatalog()
	out := make([]CraftableSpec, 0, len(catalog))
	for _, craftable := range catalog {
		for _, tag := range craftable.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				out = append(out, craftable)
				break
			}
		}
	}
	if len(out) == 0 {
		out = append(out, catalog[0], catalog[1], catalog[2])
	}
	return out
}

func (s *RunState) CraftItem(playerID int, craftID string) (CraftableSpec, error) {
	if s == nil {
		return CraftableSpec{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return CraftableSpec{}, fmt.Errorf("player %d not found", playerID)
	}

	craftID = strings.ToLower(strings.TrimSpace(craftID))
	if craftID == "" {
		return CraftableSpec{}, fmt.Errorf("craft item id required")
	}

	options := CraftablesForBiome(s.Scenario.Biome)
	var chosen CraftableSpec
	found := false
	for _, craftable := range options {
		if craftable.ID == craftID {
			chosen = craftable
			found = true
			break
		}
	}
	if !found {
		return CraftableSpec{}, fmt.Errorf("craft item not available in biome: %s", craftID)
	}

	if player.Bushcraft < chosen.MinBushcraft {
		return CraftableSpec{}, fmt.Errorf("requires bushcraft %+d", chosen.MinBushcraft)
	}
	if chosen.RequiresFire && !s.Fire.Lit {
		return CraftableSpec{}, fmt.Errorf("requires active fire")
	}
	if chosen.RequiresShelter && (s.Shelter.Type == "" || s.Shelter.Durability <= 0) {
		return CraftableSpec{}, fmt.Errorf("requires active shelter")
	}
	if chosen.WoodKg > 0 {
		woodType := s.Fire.WoodType
		if woodType == "" && len(s.WoodStock) > 0 {
			woodType = s.WoodStock[0].Type
		}
		if woodType == "" {
			woodType = WoodTypeHardwood
		}
		if !s.consumeWoodStock(woodType, chosen.WoodKg) {
			return CraftableSpec{}, fmt.Errorf("needs %.1fkg wood", chosen.WoodKg)
		}
	}

	if !slices.Contains(s.CraftedItems, chosen.ID) {
		s.CraftedItems = append(s.CraftedItems, chosen.ID)
	}
	player.Energy = clamp(player.Energy+chosen.Effects.Energy-1, 0, 100)
	player.Hydration = clamp(player.Hydration+chosen.Effects.Hydration, 0, 100)
	player.Morale = clamp(player.Morale+chosen.Effects.Morale+1, 0, 100)
	refreshEffectBars(player)
	return chosen, nil
}

func parseWoodType(raw string) WoodType {
	n := strings.ToLower(strings.TrimSpace(raw))
	switch n {
	case "hardwood", "hard":
		return WoodTypeHardwood
	case "softwood", "soft":
		return WoodTypeSoftwood
	case "resin", "resinous", "pine":
		return WoodTypeResinous
	case "bamboo":
		return WoodTypeBamboo
	case "driftwood", "drift":
		return WoodTypeDriftwood
	default:
		return ""
	}
}

func formatWoodStock(stock []WoodStock) string {
	if len(stock) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(stock))
	for _, item := range stock {
		parts = append(parts, fmt.Sprintf("%s %.1fkg", item.Type, item.Kg))
	}
	return strings.Join(parts, ", ")
}

func parseOptionalPlayerAndNumber(tokens []string) (playerID int, value float64, valueSet bool, rest []string) {
	playerID = 1
	rest = make([]string, 0, len(tokens))
	for _, token := range tokens {
		if parsed := parsePlayerToken(token); parsed > 0 {
			playerID = parsed
			continue
		}
		if !valueSet {
			if n, err := strconv.ParseFloat(token, 64); err == nil {
				value = n
				valueSet = true
				continue
			}
		}
		rest = append(rest, token)
	}
	return playerID, value, valueSet, rest
}
