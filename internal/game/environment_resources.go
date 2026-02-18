package game

import (
	"fmt"
	"hash/fnv"
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

type ResourceSpec struct {
	ID        string
	Name      string
	BiomeTags []string
	Unit      string
	GatherMin float64
	GatherMax float64
	Dryness   float64 // 0=soaked, 1=bone dry
	Flammable bool
}

type ResourceStock struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Unit string  `json:"unit"`
	Qty  float64 `json:"qty"`
}

func ResourceCatalog() []ResourceSpec {
	return []ResourceSpec{
		{ID: "dry_grass", Name: "Dry Grass", BiomeTags: []string{"savanna", "badlands", "dry", "desert", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.85, Flammable: true},
		{ID: "birch_bark", Name: "Birch Bark", BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, Unit: "sheet", GatherMin: 1, GatherMax: 3, Dryness: 0.8, Flammable: true},
		{ID: "cedar_bark", Name: "Cedar Bark", BiomeTags: []string{"coast", "temperate_rainforest", "vancouver", "forest"}, Unit: "sheet", GatherMin: 1, GatherMax: 3, Dryness: 0.78, Flammable: true},
		{ID: "inner_bark_fiber", Name: "Inner Bark Fiber", BiomeTags: []string{"forest", "boreal", "savanna", "jungle"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.72, Flammable: true},
		{ID: "fatwood", Name: "Fatwood", BiomeTags: []string{"boreal", "forest", "mountain", "dry"}, Unit: "stick", GatherMin: 1, GatherMax: 3, Dryness: 0.9, Flammable: true},
		{ID: "resin", Name: "Tree Resin", BiomeTags: []string{"forest", "boreal", "mountain", "jungle", "savanna"}, Unit: "lump", GatherMin: 1, GatherMax: 3, Dryness: 0.95, Flammable: true},
		{ID: "punkwood", Name: "Punkwood", BiomeTags: []string{"forest", "boreal", "swamp", "wetlands"}, Unit: "chunk", GatherMin: 1, GatherMax: 3, Dryness: 0.55, Flammable: true},
		{ID: "cattail_fluff", Name: "Cattail Fluff", BiomeTags: []string{"wetlands", "swamp", "delta", "lake"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.82, Flammable: true},
		{ID: "dry_moss", Name: "Dry Moss", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain", "coast"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.76, Flammable: true},
		{ID: "coconut_husk", Name: "Coconut Husk", BiomeTags: []string{"island", "coast", "tropical", "delta"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.74, Flammable: true},
		{ID: "reed_bundle", Name: "Reed Bundle", BiomeTags: []string{"wetlands", "swamp", "delta", "lake", "river"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.58, Flammable: true},
		{ID: "vine_fiber", Name: "Vine Fiber", BiomeTags: []string{"jungle", "tropical", "wetlands", "swamp"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.52, Flammable: true},
		{ID: "stone_flake", Name: "Stone Flake", BiomeTags: []string{"mountain", "badlands", "desert", "river", "coast"}, Unit: "piece", GatherMin: 1, GatherMax: 4, Dryness: 1.0, Flammable: false},
		{ID: "clay", Name: "Clay", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "badlands", "coast"}, Unit: "kg", GatherMin: 0.4, GatherMax: 2.0, Dryness: 0.2, Flammable: false},
		{ID: "charcoal", Name: "Charcoal", BiomeTags: []string{"forest", "savanna", "jungle", "mountain", "coast"}, Unit: "chunk", GatherMin: 1, GatherMax: 3, Dryness: 0.95, Flammable: true},
		{ID: "drift_reed_fiber", Name: "Drift Reed Fiber", BiomeTags: []string{"delta", "coast", "island", "wetlands"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.5, Flammable: true},
	}
}

func ResourcesForBiome(biome string) []ResourceSpec {
	norm := normalizeBiome(biome)
	catalog := ResourceCatalog()
	out := make([]ResourceSpec, 0, len(catalog))
	for _, resource := range catalog {
		for _, tag := range resource.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				out = append(out, resource)
				break
			}
		}
	}
	if len(out) == 0 {
		// Always leave a fallback starter set.
		for _, resource := range catalog {
			if resource.ID == "dry_grass" || resource.ID == "inner_bark_fiber" || resource.ID == "stone_flake" || resource.ID == "clay" {
				out = append(out, resource)
			}
		}
	}
	return out
}

func (s *RunState) addResourceStock(resource ResourceSpec, qty float64) {
	if s == nil || qty <= 0 {
		return
	}
	for i := range s.ResourceStock {
		if s.ResourceStock[i].ID == resource.ID {
			s.ResourceStock[i].Qty += qty
			return
		}
	}
	s.ResourceStock = append(s.ResourceStock, ResourceStock{
		ID:   resource.ID,
		Name: resource.Name,
		Unit: resource.Unit,
		Qty:  qty,
	})
}

func (s *RunState) consumeResourceStock(id string, qty float64) bool {
	if s == nil || qty <= 0 {
		return false
	}
	for i := range s.ResourceStock {
		if s.ResourceStock[i].ID != id || s.ResourceStock[i].Qty+1e-9 < qty {
			continue
		}
		s.ResourceStock[i].Qty -= qty
		if s.ResourceStock[i].Qty <= 0.001 {
			s.ResourceStock = append(s.ResourceStock[:i], s.ResourceStock[i+1:]...)
		}
		return true
	}
	return false
}

func (s *RunState) resourceQty(id string) float64 {
	if s == nil {
		return 0
	}
	for _, stock := range s.ResourceStock {
		if stock.ID == id {
			return stock.Qty
		}
	}
	return 0
}

func (s *RunState) findResourceForBiome(resourceID string) (ResourceSpec, bool) {
	options := ResourcesForBiome(s.Scenario.Biome)
	for _, resource := range options {
		if resource.ID == resourceID {
			return resource, true
		}
	}
	for _, resource := range ResourceCatalog() {
		if resource.ID == resourceID {
			return resource, true
		}
	}
	return ResourceSpec{}, false
}

func (s *RunState) CollectResource(playerID int, resourceID string, requestedQty float64) (ResourceSpec, float64, error) {
	if s == nil {
		return ResourceSpec{}, 0, fmt.Errorf("run state is nil")
	}
	if _, ok := s.playerByID(playerID); !ok {
		return ResourceSpec{}, 0, fmt.Errorf("player %d not found", playerID)
	}

	resourceID = strings.ToLower(strings.TrimSpace(resourceID))
	available := ResourcesForBiome(s.Scenario.Biome)
	if len(available) == 0 {
		return ResourceSpec{}, 0, fmt.Errorf("no resources available")
	}

	var resource ResourceSpec
	found := false
	if resourceID != "" && resourceID != "any" {
		for _, entry := range available {
			if entry.ID == resourceID {
				resource = entry
				found = true
				break
			}
		}
		if !found {
			return ResourceSpec{}, 0, fmt.Errorf("resource not available in biome: %s", resourceID)
		}
	} else {
		rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("collect:%s:%d:%d", normalizeBiome(s.Scenario.Biome), s.Day, playerID)))
		resource = available[rng.IntN(len(available))]
		found = true
	}
	if !found {
		return ResourceSpec{}, 0, fmt.Errorf("resource selection failed")
	}

	qty := requestedQty
	if qty <= 0 {
		rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("collect:%s:%d:%d:%s", normalizeBiome(s.Scenario.Biome), s.Day, playerID, resource.ID)))
		qty = resource.GatherMin
		if resource.GatherMax > resource.GatherMin {
			qty += rng.Float64() * (resource.GatherMax - resource.GatherMin)
		}
	}
	min := 0.1
	if resource.Unit != "kg" {
		min = 1
		qty = math.Round(qty)
	}
	if qty < min {
		qty = min
	}

	s.addResourceStock(resource, qty)
	return resource, qty, nil
}

func formatResourceStock(stock []ResourceStock) string {
	if len(stock) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(stock))
	for _, item := range stock {
		if item.Unit == "kg" {
			parts = append(parts, fmt.Sprintf("%s %.1f%s", item.ID, item.Qty, item.Unit))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %.0f%s", item.ID, item.Qty, item.Unit))
	}
	return strings.Join(parts, ", ")
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
	Type    WoodType `json:"type"`
	Kg      float64  `json:"kg"`
	Wetness float64  `json:"wetness"` // 0 = fully dry, 1 = soaked
}

func (s *RunState) addWoodStock(woodType WoodType, kg float64) {
	s.addWoodStockWithWetness(woodType, kg, 0.25)
}

func (s *RunState) addWoodStockWithWetness(woodType WoodType, kg float64, wetness float64) {
	if s == nil || kg <= 0 {
		return
	}
	wetness = clampFloat(wetness, 0, 1)
	for i := range s.WoodStock {
		if s.WoodStock[i].Type == woodType {
			total := s.WoodStock[i].Kg + kg
			if total > 0 {
				s.WoodStock[i].Wetness = ((s.WoodStock[i].Wetness * s.WoodStock[i].Kg) + (wetness * kg)) / total
			}
			s.WoodStock[i].Kg += kg
			return
		}
	}
	s.WoodStock = append(s.WoodStock, WoodStock{Type: woodType, Kg: kg, Wetness: wetness})
}

func (s *RunState) consumeWoodStock(woodType WoodType, kg float64) bool {
	_, ok := s.consumeWoodStockDetailed(woodType, kg)
	return ok
}

func (s *RunState) consumeWoodStockDetailed(woodType WoodType, kg float64) (float64, bool) {
	if s == nil || kg <= 0 {
		return 0, false
	}
	for i := range s.WoodStock {
		if s.WoodStock[i].Type != woodType || s.WoodStock[i].Kg+1e-9 < kg {
			continue
		}
		wetness := s.WoodStock[i].Wetness
		s.WoodStock[i].Kg -= kg
		if s.WoodStock[i].Kg <= 0.001 {
			s.WoodStock = append(s.WoodStock[:i], s.WoodStock[i+1:]...)
		}
		return wetness, true
	}
	return 0, false
}

func (s *RunState) consumeAnyWoodPreferDry(kg float64) (WoodType, float64, bool) {
	if s == nil || kg <= 0 {
		return "", 0, false
	}
	bestIndex := -1
	bestWetness := 2.0
	for i := range s.WoodStock {
		if s.WoodStock[i].Kg+1e-9 < kg {
			continue
		}
		if s.WoodStock[i].Wetness < bestWetness {
			bestWetness = s.WoodStock[i].Wetness
			bestIndex = i
		}
	}
	if bestIndex < 0 {
		return "", 0, false
	}
	chosen := s.WoodStock[bestIndex]
	s.WoodStock[bestIndex].Kg -= kg
	if s.WoodStock[bestIndex].Kg <= 0.001 {
		s.WoodStock = append(s.WoodStock[:bestIndex], s.WoodStock[bestIndex+1:]...)
	}
	return chosen.Type, chosen.Wetness, true
}

func (s *RunState) averageWoodWetness(woodType WoodType) float64 {
	if s == nil {
		return 0.5
	}
	totalKg := 0.0
	wetnessWeighted := 0.0
	for _, stock := range s.WoodStock {
		if woodType != "" && stock.Type != woodType {
			continue
		}
		totalKg += stock.Kg
		wetnessWeighted += stock.Wetness * stock.Kg
	}
	if totalKg <= 0 {
		return 0.55
	}
	return clampFloat(wetnessWeighted/totalKg, 0, 1)
}

func (s *RunState) DryWood(playerID int, kg float64) (float64, error) {
	if s == nil {
		return 0, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return 0, fmt.Errorf("player %d not found", playerID)
	}
	if kg <= 0 {
		kg = 1.0
	}
	remaining := kg
	dried := 0.0
	for i := range s.WoodStock {
		if remaining <= 0 {
			break
		}
		if s.WoodStock[i].Kg <= 0 {
			continue
		}
		take := math.Min(s.WoodStock[i].Kg, remaining)
		// Drying is partial per operation.
		drop := 0.28 + (float64(clamp(player.Bushcraft, -3, 3)) * 0.03)
		s.WoodStock[i].Wetness = clampFloat(s.WoodStock[i].Wetness-drop, 0, 1)
		dried += take
		remaining -= take
	}
	if dried <= 0 {
		return 0, fmt.Errorf("no wood stock to dry")
	}
	player.Energy = clamp(player.Energy-clamp(int(math.Ceil(dried)), 1, 5), 0, 100)
	player.Hydration = clamp(player.Hydration-clamp(int(math.Ceil(dried/2)), 1, 3), 0, 100)
	refreshEffectBars(player)
	return dried, nil
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
	s.addWoodStockWithWetness(tree.WoodType, kg, s.ambientWoodWetness())
	return tree, kg, nil
}

func (s *RunState) ambientWoodWetness() float64 {
	if s == nil {
		return 0.4
	}
	wetness := 0.22
	switch s.Weather.Type {
	case WeatherRain:
		wetness += 0.28
	case WeatherHeavyRain:
		wetness += 0.4
	case WeatherStorm:
		wetness += 0.48
	case WeatherSnow:
		wetness += 0.25
	case WeatherBlizzard:
		wetness += 0.35
	case WeatherSunny:
		wetness -= 0.08
	case WeatherHeatwave:
		wetness -= 0.14
	}
	if biomeIsTropicalWet(s.Scenario.Biome) {
		wetness += 0.12
	}
	if biomeIsDesertOrDry(s.Scenario.Biome) {
		wetness -= 0.1
	}
	if s.Weather.StreakDays > 1 && isRainyWeather(s.Weather.Type) {
		wetness += float64(clamp(s.Weather.StreakDays-1, 0, 4)) * 0.06
	}
	return clampFloat(wetness, 0.02, 0.98)
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
	LastMethod    string   `json:"last_method,omitempty"`
}

type FireMethod string

const (
	FireMethodFerro     FireMethod = "ferro"
	FireMethodBowDrill  FireMethod = "bow_drill"
	FireMethodHandDrill FireMethod = "hand_drill"
)

type FirePrepState struct {
	TinderBundles   int     `json:"tinder_bundles"`
	KindlingBundles int     `json:"kindling_bundles"`
	FeatherSticks   int     `json:"feather_sticks"`
	Embers          int     `json:"embers"`
	TinderQuality   float64 `json:"tinder_quality"`
	KindlingQuality float64 `json:"kindling_quality"`
	FeatherQuality  float64 `json:"feather_quality"`
}

func FireMethods() []FireMethod {
	return []FireMethod{FireMethodFerro, FireMethodBowDrill, FireMethodHandDrill}
}

func ParseFireMethod(raw string) FireMethod {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "ferro", "ferrorod", "firesteel":
		return FireMethodFerro
	case "bow", "bowdrill", "bow_drill":
		return FireMethodBowDrill
	case "hand", "handdrill", "hand_drill":
		return FireMethodHandDrill
	default:
		return ""
	}
}

func (s *RunState) StartFire(playerID int, woodType WoodType, kg float64) error {
	return s.startFireWithMethod(playerID, woodType, kg, FireMethodFerro)
}

func (s *RunState) startFireWithMethod(playerID int, woodType WoodType, kg float64, method FireMethod) error {
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
	wetness, ok := s.consumeWoodStockDetailed(woodType, kg)
	if !ok {
		return fmt.Errorf("not enough %s wood (need %.1fkg)", woodType, kg)
	}

	intensity, heat := fireMetrics(woodType, kg)
	if wetness > 0.6 {
		intensity = clamp(intensity-int(math.Round(wetness*8)), 6, 100)
		heat = clamp(heat-int(math.Round(wetness*12)), 0, 120)
	}
	s.Fire = FireState{
		Lit:           true,
		WoodType:      woodType,
		Intensity:     intensity,
		HeatC:         heat,
		FuelKg:        kg,
		LastTendedDay: s.Day,
		LastMethod:    string(method),
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
	wetness, ok := s.consumeWoodStockDetailed(woodType, kg)
	if !ok {
		return fmt.Errorf("not enough %s wood (need %.1fkg)", woodType, kg)
	}

	s.Fire.FuelKg += kg
	intensityGain, heatGain := fireMetrics(woodType, kg)
	if wetness > 0.6 {
		intensityGain = clamp(intensityGain-int(math.Round(wetness*6)), 4, 100)
		heatGain = clamp(heatGain-int(math.Round(wetness*10)), 0, 120)
	}
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

func (s *RunState) PrepareFireMaterial(playerID int, material string, count int) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return 0, fmt.Errorf("player %d not found", playerID)
	}
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	switch strings.ToLower(strings.TrimSpace(material)) {
	case "tinder":
		created := 0
		quality := 0.0
		for i := 0; i < count; i++ {
			resourceID, dryness, ok := s.consumeBestTinderResource()
			if !ok {
				break
			}
			created++
			quality += dryness
			if resourceID == "resin" || resourceID == "fatwood" {
				quality += 0.1
			}
		}
		if created == 0 {
			return 0, fmt.Errorf("need tinder resources (collect dry_grass, birch_bark, cedar_bark, cattail_fluff, dry_moss, coconut_husk, inner_bark_fiber, resin, or fatwood)")
		}
		s.FirePrep.TinderBundles += created
		s.FirePrep.TinderQuality = blendQuality(s.FirePrep.TinderQuality, s.FirePrep.TinderBundles-created, quality/float64(created), created)
		player.Energy = clamp(player.Energy-created, 0, 100)
		refreshEffectBars(player)
		return created, nil
	case "kindling":
		created := 0
		quality := 0.0
		for i := 0; i < count; i++ {
			_, wetness, ok := s.consumeAnyWoodPreferDry(0.22)
			if !ok {
				break
			}
			created++
			quality += 1.0 - wetness
		}
		if created == 0 {
			return 0, fmt.Errorf("need wood stock (use: wood gather)")
		}
		s.FirePrep.KindlingBundles += created
		s.FirePrep.KindlingQuality = blendQuality(s.FirePrep.KindlingQuality, s.FirePrep.KindlingBundles-created, quality/float64(created), created)
		player.Energy = clamp(player.Energy-created, 0, 100)
		player.Hydration = clamp(player.Hydration-created/2, 0, 100)
		refreshEffectBars(player)
		return created, nil
	case "feather", "feathersticks", "feather_sticks":
		created := 0
		quality := 0.0
		for i := 0; i < count; i++ {
			_, wetness, ok := s.consumeAnyWoodPreferDry(0.18)
			if !ok {
				break
			}
			created++
			quality += 1.0 - wetness
		}
		if created == 0 {
			return 0, fmt.Errorf("need dry-ish wood stock for feather sticks")
		}
		s.FirePrep.FeatherSticks += created
		s.FirePrep.FeatherQuality = blendQuality(s.FirePrep.FeatherQuality, s.FirePrep.FeatherSticks-created, quality/float64(created), created)
		player.Energy = clamp(player.Energy-created, 0, 100)
		refreshEffectBars(player)
		return created, nil
	default:
		return 0, fmt.Errorf("unknown material: %s (use tinder|kindling|feather)", material)
	}
}

func blendQuality(previous float64, previousCount int, next float64, nextCount int) float64 {
	total := previousCount + nextCount
	if total <= 0 {
		return 0
	}
	return clampFloat(((previous*float64(previousCount))+(next*float64(nextCount)))/float64(total), 0, 1)
}

func (s *RunState) consumeBestTinderResource() (string, float64, bool) {
	if s == nil || len(s.ResourceStock) == 0 {
		return "", 0, false
	}
	bestIndex := -1
	bestDryness := -1.0
	for i := range s.ResourceStock {
		spec, ok := s.findResourceForBiome(s.ResourceStock[i].ID)
		if !ok || !spec.Flammable {
			continue
		}
		if s.ResourceStock[i].Qty < 1 {
			continue
		}
		if spec.Dryness > bestDryness {
			bestDryness = spec.Dryness
			bestIndex = i
		}
	}
	if bestIndex < 0 {
		return "", 0, false
	}
	id := s.ResourceStock[bestIndex].ID
	s.ResourceStock[bestIndex].Qty -= 1
	if s.ResourceStock[bestIndex].Qty <= 0.001 {
		s.ResourceStock = append(s.ResourceStock[:bestIndex], s.ResourceStock[bestIndex+1:]...)
	}
	return id, clampFloat(bestDryness, 0, 1), true
}

func (s *RunState) TryCreateEmber(playerID int, method FireMethod, preferredWood WoodType) (float64, bool, error) {
	if s == nil {
		return 0, false, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return 0, false, fmt.Errorf("player %d not found", playerID)
	}
	if method != FireMethodBowDrill && method != FireMethodHandDrill {
		return 0, false, fmt.Errorf("ember method must be bow_drill or hand_drill")
	}
	if err := s.ensureFireMethodComponents(method); err != nil {
		return 0, false, err
	}

	chance := s.emberChance(method, *player, preferredWood)
	success := s.fireRoll("ember", method, playerID) <= chance
	s.FireAttemptCount++

	// Attempt consumes some prep regardless of result.
	if s.FirePrep.TinderBundles > 0 {
		s.FirePrep.TinderBundles--
	}
	if s.FirePrep.KindlingBundles > 0 {
		s.FirePrep.KindlingBundles--
	}
	if method == FireMethodBowDrill && s.FirePrep.FeatherSticks > 0 {
		s.FirePrep.FeatherSticks--
	}
	player.Energy = clamp(player.Energy-2, 0, 100)
	player.Hydration = clamp(player.Hydration-1, 0, 100)
	if success {
		s.FirePrep.Embers++
		player.Morale = clamp(player.Morale+2, 0, 100)
	} else {
		player.Morale = clamp(player.Morale-1, 0, 100)
	}
	refreshEffectBars(player)

	return chance, success, nil
}

func (s *RunState) ensureFireMethodComponents(method FireMethod) error {
	has := func(id string) bool { return slices.Contains(s.CraftedItems, id) }
	switch method {
	case FireMethodBowDrill:
		required := []string{"bow_drill_spindle", "bow_drill_hearth_board", "bow_drill_bow", "bearing_block"}
		for _, item := range required {
			if !has(item) {
				return fmt.Errorf("missing component: %s (craft it first)", item)
			}
		}
	case FireMethodHandDrill:
		required := []string{"hand_drill_spindle", "hand_drill_hearth_board"}
		for _, item := range required {
			if !has(item) {
				return fmt.Errorf("missing component: %s (craft it first)", item)
			}
		}
	default:
		return fmt.Errorf("unsupported fire method: %s", method)
	}
	return nil
}

func (s *RunState) emberChance(method FireMethod, player PlayerState, preferredWood WoodType) float64 {
	base := 0.34
	switch method {
	case FireMethodBowDrill:
		base = 0.46
	case FireMethodHandDrill:
		base = 0.31
	}

	chance := base
	chance += float64(clamp(player.Bushcraft, -3, 3)) * 0.07
	chance += float64(clamp(player.Endurance, -3, 3)) * 0.03
	chance += float64(clamp(player.Mental, -3, 3)) * 0.02

	if s.FirePrep.TinderBundles > 0 {
		chance += 0.08 + (s.FirePrep.TinderQuality * 0.08)
	} else {
		chance -= 0.1
	}
	if s.FirePrep.KindlingBundles > 0 {
		chance += 0.05 + (s.FirePrep.KindlingQuality * 0.08)
	} else {
		chance -= 0.08
	}
	if s.FirePrep.FeatherSticks > 0 {
		chance += 0.04 + (s.FirePrep.FeatherQuality * 0.06)
	}

	wetness := s.averageWoodWetness(preferredWood)
	chance -= wetness * 0.38

	switch s.Weather.Type {
	case WeatherRain:
		chance -= 0.12
	case WeatherHeavyRain:
		chance -= 0.2
	case WeatherStorm:
		chance -= 0.28
	case WeatherSnow:
		chance -= 0.14
	case WeatherBlizzard:
		chance -= 0.24
	case WeatherSunny:
		chance += 0.04
	case WeatherHeatwave:
		chance += 0.06
	}

	if method == FireMethodHandDrill && s.Weather.TemperatureC <= 2 {
		chance -= 0.08
	}
	if method == FireMethodBowDrill && s.Weather.TemperatureC <= 2 {
		chance -= 0.04
	}

	if preferredWood == WoodTypeResinous {
		chance += 0.05
	}
	if preferredWood == WoodTypeDriftwood {
		chance -= 0.06
	}

	return clampFloat(chance, 0.03, 0.95)
}

func (s *RunState) IgniteFromEmber(playerID int, woodType WoodType, kg float64) (float64, bool, error) {
	if s == nil {
		return 0, false, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return 0, false, fmt.Errorf("player %d not found", playerID)
	}
	if s.FirePrep.Embers < 1 {
		return 0, false, fmt.Errorf("no ember ready (use: fire ember bow|hand)")
	}
	if s.FirePrep.TinderBundles < 1 || s.FirePrep.KindlingBundles < 1 {
		return 0, false, fmt.Errorf("need at least 1 tinder bundle and 1 kindling bundle (use: fire prep ...)")
	}
	if kg <= 0 {
		kg = 1.0
	}
	if woodType == "" {
		if len(s.WoodStock) > 0 {
			woodType = s.WoodStock[0].Type
		} else {
			woodType = WoodTypeHardwood
		}
	}

	chance := 0.62
	chance += s.FirePrep.TinderQuality * 0.12
	chance += s.FirePrep.KindlingQuality * 0.12
	chance += s.FirePrep.FeatherQuality * 0.08
	if s.FirePrep.FeatherSticks > 0 {
		chance += 0.06
	}
	if s.resourceQty("resin") >= 1 {
		chance += 0.05
		_ = s.consumeResourceStock("resin", 1)
	}
	wetness := s.averageWoodWetness(woodType)
	chance -= wetness * 0.42
	switch s.Weather.Type {
	case WeatherRain:
		chance -= 0.09
	case WeatherHeavyRain:
		chance -= 0.16
	case WeatherStorm:
		chance -= 0.24
	case WeatherSnow:
		chance -= 0.1
	case WeatherBlizzard:
		chance -= 0.18
	case WeatherSunny:
		chance += 0.04
	}
	chance += float64(clamp(player.Bushcraft, -3, 3)) * 0.04
	chance = clampFloat(chance, 0.05, 0.95)

	success := s.fireRoll("ignite", FireMethodBowDrill, playerID) <= chance
	s.FireAttemptCount++

	// Consume prep for every ignition attempt.
	s.FirePrep.Embers = maxInt(0, s.FirePrep.Embers-1)
	s.FirePrep.TinderBundles = maxInt(0, s.FirePrep.TinderBundles-1)
	s.FirePrep.KindlingBundles = maxInt(0, s.FirePrep.KindlingBundles-1)
	if s.FirePrep.FeatherSticks > 0 {
		s.FirePrep.FeatherSticks--
	}

	player.Energy = clamp(player.Energy-1, 0, 100)
	player.Hydration = clamp(player.Hydration-1, 0, 100)

	if !success {
		player.Morale = clamp(player.Morale-1, 0, 100)
		refreshEffectBars(player)
		return chance, false, nil
	}

	if err := s.startFireWithMethod(playerID, woodType, kg, FireMethodBowDrill); err != nil {
		return chance, false, err
	}
	refreshEffectBars(player)
	return chance, true, nil
}

func (s *RunState) fireRoll(stage string, method FireMethod, playerID int) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d:%d:%s:%s:%d", s.Config.Seed, s.Day, playerID, stage, method, s.FireAttemptCount)))
	return float64(h.Sum64()%10000) / 10000.0
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

	// Prepared fire materials can degrade in persistent wet weather.
	if isRainyWeather(s.Weather.Type) || s.Weather.Type == WeatherSnow || s.Weather.Type == WeatherBlizzard {
		wetHit := 0.08
		if s.Shelter.Type != "" && s.Shelter.Durability > 0 {
			wetHit *= 0.45
		}
		s.FirePrep.TinderQuality = clampFloat(s.FirePrep.TinderQuality-wetHit, 0, 1)
		s.FirePrep.KindlingQuality = clampFloat(s.FirePrep.KindlingQuality-wetHit*0.8, 0, 1)
		s.FirePrep.FeatherQuality = clampFloat(s.FirePrep.FeatherQuality-wetHit*0.7, 0, 1)
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
	ID                string
	Name              string
	BiomeTags         []string
	Description       string
	MinBushcraft      int
	RequiresFire      bool
	RequiresShelter   bool
	WoodKg            float64
	RequiresItems     []string
	RequiresResources []ResourceRequirement
	Effects           statDelta
}

type ResourceRequirement struct {
	ID  string
	Qty float64
}

func CraftableCatalog() []CraftableSpec {
	return []CraftableSpec{
		// Fire method components.
		{ID: "bow_drill_spindle", Name: "Bow Drill Spindle", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Straight spindle for bow drill fire set.", MinBushcraft: 1, WoodKg: 0.2, Effects: statDelta{Morale: 1}},
		{ID: "bow_drill_hearth_board", Name: "Bow Drill Hearth Board", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Hearth board for receiving ember dust.", MinBushcraft: 1, WoodKg: 0.25, Effects: statDelta{Morale: 1}},
		{ID: "bow_drill_bow", Name: "Bow Drill Bow", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Flexible bow used to spin spindle.", MinBushcraft: 1, WoodKg: 0.3, RequiresResources: []ResourceRequirement{{ID: "vine_fiber", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "bearing_block", Name: "Bearing Block", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Handhold block reducing friction loss.", MinBushcraft: 1, WoodKg: 0.18, Effects: statDelta{Energy: 1}},
		{ID: "hand_drill_spindle", Name: "Hand Drill Spindle", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Long spindle for hand drill method.", MinBushcraft: 1, WoodKg: 0.2, Effects: statDelta{Morale: 1}},
		{ID: "hand_drill_hearth_board", Name: "Hand Drill Hearth Board", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands", "desert"}, Description: "Hearth board for hand drill ember slot.", MinBushcraft: 1, WoodKg: 0.22, Effects: statDelta{Morale: 1}},

		// Tree and wood products.
		{ID: "tarp_stakes", Name: "Tarp Stakes", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "savanna", "badlands"}, Description: "Sharpened stakes for faster shelter setup.", MinBushcraft: 0, WoodKg: 0.25, Effects: statDelta{Energy: 1}},
		{ID: "ridge_pole_kit", Name: "Ridge Pole Kit", BiomeTags: []string{"forest", "mountain", "boreal", "coast", "jungle"}, Description: "Pre-cut poles for lean-to and tarp frames.", MinBushcraft: 1, WoodKg: 1.1, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "shelter_lattice", Name: "Shelter Lattice", BiomeTags: []string{"forest", "jungle", "wetlands", "swamp", "coast"}, Description: "Interlaced branchwork to improve walling.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.2, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "fish_spear_shaft", Name: "Fish Spear Shaft", BiomeTags: []string{"delta", "river", "lake", "coast", "wetlands", "jungle"}, Description: "Balanced shaft for spear fishing builds.", MinBushcraft: 1, WoodKg: 0.4, Effects: statDelta{Energy: 1}},
		{ID: "fish_gorge_hooks", Name: "Fish Gorge Hooks", BiomeTags: []string{"river", "lake", "delta", "coast", "wetlands"}, Description: "Carved gorge hooks for passive fish lines.", MinBushcraft: 2, WoodKg: 0.18, RequiresResources: []ResourceRequirement{{ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "trap_trigger_set", Name: "Trap Trigger Set", BiomeTags: []string{"forest", "boreal", "mountain", "savanna", "badlands"}, Description: "Notched trigger kit for snare systems.", MinBushcraft: 2, WoodKg: 0.45, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "carving_wedges", Name: "Carving Wedges", BiomeTags: []string{"forest", "mountain", "coast", "boreal", "savanna"}, Description: "Simple wedges to split branches cleanly.", MinBushcraft: 1, WoodKg: 0.3, Effects: statDelta{Energy: 1}},
		{ID: "pack_frame", Name: "Pack Frame", BiomeTags: []string{"forest", "mountain", "boreal", "savanna", "badlands"}, Description: "Frame for carrying wood and game loads.", MinBushcraft: 2, WoodKg: 1.4, RequiresResources: []ResourceRequirement{{ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "digging_stick", Name: "Digging Stick", BiomeTags: []string{"desert", "dry", "savanna", "forest", "jungle", "badlands"}, Description: "Hardened digging tool for roots and clay.", MinBushcraft: 0, WoodKg: 0.35, Effects: statDelta{Energy: 1}},
		{ID: "bark_container", Name: "Bark Container", BiomeTags: []string{"forest", "boreal", "coast", "mountain", "jungle"}, Description: "Folded bark container for transport and storage.", MinBushcraft: 1, WoodKg: 0.2, RequiresResources: []ResourceRequirement{{ID: "birch_bark", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "wooden_cup", Name: "Wooden Cup", BiomeTags: []string{"forest", "boreal", "mountain", "coast", "savanna"}, Description: "Simple carved cup for hot liquids.", MinBushcraft: 1, WoodKg: 0.3, Effects: statDelta{Morale: 1}},
		{ID: "wooden_spoon", Name: "Wooden Spoon", BiomeTags: []string{"forest", "boreal", "mountain", "coast", "jungle"}, Description: "Carved spoon improving camp meal routine.", MinBushcraft: 0, WoodKg: 0.08, Effects: statDelta{Morale: 1}},
		{ID: "char_box", Name: "Char Box", BiomeTags: []string{"forest", "boreal", "mountain", "coast", "savanna", "badlands"}, Description: "Container for producing charred tinder.", MinBushcraft: 2, RequiresFire: true, WoodKg: 0.4, RequiresResources: []ResourceRequirement{{ID: "charcoal", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "resin_torch", Name: "Resin Torch", BiomeTags: []string{"forest", "boreal", "mountain", "jungle", "savanna"}, Description: "Hand torch for night movement and signaling.", MinBushcraft: 1, WoodKg: 0.35, RequiresResources: []ResourceRequirement{{ID: "resin", Qty: 1}, {ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Morale: 2}},
		{ID: "pitch_glue", Name: "Pitch Glue", BiomeTags: []string{"forest", "boreal", "mountain", "savanna", "jungle"}, Description: "Sticky resin pitch for hafting repairs.", MinBushcraft: 2, RequiresFire: true, WoodKg: 0.2, RequiresResources: []ResourceRequirement{{ID: "resin", Qty: 1}, {ID: "charcoal", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "arrow_shafts", Name: "Arrow Shafts", BiomeTags: []string{"forest", "savanna", "badlands", "mountain", "coast"}, Description: "Shaft bundle for bow maintenance.", MinBushcraft: 2, WoodKg: 0.35, RequiresItems: []string{"bow_stave"}, Effects: statDelta{Morale: 1}},
		{ID: "bow_stave", Name: "Bow Stave", BiomeTags: []string{"forest", "savanna", "badlands", "mountain", "coast"}, Description: "Seasoned stave for primitive bow builds.", MinBushcraft: 2, WoodKg: 0.9, RequiresResources: []ResourceRequirement{{ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Morale: 2}},
		{ID: "walking_staff", Name: "Walking Staff", BiomeTags: []string{"forest", "mountain", "badlands", "savanna", "coast"}, Description: "Staff improves rough-terrain movement confidence.", MinBushcraft: 0, WoodKg: 0.45, Effects: statDelta{Energy: 1}},
		{ID: "split_basket", Name: "Split-Wood Basket", BiomeTags: []string{"forest", "boreal", "wetlands", "delta", "jungle"}, Description: "Basket for carrying food and gathered supplies.", MinBushcraft: 2, WoodKg: 0.9, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 1}}, Effects: statDelta{Morale: 1}},

		// Existing camp utility builds, expanded.
		{ID: "tripod", Name: "Cooking Tripod", BiomeTags: []string{"forest", "coast", "mountain", "jungle"}, Description: "Stabilizes cooking over open flame.", MinBushcraft: 0, RequiresFire: true, WoodKg: 0.8, Effects: statDelta{Morale: 1}},
		{ID: "rain_catcher", Name: "Rain Catcher", BiomeTags: []string{"jungle", "wetlands", "coast", "island", "forest"}, Description: "Collects safer water during rain events.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 0.6, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 1}}, Effects: statDelta{Hydration: 2, Morale: 1}},
		{ID: "smoke_rack", Name: "Smoke Rack", BiomeTags: []string{"forest", "coast", "savanna", "jungle"}, Description: "Improves food preservation and morale.", MinBushcraft: 1, RequiresFire: true, WoodKg: 1.2, Effects: statDelta{Energy: 1, Morale: 2}},
		{ID: "fish_trap", Name: "Fish Trap", BiomeTags: []string{"delta", "river", "lake", "swamp", "coast"}, Description: "Passive fish capture near water channels.", MinBushcraft: 1, WoodKg: 0.7, RequiresResources: []ResourceRequirement{{ID: "vine_fiber", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "windbreak", Name: "Windbreak Wall", BiomeTags: []string{"arctic", "subarctic", "mountain", "badlands", "coast"}, Description: "Reduces nightly wind chill exposure.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.0, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "charcoal_bed", Name: "Charcoal Bed", BiomeTags: []string{"desert", "dry", "savanna", "badlands"}, Description: "Retains heat after sunset in dry zones.", MinBushcraft: 2, RequiresFire: true, WoodKg: 1.4, Effects: statDelta{Energy: 2}},
		{ID: "bug_smudge", Name: "Bug Smudge Fire", BiomeTags: []string{"jungle", "swamp", "wetlands", "tropical", "island"}, Description: "Smoky fire that repels insects around camp.", MinBushcraft: 1, RequiresFire: true, WoodKg: 0.5, RequiresResources: []ResourceRequirement{{ID: "resin", Qty: 1}}, Effects: statDelta{Morale: 2}},
		{ID: "snow_melt_station", Name: "Snow Melt Station", BiomeTags: []string{"arctic", "subarctic", "tundra", "winter"}, Description: "Converts snow to usable water efficiently.", MinBushcraft: 1, RequiresFire: true, WoodKg: 0.9, Effects: statDelta{Hydration: 2}},
		{ID: "raised_bed", Name: "Raised Bed", BiomeTags: []string{"swamp", "wetlands", "jungle", "forest"}, Description: "Improves overnight rest by lifting off damp ground.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 1.1, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 1}}, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "signal_beacon", Name: "Signal Beacon", BiomeTags: []string{"coast", "island", "mountain", "badlands", "savanna"}, Description: "High-visibility signal structure.", MinBushcraft: 0, RequiresFire: true, WoodKg: 1.3, Effects: statDelta{Morale: 2}},

		// Clay-enabled builds for biomes where clay is available.
		{ID: "clay_pot", Name: "Clay Pot", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "badlands", "coast"}, Description: "Fire-hardened pot for boiling and stewing.", MinBushcraft: 2, RequiresFire: true, WoodKg: 0.4, RequiresResources: []ResourceRequirement{{ID: "clay", Qty: 1.2}}, Effects: statDelta{Hydration: 2, Morale: 1}},
		{ID: "clay_cook_plate", Name: "Clay Cook Plate", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "badlands", "coast"}, Description: "Flat fired clay plate for roasting and drying food.", MinBushcraft: 2, RequiresFire: true, WoodKg: 0.3, RequiresResources: []ResourceRequirement{{ID: "clay", Qty: 1.0}}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "clay_heat_core", Name: "Clay Heat Core", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "badlands", "coast"}, Description: "Baked clay core to retain heat into the night.", MinBushcraft: 3, RequiresFire: true, RequiresShelter: true, WoodKg: 0.6, RequiresResources: []ResourceRequirement{{ID: "clay", Qty: 1.5}}, Effects: statDelta{Energy: 2}},
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
	for _, needed := range chosen.RequiresItems {
		if !slices.Contains(s.CraftedItems, needed) {
			return CraftableSpec{}, fmt.Errorf("requires crafted item: %s", needed)
		}
	}
	for _, needed := range chosen.RequiresResources {
		if s.resourceQty(needed.ID) < needed.Qty {
			return CraftableSpec{}, fmt.Errorf("requires resource: %s %.1f", needed.ID, needed.Qty)
		}
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
	for _, needed := range chosen.RequiresResources {
		_ = s.consumeResourceStock(needed.ID, needed.Qty)
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
		parts = append(parts, fmt.Sprintf("%s %.1fkg (%d%% wet)", item.Type, item.Kg, int(math.Round(item.Wetness*100))))
	}
	return strings.Join(parts, ", ")
}

func formatFirePrep(prep FirePrepState) string {
	return fmt.Sprintf("tinder:%d (q%.2f) kindling:%d (q%.2f) feather:%d (q%.2f) embers:%d",
		prep.TinderBundles, prep.TinderQuality,
		prep.KindlingBundles, prep.KindlingQuality,
		prep.FeatherSticks, prep.FeatherQuality,
		prep.Embers,
	)
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
