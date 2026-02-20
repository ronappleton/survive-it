package game

import (
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Discovery summary:
// - Catalogs are defined as slice-returning registry functions and selected by biome tag substring matching.
// - Foraging is deterministic from seed/day/actor and currently consumes immediately via ForageAndConsume.
// - Shelter/crafting/trap systems use lightweight data-driven specs and command handlers over shared RunState.
type PlantCategory string

const (
	PlantCategoryAny       PlantCategory = "any"
	PlantCategoryRoots     PlantCategory = "roots"
	PlantCategoryBerries   PlantCategory = "berries"
	PlantCategoryFruits    PlantCategory = "fruits"
	PlantCategoryVegetable PlantCategory = "vegetables"
	PlantCategoryNutsSeeds PlantCategory = "nuts_seeds"
	PlantCategoryMedicinal PlantCategory = "medicinal"
	PlantCategoryToxic     PlantCategory = "toxic"
	PlantCategoryUtility   PlantCategory = "utility"
)

type PlantSpec struct {
	ID               string
	Name             string
	Category         PlantCategory
	BiomeTags        []string
	SeasonTags       []SeasonID
	YieldMinG        int
	YieldMaxG        int
	NutritionPer100g NutritionPer100g
	UtilityTags      []string
	Medicinal        int
	Toxicity         int
	ToxicSymptoms    []string
}

type ForageResult struct {
	Plant        PlantSpec
	HarvestGrams int
	Nutrition    NutritionTotals
}

func PlantCatalog() []PlantSpec {
	base := []PlantSpec{
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
	return append(base, expandedPlantCatalog()...)
}

func expandedPlantCatalog() []PlantSpec {
	return []PlantSpec{
		{ID: "wild_onion", Name: "Wild Onion", Category: PlantCategoryRoots, BiomeTags: []string{"forest", "grassland", "river", "mountain"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonDry}, YieldMinG: 60, YieldMaxG: 260, NutritionPer100g: NutritionPer100g{CaloriesKcal: 40, ProteinG: 1, FatG: 0, SugarG: 4}, UtilityTags: []string{"flavoring"}},
		{ID: "arrowhead_tuber", Name: "Arrowhead Tuber", Category: PlantCategoryRoots, BiomeTags: []string{"wetlands", "swamp", "lake", "river"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 100, YieldMaxG: 520, NutritionPer100g: NutritionPer100g{CaloriesKcal: 98, ProteinG: 2, FatG: 0, SugarG: 2}},
		{ID: "lotus_root", Name: "Lotus Root", Category: PlantCategoryRoots, BiomeTags: []string{"delta", "wetlands", "lake", "jungle"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 110, YieldMaxG: 640, NutritionPer100g: NutritionPer100g{CaloriesKcal: 74, ProteinG: 2, FatG: 0, SugarG: 1}},
		{ID: "wild_garlic", Name: "Wild Garlic", Category: PlantCategoryVegetable, BiomeTags: []string{"forest", "mountain", "river", "boreal"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWet}, YieldMinG: 50, YieldMaxG: 220, NutritionPer100g: NutritionPer100g{CaloriesKcal: 110, ProteinG: 6, FatG: 0, SugarG: 1}, Medicinal: 1, UtilityTags: []string{"antimicrobial"}},
		{ID: "amaranth_leaf", Name: "Wild Amaranth Greens", Category: PlantCategoryVegetable, BiomeTags: []string{"savanna", "badlands", "river", "forest"}, SeasonTags: []SeasonID{SeasonDry, SeasonWet}, YieldMinG: 90, YieldMaxG: 420, NutritionPer100g: NutritionPer100g{CaloriesKcal: 23, ProteinG: 3, FatG: 0, SugarG: 1}},
		{ID: "sea_beet", Name: "Sea Beet", Category: PlantCategoryVegetable, BiomeTags: []string{"coast", "island", "delta"}, SeasonTags: []SeasonID{SeasonDry, SeasonAutumn}, YieldMinG: 80, YieldMaxG: 390, NutritionPer100g: NutritionPer100g{CaloriesKcal: 19, ProteinG: 2, FatG: 0, SugarG: 0}},
		{ID: "purslane", Name: "Purslane", Category: PlantCategoryVegetable, BiomeTags: []string{"desert", "dry", "coast", "savanna", "forest"}, SeasonTags: []SeasonID{SeasonDry, SeasonWet}, YieldMinG: 80, YieldMaxG: 360, NutritionPer100g: NutritionPer100g{CaloriesKcal: 20, ProteinG: 2, FatG: 0, SugarG: 1}, Medicinal: 1},
		{ID: "chickweed", Name: "Chickweed", Category: PlantCategoryVegetable, BiomeTags: []string{"forest", "river", "wetlands", "coast"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 70, YieldMaxG: 350, NutritionPer100g: NutritionPer100g{CaloriesKcal: 18, ProteinG: 2, FatG: 0, SugarG: 1}, Medicinal: 1},

		{ID: "huckleberry", Name: "Huckleberry", Category: PlantCategoryBerries, BiomeTags: []string{"mountain", "forest", "boreal"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 60, YieldMaxG: 320, NutritionPer100g: NutritionPer100g{CaloriesKcal: 50, ProteinG: 1, FatG: 0, SugarG: 8}},
		{ID: "lingonberry", Name: "Lingonberry", Category: PlantCategoryBerries, BiomeTags: []string{"boreal", "subarctic", "tundra", "forest"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWinter}, YieldMinG: 50, YieldMaxG: 250, NutritionPer100g: NutritionPer100g{CaloriesKcal: 43, ProteinG: 1, FatG: 0, SugarG: 6}, Medicinal: 1},
		{ID: "elderberry", Name: "Elderberry", Category: PlantCategoryBerries, BiomeTags: []string{"forest", "river", "wetlands", "coast"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 70, YieldMaxG: 420, NutritionPer100g: NutritionPer100g{CaloriesKcal: 73, ProteinG: 1, FatG: 0, SugarG: 7}, Medicinal: 1, Toxicity: 1, ToxicSymptoms: []string{"nausea if unripe"}},
		{ID: "juniper_berry", Name: "Juniper Berry", Category: PlantCategoryBerries, BiomeTags: []string{"mountain", "dry", "boreal", "forest"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWinter}, YieldMinG: 25, YieldMaxG: 120, NutritionPer100g: NutritionPer100g{CaloriesKcal: 44, ProteinG: 0, FatG: 1, SugarG: 4}, Medicinal: 1, UtilityTags: []string{"flavoring"}},
		{ID: "cloudberry", Name: "Cloudberry", Category: PlantCategoryBerries, BiomeTags: []string{"tundra", "subarctic", "wetlands", "boreal"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 40, YieldMaxG: 210, NutritionPer100g: NutritionPer100g{CaloriesKcal: 51, ProteinG: 1, FatG: 0, SugarG: 6}},
		{ID: "cranberry", Name: "Cranberry", Category: PlantCategoryBerries, BiomeTags: []string{"wetlands", "swamp", "lake", "boreal"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWet}, YieldMinG: 60, YieldMaxG: 330, NutritionPer100g: NutritionPer100g{CaloriesKcal: 46, ProteinG: 0, FatG: 0, SugarG: 4}, Medicinal: 1},

		{ID: "wild_plum", Name: "Wild Plum", Category: PlantCategoryFruits, BiomeTags: []string{"forest", "savanna", "river"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 90, YieldMaxG: 460, NutritionPer100g: NutritionPer100g{CaloriesKcal: 46, ProteinG: 1, FatG: 0, SugarG: 10}},
		{ID: "fig", Name: "Wild Fig", Category: PlantCategoryFruits, BiomeTags: []string{"jungle", "tropical", "coast", "island"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 110, YieldMaxG: 560, NutritionPer100g: NutritionPer100g{CaloriesKcal: 74, ProteinG: 1, FatG: 0, SugarG: 16}},
		{ID: "persimmon", Name: "Persimmon", Category: PlantCategoryFruits, BiomeTags: []string{"forest", "mountain", "river", "coast"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 80, YieldMaxG: 430, NutritionPer100g: NutritionPer100g{CaloriesKcal: 81, ProteinG: 1, FatG: 0, SugarG: 18}},
		{ID: "soursop", Name: "Soursop", Category: PlantCategoryFruits, BiomeTags: []string{"jungle", "tropical", "wetlands"}, SeasonTags: []SeasonID{SeasonWet}, YieldMinG: 150, YieldMaxG: 950, NutritionPer100g: NutritionPer100g{CaloriesKcal: 66, ProteinG: 1, FatG: 0, SugarG: 13}},
		{ID: "breadfruit", Name: "Breadfruit", Category: PlantCategoryFruits, BiomeTags: []string{"island", "coast", "tropical"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 200, YieldMaxG: 1300, NutritionPer100g: NutritionPer100g{CaloriesKcal: 103, ProteinG: 1, FatG: 0, SugarG: 11}},

		{ID: "acorn", Name: "Acorn", Category: PlantCategoryNutsSeeds, BiomeTags: []string{"forest", "temperate", "mountain", "coast"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWinter}, YieldMinG: 120, YieldMaxG: 880, NutritionPer100g: NutritionPer100g{CaloriesKcal: 387, ProteinG: 6, FatG: 24, SugarG: 0}, UtilityTags: []string{"acorn flour"}},
		{ID: "pine_nut", Name: "Pine Nut", Category: PlantCategoryNutsSeeds, BiomeTags: []string{"boreal", "mountain", "forest", "subarctic"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWinter}, YieldMinG: 50, YieldMaxG: 300, NutritionPer100g: NutritionPer100g{CaloriesKcal: 673, ProteinG: 14, FatG: 68, SugarG: 4}},
		{ID: "hazelnut", Name: "Hazelnut", Category: PlantCategoryNutsSeeds, BiomeTags: []string{"forest", "temperate", "river"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 80, YieldMaxG: 460, NutritionPer100g: NutritionPer100g{CaloriesKcal: 628, ProteinG: 15, FatG: 61, SugarG: 4}},
		{ID: "beech_nut", Name: "Beech Nut", Category: PlantCategoryNutsSeeds, BiomeTags: []string{"forest", "boreal", "mountain"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 70, YieldMaxG: 360, NutritionPer100g: NutritionPer100g{CaloriesKcal: 576, ProteinG: 6, FatG: 50, SugarG: 1}},
		{ID: "water_lily_seed", Name: "Water Lily Seed", Category: PlantCategoryNutsSeeds, BiomeTags: []string{"lake", "wetlands", "swamp", "delta"}, SeasonTags: []SeasonID{SeasonWet}, YieldMinG: 90, YieldMaxG: 440, NutritionPer100g: NutritionPer100g{CaloriesKcal: 353, ProteinG: 17, FatG: 2, SugarG: 1}},

		{ID: "willow_herb", Name: "Willow Herb", Category: PlantCategoryMedicinal, BiomeTags: []string{"forest", "river", "boreal", "mountain"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 30, YieldMaxG: 140, NutritionPer100g: NutritionPer100g{CaloriesKcal: 18, ProteinG: 2, FatG: 0, SugarG: 1}, Medicinal: 2, UtilityTags: []string{"anti-inflammatory tea"}},
		{ID: "yarrow", Name: "Yarrow", Category: PlantCategoryMedicinal, BiomeTags: []string{"grassland", "savanna", "mountain", "forest"}, SeasonTags: []SeasonID{SeasonDry, SeasonAutumn}, YieldMinG: 25, YieldMaxG: 110, NutritionPer100g: NutritionPer100g{CaloriesKcal: 20, ProteinG: 1, FatG: 0, SugarG: 1}, Medicinal: 2, UtilityTags: []string{"wound herb"}},
		{ID: "chamomile", Name: "Wild Chamomile", Category: PlantCategoryMedicinal, BiomeTags: []string{"grassland", "forest", "river", "coast"}, SeasonTags: []SeasonID{SeasonDry, SeasonWet}, YieldMinG: 20, YieldMaxG: 90, NutritionPer100g: NutritionPer100g{CaloriesKcal: 17, ProteinG: 1, FatG: 0, SugarG: 0}, Medicinal: 2, UtilityTags: []string{"calming tea"}},
		{ID: "echinacea", Name: "Echinacea", Category: PlantCategoryMedicinal, BiomeTags: []string{"savanna", "forest", "badlands"}, SeasonTags: []SeasonID{SeasonDry, SeasonAutumn}, YieldMinG: 25, YieldMaxG: 110, NutritionPer100g: NutritionPer100g{CaloriesKcal: 24, ProteinG: 1, FatG: 0, SugarG: 1}, Medicinal: 2, UtilityTags: []string{"immune support"}},
		{ID: "comfrey", Name: "Comfrey", Category: PlantCategoryMedicinal, BiomeTags: []string{"river", "wetlands", "forest", "coast"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 30, YieldMaxG: 150, NutritionPer100g: NutritionPer100g{CaloriesKcal: 28, ProteinG: 4, FatG: 0, SugarG: 1}, Medicinal: 1, UtilityTags: []string{"poultice"}},
		{ID: "aloe_vera", Name: "Aloe Vera", Category: PlantCategoryMedicinal, BiomeTags: []string{"desert", "dry", "coast", "island"}, SeasonTags: []SeasonID{SeasonDry}, YieldMinG: 60, YieldMaxG: 280, NutritionPer100g: NutritionPer100g{CaloriesKcal: 15, ProteinG: 0, FatG: 0, SugarG: 0}, Medicinal: 2, UtilityTags: []string{"burn gel"}},
		{ID: "usnea_lichen", Name: "Usnea Lichen", Category: PlantCategoryMedicinal, BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, SeasonTags: []SeasonID{SeasonWet, SeasonWinter}, YieldMinG: 10, YieldMaxG: 60, NutritionPer100g: NutritionPer100g{CaloriesKcal: 12, ProteinG: 0, FatG: 0, SugarG: 0}, Medicinal: 1, UtilityTags: []string{"antiseptic wash"}},

		{ID: "hemlock_shoot", Name: "Poison Hemlock Shoot", Category: PlantCategoryToxic, BiomeTags: []string{"river", "wetlands", "forest", "temperate"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 20, YieldMaxG: 90, NutritionPer100g: NutritionPer100g{CaloriesKcal: 15, ProteinG: 1, FatG: 0, SugarG: 0}, Toxicity: 5, ToxicSymptoms: []string{"neurological collapse"}},
		{ID: "oleander_leaf", Name: "Oleander Leaf", Category: PlantCategoryToxic, BiomeTags: []string{"coast", "desert", "dry", "island"}, SeasonTags: []SeasonID{SeasonDry}, YieldMinG: 15, YieldMaxG: 70, NutritionPer100g: NutritionPer100g{CaloriesKcal: 11, ProteinG: 1, FatG: 0, SugarG: 0}, Toxicity: 5, ToxicSymptoms: []string{"cardiac distress"}},
		{ID: "nightshade_berry", Name: "Nightshade Berry", Category: PlantCategoryToxic, BiomeTags: []string{"forest", "badlands", "savanna", "coast"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonWet}, YieldMinG: 25, YieldMaxG: 120, NutritionPer100g: NutritionPer100g{CaloriesKcal: 22, ProteinG: 1, FatG: 0, SugarG: 2}, Toxicity: 4, ToxicSymptoms: []string{"vomiting", "confusion"}},
		{ID: "moonseed", Name: "Moonseed", Category: PlantCategoryToxic, BiomeTags: []string{"forest", "river", "wetlands"}, SeasonTags: []SeasonID{SeasonAutumn}, YieldMinG: 20, YieldMaxG: 110, NutritionPer100g: NutritionPer100g{CaloriesKcal: 20, ProteinG: 1, FatG: 0, SugarG: 2}, Toxicity: 4, ToxicSymptoms: []string{"severe cramps"}},
		{ID: "castor_seed", Name: "Castor Seed", Category: PlantCategoryToxic, BiomeTags: []string{"savanna", "badlands", "dry", "tropical"}, SeasonTags: []SeasonID{SeasonDry, SeasonWet}, YieldMinG: 20, YieldMaxG: 100, NutritionPer100g: NutritionPer100g{CaloriesKcal: 80, ProteinG: 4, FatG: 3, SugarG: 1}, Toxicity: 5, ToxicSymptoms: []string{"organ damage"}},

		{ID: "dogbane_fiber", Name: "Dogbane Fiber Plant", Category: PlantCategoryUtility, BiomeTags: []string{"forest", "river", "savanna", "badlands"}, SeasonTags: []SeasonID{SeasonAutumn, SeasonDry}, YieldMinG: 40, YieldMaxG: 180, NutritionPer100g: NutritionPer100g{CaloriesKcal: 8, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"strong cordage"}},
		{ID: "flax_stalk", Name: "Wild Flax Stalk", Category: PlantCategoryUtility, BiomeTags: []string{"grassland", "savanna", "river", "forest"}, SeasonTags: []SeasonID{SeasonDry, SeasonAutumn}, YieldMinG: 35, YieldMaxG: 170, NutritionPer100g: NutritionPer100g{CaloriesKcal: 12, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"thread", "cloth fiber"}},
		{ID: "milkweed_stalk", Name: "Milkweed Stalk", Category: PlantCategoryUtility, BiomeTags: []string{"savanna", "badlands", "forest", "coast"}, SeasonTags: []SeasonID{SeasonDry, SeasonAutumn}, YieldMinG: 45, YieldMaxG: 190, NutritionPer100g: NutritionPer100g{CaloriesKcal: 12, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"cordage", "insulation floss"}},
		{ID: "cattail_leaf", Name: "Cattail Leaf", Category: PlantCategoryUtility, BiomeTags: []string{"wetlands", "swamp", "delta", "lake"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 80, YieldMaxG: 420, NutritionPer100g: NutritionPer100g{CaloriesKcal: 12, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"matting", "thatch", "basketry"}},
		{ID: "bulrush", Name: "Bulrush", Category: PlantCategoryUtility, BiomeTags: []string{"wetlands", "swamp", "delta", "river", "lake"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 90, YieldMaxG: 500, NutritionPer100g: NutritionPer100g{CaloriesKcal: 10, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"raft lashings", "weaving"}},
		{ID: "palm_frond", Name: "Palm Frond", Category: PlantCategoryUtility, BiomeTags: []string{"island", "coast", "tropical", "delta"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 90, YieldMaxG: 380, NutritionPer100g: NutritionPer100g{CaloriesKcal: 10, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"thatch", "fans", "screens"}},
		{ID: "bamboo_culm", Name: "Bamboo Culm", Category: PlantCategoryUtility, BiomeTags: []string{"jungle", "wetlands", "tropical"}, SeasonTags: []SeasonID{SeasonWet, SeasonDry}, YieldMinG: 200, YieldMaxG: 1200, NutritionPer100g: NutritionPer100g{CaloriesKcal: 10, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"poles", "containers", "rafts"}},
		{ID: "reed_mace", Name: "Reed Mace", Category: PlantCategoryUtility, BiomeTags: []string{"delta", "wetlands", "lake", "river"}, SeasonTags: []SeasonID{SeasonWet, SeasonAutumn}, YieldMinG: 30, YieldMaxG: 140, NutritionPer100g: NutritionPer100g{CaloriesKcal: 14, ProteinG: 0, FatG: 0, SugarG: 0}, UtilityTags: []string{"tinder fluff", "insulation"}},
	}
}

func PlantCategories() []PlantCategory {
	return []PlantCategory{
		PlantCategoryAny,
		PlantCategoryRoots,
		PlantCategoryBerries,
		PlantCategoryFruits,
		PlantCategoryVegetable,
		PlantCategoryNutsSeeds,
		PlantCategoryMedicinal,
		PlantCategoryToxic,
		PlantCategoryUtility,
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
	case "nuts", "nut", "seeds", "seed", "nuts_seeds":
		return PlantCategoryNutsSeeds
	case "medicinal", "medicine", "herb", "herbs":
		return PlantCategoryMedicinal
	case "toxic", "poison", "poisonous":
		return PlantCategoryToxic
	case "utility", "fiber", "fibre", "cordage", "craft":
		return PlantCategoryUtility
	default:
		return PlantCategoryAny
	}
}

func PlantsForBiome(biome string, category PlantCategory) []PlantSpec {
	return filteredPlantsForBiome(biome, category, "")
}

func PlantsForBiomeSeason(biome string, category PlantCategory, season SeasonID) []PlantSpec {
	return filteredPlantsForBiome(biome, category, season)
}

func filteredPlantsForBiome(biome string, category PlantCategory, season SeasonID) []PlantSpec {
	norm := normalizeBiome(biome)
	catalog := PlantCatalog()
	filtered := make([]PlantSpec, 0, len(catalog))
	for _, plant := range catalog {
		if category != PlantCategoryAny && plant.Category != category {
			continue
		}
		if !plantSeasonMatches(plant, season) {
			continue
		}
		for _, tag := range plant.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				filtered = append(filtered, plant)
				break
			}
		}
	}

	if len(filtered) == 0 && season != "" {
		return filteredPlantsForBiome(biome, category, "")
	}
	if len(filtered) == 0 && category != PlantCategoryAny {
		return filteredPlantsForBiome(biome, PlantCategoryAny, season)
	}
	return filtered
}

func RandomForage(seed int64, biome string, category PlantCategory, day, actorID int) (ForageResult, error) {
	return randomForageWithSeason(seed, biome, category, day, actorID, "")
}

func RandomForageForSeason(seed int64, biome string, category PlantCategory, season SeasonID, day, actorID int) (ForageResult, error) {
	return randomForageWithSeason(seed, biome, category, day, actorID, season)
}

func RandomForageForSeasonWithClimate(seed int64, biome string, category PlantCategory, season SeasonID, day, actorID int, climate *ClimateProfile, tempC int) (ForageResult, error) {
	return randomForageWithSeasonClimate(seed, biome, category, day, actorID, season, climate, tempC)
}

func randomForageWithSeason(seed int64, biome string, category PlantCategory, day, actorID int, season SeasonID) (ForageResult, error) {
	return randomForageWithSeasonClimate(seed, biome, category, day, actorID, season, nil, 0)
}

func randomForageWithSeasonClimate(seed int64, biome string, category PlantCategory, day, actorID int, season SeasonID, climate *ClimateProfile, tempC int) (ForageResult, error) {
	pool := filteredPlantsForBiome(biome, category, season)
	pool = filterPlantsForClimate(pool, climate, season, tempC)
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

	season, ok := s.CurrentSeason()
	if !ok {
		season = ""
	}
	biome := s.CurrentBiomeQuery()
	if strings.TrimSpace(biome) == "" {
		biome = s.Scenario.Biome
	}
	forage, err := RandomForageForSeasonWithClimate(s.Config.Seed, biome, category, season, s.Day, playerID, s.ActiveClimateProfile(), s.Weather.TemperatureC)
	if err != nil {
		return ForageResult{}, err
	}
	applySkillEffort(&player.Foraging, 16, true)
	applySkillEffort(&player.Gathering, 10, true)
	bonusPct := player.Foraging/10 + player.Agility + positiveTraitModifier(player.Traits)/2
	if bonusPct != 0 {
		forage.HarvestGrams = max(1, forage.HarvestGrams+(forage.HarvestGrams*bonusPct)/100)
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
	s.applyForagePlantEffects(playerID, player, forage)
	refreshEffectBars(player)

	return forage, nil
}

func plantSeasonMatches(spec PlantSpec, season SeasonID) bool {
	if len(spec.SeasonTags) == 0 || season == "" {
		return true
	}
	for _, tag := range spec.SeasonTags {
		if tag == season {
			return true
		}
	}
	return false
}

func (s *RunState) applyForagePlantEffects(playerID int, player *PlayerState, forage ForageResult) {
	if s == nil || player == nil {
		return
	}
	if forage.Plant.Medicinal > 0 {
		player.Morale = clamp(player.Morale+clamp(forage.Plant.Medicinal, 1, 3), 0, 100)
		player.Hydration = clamp(player.Hydration+clamp(forage.Plant.Medicinal/2, 0, 2), 0, 100)
		if len(player.Ailments) > 0 {
			reliefRoll := deterministicForageRoll(s.Config.Seed, s.Day, playerID, forage.Plant.ID, "medicinal")
			if reliefRoll <= 0.12+float64(forage.Plant.Medicinal)*0.05 {
				player.Ailments[0].DaysRemaining = maxInt(0, player.Ailments[0].DaysRemaining-1)
				if player.Ailments[0].DaysRemaining == 0 {
					player.Ailments = append([]Ailment{}, player.Ailments[1:]...)
				}
			}
		}
	}
	if forage.Plant.Toxicity <= 0 {
		return
	}
	toxicChance := float64(clamp(forage.Plant.Toxicity, 1, 5))*0.08 - float64(player.Foraging)/300.0 - float64(max(0, player.MentalStrength))/300.0
	toxicChance = clampFloat(toxicChance, 0.02, 0.42)
	roll := deterministicForageRoll(s.Config.Seed, s.Day, playerID, forage.Plant.ID, "toxicity")
	if roll > toxicChance {
		return
	}
	penaltyScale := clamp(forage.Plant.Toxicity, 1, 5)
	player.Energy = clamp(player.Energy-penaltyScale, 0, 100)
	player.Hydration = clamp(player.Hydration-(penaltyScale+1), 0, 100)
	player.Morale = clamp(player.Morale-penaltyScale, 0, 100)
	ailment := Ailment{
		Type:             AilmentFoodPoison,
		Name:             "Foraged Plant Poisoning",
		DaysRemaining:    1 + penaltyScale/2,
		EnergyPenalty:    penaltyScale,
		HydrationPenalty: penaltyScale + 1,
		MoralePenalty:    penaltyScale,
	}
	player.applyAilment(ailment)
	if penaltyScale >= 4 {
		player.applyAilment(Ailment{
			Type:             AilmentVomiting,
			Name:             "Severe Plant Toxin Reaction",
			DaysRemaining:    1,
			EnergyPenalty:    penaltyScale,
			HydrationPenalty: penaltyScale + 2,
			MoralePenalty:    penaltyScale,
		})
	}
}

func deterministicForageRoll(seed int64, day, playerID int, plantID, salt string) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d:%d:%s:%s", seed, day, playerID, plantID, salt)))
	return float64(h.Sum64()%10000) / 10000.0
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
	Uses      []string
}

type ResourceStock struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Unit string  `json:"unit"`
	Qty  float64 `json:"qty"`
}

func ResourceCatalog() []ResourceSpec {
	base := []ResourceSpec{
		{ID: "dry_grass", Name: "Dry Grass", BiomeTags: []string{"savanna", "badlands", "dry", "desert", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.85, Flammable: true, Uses: []string{"tinder", "thatch", "insulation"}},
		{ID: "birch_bark", Name: "Birch Bark", BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, Unit: "sheet", GatherMin: 1, GatherMax: 3, Dryness: 0.8, Flammable: true, Uses: []string{"tinder", "container", "roofing"}},
		{ID: "cedar_bark", Name: "Cedar Bark", BiomeTags: []string{"coast", "temperate_rainforest", "vancouver", "forest"}, Unit: "sheet", GatherMin: 1, GatherMax: 3, Dryness: 0.78, Flammable: true, Uses: []string{"cordage", "roofing", "bedding"}},
		{ID: "inner_bark_fiber", Name: "Inner Bark Fiber", BiomeTags: []string{"forest", "boreal", "savanna", "jungle"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.72, Flammable: true, Uses: []string{"cordage", "twine", "basketry"}},
		{ID: "fatwood", Name: "Fatwood", BiomeTags: []string{"boreal", "forest", "mountain", "dry"}, Unit: "stick", GatherMin: 1, GatherMax: 3, Dryness: 0.9, Flammable: true, Uses: []string{"firestarter"}},
		{ID: "resin", Name: "Tree Resin", BiomeTags: []string{"forest", "boreal", "mountain", "jungle", "savanna"}, Unit: "lump", GatherMin: 1, GatherMax: 3, Dryness: 0.95, Flammable: true, Uses: []string{"pitch glue", "torch fuel", "sealant"}},
		{ID: "punkwood", Name: "Punkwood", BiomeTags: []string{"forest", "boreal", "swamp", "wetlands"}, Unit: "chunk", GatherMin: 1, GatherMax: 3, Dryness: 0.55, Flammable: true, Uses: []string{"ember catch", "smudge"}},
		{ID: "cattail_fluff", Name: "Cattail Fluff", BiomeTags: []string{"wetlands", "swamp", "delta", "lake"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.82, Flammable: true, Uses: []string{"tinder", "insulation", "wound dressing"}},
		{ID: "dry_moss", Name: "Dry Moss", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain", "coast"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.76, Flammable: true, Uses: []string{"tinder", "padding", "water filter pre-layer"}},
		{ID: "coconut_husk", Name: "Coconut Husk", BiomeTags: []string{"island", "coast", "tropical", "delta"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.74, Flammable: true, Uses: []string{"coir fiber", "tinder"}},
		{ID: "reed_bundle", Name: "Reed Bundle", BiomeTags: []string{"wetlands", "swamp", "delta", "lake", "river"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.58, Flammable: true, Uses: []string{"basketry", "matting", "thatch"}},
		{ID: "vine_fiber", Name: "Vine Fiber", BiomeTags: []string{"jungle", "tropical", "wetlands", "swamp"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.52, Flammable: true, Uses: []string{"lashing", "traps", "raft lashings"}},
		{ID: "yucca_fiber", Name: "Yucca Fiber", BiomeTags: []string{"desert", "dry", "badlands", "savanna"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.62, Flammable: true, Uses: []string{"twine", "sandals", "rope"}},
		{ID: "nettle_fiber", Name: "Nettle Fiber", BiomeTags: []string{"forest", "boreal", "river", "mountain"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.66, Flammable: true, Uses: []string{"twine", "thread", "cloth"}},
		{ID: "milkweed_fiber", Name: "Milkweed Fiber", BiomeTags: []string{"savanna", "badlands", "temperate", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.65, Flammable: true, Uses: []string{"cordage", "insulation"}},
		{ID: "willow_bark", Name: "Willow Bark", BiomeTags: []string{"river", "lake", "wetlands", "forest", "boreal"}, Unit: "sheet", GatherMin: 1, GatherMax: 3, Dryness: 0.45, Flammable: false, Uses: []string{"medicinal tea", "pain relief"}},
		{ID: "pine_needles", Name: "Pine Needles", BiomeTags: []string{"boreal", "forest", "mountain", "subarctic"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.64, Flammable: true, Uses: []string{"vitamin tea", "bedding", "smudge"}},
		{ID: "spruce_root", Name: "Spruce Root", BiomeTags: []string{"boreal", "forest", "subarctic", "mountain"}, Unit: "bundle", GatherMin: 1, GatherMax: 2, Dryness: 0.44, Flammable: false, Uses: []string{"sewing", "lashing"}},
		{ID: "sphagnum_moss", Name: "Sphagnum Moss", BiomeTags: []string{"wetlands", "swamp", "boreal", "subarctic"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.36, Flammable: false, Uses: []string{"wound dressing", "padding"}},
		{ID: "plantain_leaf", Name: "Plantain Leaf", BiomeTags: []string{"temperate", "forest", "river", "savanna"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.42, Flammable: false, Uses: []string{"topical poultice", "wraps"}},
		{ID: "aloe_leaf", Name: "Aloe Leaf", BiomeTags: []string{"desert", "dry", "coast", "tropical"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.2, Flammable: false, Uses: []string{"burn care", "skin treatment"}},
		{ID: "dock_leaf", Name: "Dock Leaf", BiomeTags: []string{"river", "lake", "wetlands", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.38, Flammable: false, Uses: []string{"sting relief", "wrap"}},
		{ID: "bast_strip", Name: "Bast Strip", BiomeTags: []string{"forest", "boreal", "coast", "mountain"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.63, Flammable: true, Uses: []string{"strong cordage", "weaving"}},
		{ID: "rawhide_strip", Name: "Rawhide Strip", BiomeTags: []string{"savanna", "badlands", "forest", "mountain", "coast"}, Unit: "hide", GatherMin: 1, GatherMax: 2, Dryness: 0.3, Flammable: false, Uses: []string{"clothing", "lace", "container"}},
		{ID: "stone_flake", Name: "Stone Flake", BiomeTags: []string{"mountain", "badlands", "desert", "river", "coast"}, Unit: "piece", GatherMin: 1, GatherMax: 4, Dryness: 1.0, Flammable: false, Uses: []string{"scraping", "notching", "cutting"}},
		{ID: "clay", Name: "Clay", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "badlands", "coast"}, Unit: "kg", GatherMin: 0.4, GatherMax: 2.0, Dryness: 0.2, Flammable: false, Uses: []string{"pottery", "heat cores", "sealant"}},
		{ID: "charcoal", Name: "Charcoal", BiomeTags: []string{"forest", "savanna", "jungle", "mountain", "coast"}, Unit: "chunk", GatherMin: 1, GatherMax: 3, Dryness: 0.95, Flammable: true, Uses: []string{"water filter", "pigment", "fire extender"}},
		{ID: "drift_reed_fiber", Name: "Drift Reed Fiber", BiomeTags: []string{"delta", "coast", "island", "wetlands"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.5, Flammable: true, Uses: []string{"netting", "cordage"}},
		{ID: "seaweed_blade", Name: "Seaweed Blade", BiomeTags: []string{"coast", "island", "delta"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.25, Flammable: false, Uses: []string{"wrap", "compost", "iodine rinse"}},
	}
	return append(base, expandedResourceCatalog()...)
}

func expandedResourceCatalog() []ResourceSpec {
	return []ResourceSpec{
		{ID: "spruce_bough", Name: "Spruce Bough", BiomeTags: []string{"boreal", "subarctic", "mountain", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.52, Flammable: true, Uses: []string{"bedding", "thatch", "shelter lining"}},
		{ID: "fir_bough", Name: "Fir Bough", BiomeTags: []string{"forest", "mountain", "boreal"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.54, Flammable: true, Uses: []string{"bedding", "insulation"}},
		{ID: "willow_withy", Name: "Willow Withy", BiomeTags: []string{"river", "lake", "wetlands", "delta", "swamp"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.34, Flammable: false, Uses: []string{"wattle walls", "basket frame", "fish weir"}},
		{ID: "hemp_fiber", Name: "Hemp Fiber", BiomeTags: []string{"savanna", "river", "forest", "badlands"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.62, Flammable: true, Uses: []string{"rope", "cloth", "net"}},
		{ID: "flax_fiber", Name: "Flax Fiber", BiomeTags: []string{"grassland", "river", "forest", "coast"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.66, Flammable: true, Uses: []string{"thread", "cloth"}},
		{ID: "bamboo_strip", Name: "Bamboo Strip", BiomeTags: []string{"jungle", "wetlands", "tropical"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.58, Flammable: true, Uses: []string{"lashing", "trap body", "splints"}},
		{ID: "palm_frond", Name: "Palm Frond", BiomeTags: []string{"island", "coast", "tropical", "delta"}, Unit: "bundle", GatherMin: 1, GatherMax: 5, Dryness: 0.5, Flammable: true, Uses: []string{"thatch", "screen", "mat"}},
		{ID: "liana_vine", Name: "Liana Vine", BiomeTags: []string{"jungle", "wetlands", "swamp", "tropical"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.44, Flammable: true, Uses: []string{"heavy lashing", "traps"}},
		{ID: "bark_pitch", Name: "Bark Pitch", BiomeTags: []string{"forest", "boreal", "mountain", "coast"}, Unit: "lump", GatherMin: 1, GatherMax: 2, Dryness: 0.9, Flammable: true, Uses: []string{"sealant", "adhesive", "waterproofing"}},
		{ID: "birch_tar", Name: "Birch Tar", BiomeTags: []string{"boreal", "forest", "mountain", "subarctic"}, Unit: "lump", GatherMin: 1, GatherMax: 2, Dryness: 0.93, Flammable: true, Uses: []string{"adhesive", "waterproofing"}},
		{ID: "lichen", Name: "Lichen", BiomeTags: []string{"tundra", "arctic", "boreal", "mountain"}, Unit: "bundle", GatherMin: 1, GatherMax: 3, Dryness: 0.43, Flammable: true, Uses: []string{"tinder", "medicine", "insulation"}},
		{ID: "horsetail_reed", Name: "Horsetail Reed", BiomeTags: []string{"wetlands", "river", "lake", "forest"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.57, Flammable: true, Uses: []string{"abrasive polish", "matting"}},
		{ID: "mushroom_tinder", Name: "Tinder Fungus", BiomeTags: []string{"forest", "boreal", "mountain", "wetlands"}, Unit: "chunk", GatherMin: 1, GatherMax: 3, Dryness: 0.7, Flammable: true, Uses: []string{"tinder", "ember transport"}},
		{ID: "thatch_bundle", Name: "Thatch Bundle", BiomeTags: []string{"savanna", "wetlands", "river", "coast"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.68, Flammable: true, Uses: []string{"roofing", "insulation"}},
		{ID: "drift_log", Name: "Drift Log", BiomeTags: []string{"coast", "delta", "island", "lake"}, Unit: "piece", GatherMin: 1, GatherMax: 3, Dryness: 0.55, Flammable: true, Uses: []string{"platform beams", "raft body"}},
		{ID: "gravel", Name: "Gravel", BiomeTags: []string{"river", "delta", "mountain", "badlands", "coast"}, Unit: "kg", GatherMin: 0.5, GatherMax: 3.0, Dryness: 1.0, Flammable: false, Uses: []string{"drainage", "hearth base"}},
		{ID: "sand", Name: "Sand", BiomeTags: []string{"coast", "delta", "desert", "river"}, Unit: "kg", GatherMin: 0.5, GatherMax: 3.0, Dryness: 1.0, Flammable: false, Uses: []string{"drainage", "pit lining"}},
		{ID: "shell_fragment", Name: "Shell Fragment", BiomeTags: []string{"coast", "island", "delta", "wetlands"}, Unit: "piece", GatherMin: 1, GatherMax: 6, Dryness: 1.0, Flammable: false, Uses: []string{"scraper", "awls", "ornament"}},
		{ID: "stone_cobble", Name: "Stone Cobble", BiomeTags: []string{"mountain", "river", "badlands", "coast", "desert"}, Unit: "piece", GatherMin: 1, GatherMax: 5, Dryness: 1.0, Flammable: false, Uses: []string{"hearth", "deadfall weight", "oven"}},
		{ID: "mud", Name: "Mud", BiomeTags: []string{"wetlands", "swamp", "river", "delta", "lake"}, Unit: "kg", GatherMin: 0.5, GatherMax: 2.4, Dryness: 0.15, Flammable: false, Uses: []string{"wattle daub", "seal"}},
		{ID: "peat", Name: "Peat", BiomeTags: []string{"wetlands", "swamp", "boreal", "subarctic"}, Unit: "kg", GatherMin: 0.4, GatherMax: 1.8, Dryness: 0.4, Flammable: true, Uses: []string{"smolder fuel", "insulation"}},
		{ID: "dry_leaf_litter", Name: "Dry Leaf Litter", BiomeTags: []string{"forest", "boreal", "savanna", "mountain"}, Unit: "bundle", GatherMin: 1, GatherMax: 5, Dryness: 0.78, Flammable: true, Uses: []string{"tinder", "bedding"}},
		{ID: "cane_stalk", Name: "Cane Stalk", BiomeTags: []string{"wetlands", "delta", "jungle", "river"}, Unit: "bundle", GatherMin: 1, GatherMax: 4, Dryness: 0.53, Flammable: true, Uses: []string{"frames", "screens", "fishing stakes"}},
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

func (s *RunState) addResourceStock(resource ResourceSpec, qty float64) error {
	if s == nil || qty <= 0 {
		return nil
	}
	if !s.canStoreAtCamp(qty * defaultUnitWeightKg(resource.Unit)) {
		return fmt.Errorf("camp inventory full (%.1f/%.1fkg)", s.campUsedKg(), s.campCapacityKg())
	}
	for i := range s.ResourceStock {
		if s.ResourceStock[i].ID == resource.ID {
			s.ResourceStock[i].Qty += qty
			return nil
		}
	}
	s.ResourceStock = append(s.ResourceStock, ResourceStock{
		ID:   resource.ID,
		Name: resource.Name,
		Unit: resource.Unit,
		Qty:  qty,
	})
	return nil
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

	if err := s.addResourceStock(resource, qty); err != nil {
		return ResourceSpec{}, 0, err
	}
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
	ID            string
	Name          string
	BiomeTags     []string
	WoodType      WoodType
	GatherMinKg   float64
	GatherMaxKg   float64
	HeatFactor    float64
	BurnFactor    float64
	SparkEase     int
	Hardness      int
	Structural    int
	ResinQuality  float64
	RotResistance int
	SmokeFactor   float64
	BarkResource  string
	BarkUses      []string
	Tags          []string
}

func TreeCatalog() []TreeSpec {
	base := []TreeSpec{
		{ID: "cedar", Name: "Cedar", BiomeTags: []string{"coast", "temperate_rainforest", "vancouver", "forest"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.8, GatherMaxKg: 4.2, HeatFactor: 0.85, BurnFactor: 0.8, SparkEase: 3, Hardness: 2, Structural: 3, ResinQuality: 0.5, RotResistance: 5, SmokeFactor: 0.8, BarkResource: "cedar_bark", BarkUses: []string{"cordage", "roofing"}, Tags: []string{"conifer"}},
		{ID: "spruce", Name: "Spruce", BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, WoodType: WoodTypeResinous, GatherMinKg: 0.9, GatherMaxKg: 4.5, HeatFactor: 0.92, BurnFactor: 0.85, SparkEase: 4, Hardness: 3, Structural: 4, ResinQuality: 0.85, RotResistance: 3, SmokeFactor: 1.1, BarkResource: "spruce_root", BarkUses: []string{"sewing", "lashing"}, Tags: []string{"conifer"}},
		{ID: "pine", Name: "Pine", BiomeTags: []string{"mountain", "boreal", "forest", "dry"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.1, HeatFactor: 0.9, BurnFactor: 0.82, SparkEase: 4, Hardness: 2, Structural: 3, ResinQuality: 0.9, RotResistance: 3, SmokeFactor: 1.2, BarkResource: "inner_bark_fiber", BarkUses: []string{"tinder", "fiber"}, Tags: []string{"conifer"}},
		{ID: "fir", Name: "Fir", BiomeTags: []string{"forest", "mountain", "boreal"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.8, GatherMaxKg: 4.0, HeatFactor: 0.86, BurnFactor: 0.82, SparkEase: 3, Hardness: 2, Structural: 3, ResinQuality: 0.55, RotResistance: 3, SmokeFactor: 1.0, BarkResource: "inner_bark_fiber", BarkUses: []string{"tinder", "bedding"}, Tags: []string{"conifer"}},
		{ID: "birch", Name: "Birch", BiomeTags: []string{"boreal", "forest", "lake", "subarctic"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.8, HeatFactor: 1.0, BurnFactor: 1.05, SparkEase: 3, Hardness: 4, Structural: 4, ResinQuality: 0.35, RotResistance: 3, SmokeFactor: 0.7, BarkResource: "birch_bark", BarkUses: []string{"container", "tinder"}, Tags: []string{"hardwood"}},
		{ID: "oak", Name: "Oak", BiomeTags: []string{"forest", "temperate", "coast", "river"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.9, GatherMaxKg: 5.0, HeatFactor: 1.15, BurnFactor: 1.18, SparkEase: 2, Hardness: 5, Structural: 5, ResinQuality: 0.2, RotResistance: 5, SmokeFactor: 0.6, BarkResource: "bast_strip", BarkUses: []string{"tannin", "cordage"}, Tags: []string{"hardwood"}},
		{ID: "maple", Name: "Maple", BiomeTags: []string{"forest", "mountain", "temperate"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.4, HeatFactor: 1.1, BurnFactor: 1.1, SparkEase: 2, Hardness: 4, Structural: 4, ResinQuality: 0.18, RotResistance: 4, SmokeFactor: 0.65, BarkResource: "inner_bark_fiber", BarkUses: []string{"fiber", "container"}, Tags: []string{"hardwood"}},
		{ID: "willow", Name: "Willow", BiomeTags: []string{"wetlands", "swamp", "delta", "river", "lake"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 3.0, HeatFactor: 0.74, BurnFactor: 0.7, SparkEase: 2, Hardness: 2, Structural: 2, ResinQuality: 0.1, RotResistance: 2, SmokeFactor: 0.9, BarkResource: "willow_bark", BarkUses: []string{"medicine", "withies"}, Tags: []string{"wetland"}},
		{ID: "mangrove", Name: "Mangrove", BiomeTags: []string{"delta", "coast", "swamp", "island"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 3.8, HeatFactor: 1.12, BurnFactor: 1.06, SparkEase: 2, Hardness: 4, Structural: 4, ResinQuality: 0.2, RotResistance: 5, SmokeFactor: 0.85, BarkResource: "bast_strip", BarkUses: []string{"lashing", "waterproofing"}, Tags: []string{"wetland", "tropical"}},
		{ID: "bamboo", Name: "Bamboo", BiomeTags: []string{"jungle", "tropical", "wetlands"}, WoodType: WoodTypeBamboo, GatherMinKg: 0.9, GatherMaxKg: 5.5, HeatFactor: 0.78, BurnFactor: 0.65, SparkEase: 4, Hardness: 3, Structural: 4, ResinQuality: 0.05, RotResistance: 3, SmokeFactor: 0.75, BarkResource: "bamboo_strip", BarkUses: []string{"splints", "containers"}, Tags: []string{"tropical"}},
		{ID: "acacia", Name: "Acacia", BiomeTags: []string{"savanna", "badlands", "dry"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.5, HeatFactor: 1.08, BurnFactor: 1.02, SparkEase: 2, Hardness: 5, Structural: 4, ResinQuality: 0.18, RotResistance: 4, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"cordage", "tannin"}, Tags: []string{"hardwood", "dryland"}},
		{ID: "baobab", Name: "Baobab", BiomeTags: []string{"savanna", "dry", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 1.0, GatherMaxKg: 5.8, HeatFactor: 1.18, BurnFactor: 1.16, SparkEase: 2, Hardness: 4, Structural: 5, ResinQuality: 0.1, RotResistance: 4, SmokeFactor: 0.8, BarkResource: "bast_strip", BarkUses: []string{"fiber", "rope"}, Tags: []string{"hardwood", "dryland"}},
		{ID: "palm", Name: "Palm", BiomeTags: []string{"island", "coast", "tropical", "delta"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 2.6, HeatFactor: 0.72, BurnFactor: 0.68, SparkEase: 3, Hardness: 2, Structural: 2, ResinQuality: 0.08, RotResistance: 2, SmokeFactor: 0.9, BarkResource: "palm_frond", BarkUses: []string{"thatch", "weaving"}, Tags: []string{"tropical"}},
		{ID: "mesquite", Name: "Mesquite", BiomeTags: []string{"desert", "dry", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.4, HeatFactor: 1.2, BurnFactor: 1.2, SparkEase: 2, Hardness: 5, Structural: 4, ResinQuality: 0.16, RotResistance: 5, SmokeFactor: 0.6, BarkResource: "bast_strip", BarkUses: []string{"fiber", "smoke cure"}, Tags: []string{"hardwood", "dryland"}},
		{ID: "driftwood", Name: "Driftwood", BiomeTags: []string{"coast", "delta", "island", "lake"}, WoodType: WoodTypeDriftwood, GatherMinKg: 0.4, GatherMaxKg: 2.4, HeatFactor: 0.62, BurnFactor: 0.54, SparkEase: 3, Hardness: 3, Structural: 2, ResinQuality: 0.03, RotResistance: 3, SmokeFactor: 1.3, BarkUses: []string{"rafts", "platforms"}, Tags: []string{"coastal"}},
	}
	return append(base, expandedTreeCatalog()...)
}

func expandedTreeCatalog() []TreeSpec {
	return []TreeSpec{
		{ID: "hemlock", Name: "Hemlock", BiomeTags: []string{"forest", "boreal", "coast", "mountain"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.7, GatherMaxKg: 3.9, HeatFactor: 0.84, BurnFactor: 0.78, SparkEase: 3, Hardness: 2, Structural: 3, ResinQuality: 0.45, RotResistance: 3, SmokeFactor: 0.95, BarkResource: "inner_bark_fiber", BarkUses: []string{"tinder", "fiber"}, Tags: []string{"conifer"}},
		{ID: "larch", Name: "Larch", BiomeTags: []string{"boreal", "subarctic", "mountain", "forest"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.3, HeatFactor: 0.98, BurnFactor: 0.92, SparkEase: 4, Hardness: 3, Structural: 4, ResinQuality: 0.7, RotResistance: 4, SmokeFactor: 0.9, BarkResource: "inner_bark_fiber", BarkUses: []string{"cordage", "tinder"}, Tags: []string{"conifer"}},
		{ID: "cypress", Name: "Cypress", BiomeTags: []string{"swamp", "wetlands", "delta", "coast"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.8, GatherMaxKg: 4.1, HeatFactor: 0.9, BurnFactor: 0.86, SparkEase: 3, Hardness: 3, Structural: 4, ResinQuality: 0.4, RotResistance: 5, SmokeFactor: 0.85, BarkResource: "bast_strip", BarkUses: []string{"planking", "cordage"}, Tags: []string{"wetland"}},
		{ID: "juniper", Name: "Juniper", BiomeTags: []string{"desert", "dry", "mountain", "badlands"}, WoodType: WoodTypeResinous, GatherMinKg: 0.6, GatherMaxKg: 3.0, HeatFactor: 0.95, BurnFactor: 0.88, SparkEase: 4, Hardness: 3, Structural: 3, ResinQuality: 0.74, RotResistance: 4, SmokeFactor: 1.1, BarkResource: "inner_bark_fiber", BarkUses: []string{"tinder", "smoke cure"}, Tags: []string{"conifer", "dryland"}},
		{ID: "ash", Name: "Ash", BiomeTags: []string{"forest", "river", "mountain", "temperate"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.7, HeatFactor: 1.12, BurnFactor: 1.06, SparkEase: 2, Hardness: 4, Structural: 5, ResinQuality: 0.12, RotResistance: 4, SmokeFactor: 0.65, BarkResource: "bast_strip", BarkUses: []string{"tool handles", "splints"}, Tags: []string{"hardwood"}},
		{ID: "elm", Name: "Elm", BiomeTags: []string{"forest", "river", "temperate", "coast"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.6, HeatFactor: 1.07, BurnFactor: 1.04, SparkEase: 2, Hardness: 4, Structural: 5, ResinQuality: 0.1, RotResistance: 4, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"cordage", "basket frame"}, Tags: []string{"hardwood"}},
		{ID: "hickory", Name: "Hickory", BiomeTags: []string{"forest", "mountain", "temperate", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.9, GatherMaxKg: 5.1, HeatFactor: 1.2, BurnFactor: 1.17, SparkEase: 2, Hardness: 5, Structural: 5, ResinQuality: 0.1, RotResistance: 4, SmokeFactor: 0.65, BarkResource: "bast_strip", BarkUses: []string{"tool hafts", "smoke curing"}, Tags: []string{"hardwood"}},
		{ID: "poplar", Name: "Poplar", BiomeTags: []string{"forest", "river", "lake", "wetlands"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.7, GatherMaxKg: 3.6, HeatFactor: 0.8, BurnFactor: 0.75, SparkEase: 3, Hardness: 2, Structural: 2, ResinQuality: 0.1, RotResistance: 2, SmokeFactor: 0.8, BarkResource: "inner_bark_fiber", BarkUses: []string{"fiber", "kindling"}, Tags: []string{"hardwood_light"}},
		{ID: "sycamore", Name: "Sycamore", BiomeTags: []string{"river", "wetlands", "forest", "coast"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.4, HeatFactor: 1.03, BurnFactor: 0.98, SparkEase: 2, Hardness: 3, Structural: 4, ResinQuality: 0.08, RotResistance: 3, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"weaving", "lining"}, Tags: []string{"wetland"}},
		{ID: "aspen", Name: "Aspen", BiomeTags: []string{"boreal", "subarctic", "forest", "mountain"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 3.3, HeatFactor: 0.8, BurnFactor: 0.76, SparkEase: 3, Hardness: 2, Structural: 2, ResinQuality: 0.08, RotResistance: 2, SmokeFactor: 0.85, BarkResource: "inner_bark_fiber", BarkUses: []string{"tinder", "inner bark"}, Tags: []string{"boreal"}},
		{ID: "alder", Name: "Alder", BiomeTags: []string{"coast", "river", "wetlands", "forest"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.8, HeatFactor: 0.95, BurnFactor: 0.9, SparkEase: 2, Hardness: 3, Structural: 3, ResinQuality: 0.12, RotResistance: 3, SmokeFactor: 0.9, BarkResource: "bast_strip", BarkUses: []string{"smoke curing", "dye"}, Tags: []string{"wetland"}},
		{ID: "cottonwood", Name: "Cottonwood", BiomeTags: []string{"river", "delta", "wetlands", "coast"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.7, GatherMaxKg: 3.5, HeatFactor: 0.76, BurnFactor: 0.72, SparkEase: 3, Hardness: 2, Structural: 2, ResinQuality: 0.1, RotResistance: 2, SmokeFactor: 0.9, BarkResource: "inner_bark_fiber", BarkUses: []string{"cordage", "kindling"}, Tags: []string{"wetland"}},
		{ID: "beech", Name: "Beech", BiomeTags: []string{"forest", "mountain", "temperate", "coast"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.8, HeatFactor: 1.12, BurnFactor: 1.09, SparkEase: 2, Hardness: 4, Structural: 4, ResinQuality: 0.08, RotResistance: 3, SmokeFactor: 0.65, BarkResource: "bast_strip", BarkUses: []string{"bark strips", "containers"}, Tags: []string{"hardwood"}},
		{ID: "walnut", Name: "Walnut", BiomeTags: []string{"forest", "river", "temperate", "mountain"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.8, HeatFactor: 1.14, BurnFactor: 1.11, SparkEase: 2, Hardness: 5, Structural: 4, ResinQuality: 0.07, RotResistance: 4, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"bow staves", "tools"}, Tags: []string{"hardwood"}},
		{ID: "teak", Name: "Teak", BiomeTags: []string{"jungle", "tropical", "wetlands", "coast"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.9, GatherMaxKg: 5.2, HeatFactor: 1.1, BurnFactor: 1.02, SparkEase: 2, Hardness: 5, Structural: 5, ResinQuality: 0.15, RotResistance: 5, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"planking", "rafts"}, Tags: []string{"tropical", "hardwood"}},
		{ID: "mahogany", Name: "Mahogany", BiomeTags: []string{"jungle", "tropical", "island", "coast"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.9, GatherMaxKg: 5.1, HeatFactor: 1.09, BurnFactor: 1.03, SparkEase: 2, Hardness: 4, Structural: 5, ResinQuality: 0.1, RotResistance: 5, SmokeFactor: 0.7, BarkResource: "bast_strip", BarkUses: []string{"frames", "paddles"}, Tags: []string{"tropical", "hardwood"}},
		{ID: "kapok", Name: "Kapok", BiomeTags: []string{"jungle", "wetlands", "tropical", "delta"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.7, GatherMaxKg: 3.6, HeatFactor: 0.78, BurnFactor: 0.73, SparkEase: 3, Hardness: 2, Structural: 2, ResinQuality: 0.06, RotResistance: 2, SmokeFactor: 0.85, BarkResource: "inner_bark_fiber", BarkUses: []string{"fiber floss", "cordage"}, Tags: []string{"tropical"}},
		{ID: "ironwood", Name: "Ironwood", BiomeTags: []string{"island", "coast", "tropical", "dry"}, WoodType: WoodTypeHardwood, GatherMinKg: 1.0, GatherMaxKg: 5.6, HeatFactor: 1.24, BurnFactor: 1.2, SparkEase: 1, Hardness: 5, Structural: 5, ResinQuality: 0.06, RotResistance: 5, SmokeFactor: 0.6, BarkResource: "bast_strip", BarkUses: []string{"mallets", "wedges"}, Tags: []string{"hardwood", "tropical"}},
		{ID: "eucalyptus", Name: "Eucalyptus", BiomeTags: []string{"savanna", "dry", "coast", "badlands"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.3, HeatFactor: 1.0, BurnFactor: 0.92, SparkEase: 4, Hardness: 3, Structural: 4, ResinQuality: 0.55, RotResistance: 4, SmokeFactor: 1.15, BarkResource: "bast_strip", BarkUses: []string{"medicine leaves", "fuel"}, Tags: []string{"dryland"}},
		{ID: "casuarina", Name: "Casuarina", BiomeTags: []string{"coast", "island", "dry", "savanna"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.8, GatherMaxKg: 4.2, HeatFactor: 1.13, BurnFactor: 1.07, SparkEase: 2, Hardness: 4, Structural: 4, ResinQuality: 0.12, RotResistance: 4, SmokeFactor: 0.75, BarkResource: "bast_strip", BarkUses: []string{"posts", "stakes"}, Tags: []string{"coastal"}},
		{ID: "olive", Name: "Wild Olive", BiomeTags: []string{"coast", "dry", "badlands", "savanna"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.7, GatherMaxKg: 3.6, HeatFactor: 1.16, BurnFactor: 1.1, SparkEase: 2, Hardness: 5, Structural: 4, ResinQuality: 0.1, RotResistance: 5, SmokeFactor: 0.68, BarkResource: "bast_strip", BarkUses: []string{"tool handles", "stakes"}, Tags: []string{"dryland"}},
		{ID: "tamarisk", Name: "Tamarisk", BiomeTags: []string{"desert", "dry", "delta", "badlands"}, WoodType: WoodTypeHardwood, GatherMinKg: 0.6, GatherMaxKg: 3.2, HeatFactor: 0.98, BurnFactor: 0.94, SparkEase: 3, Hardness: 3, Structural: 3, ResinQuality: 0.18, RotResistance: 3, SmokeFactor: 0.9, BarkResource: "bast_strip", BarkUses: []string{"fencing", "twigs"}, Tags: []string{"dryland", "wetland_edge"}},
		{ID: "saguaro_rib", Name: "Saguaro Rib", BiomeTags: []string{"desert", "dry", "badlands"}, WoodType: WoodTypeDriftwood, GatherMinKg: 0.3, GatherMaxKg: 1.8, HeatFactor: 0.64, BurnFactor: 0.6, SparkEase: 3, Hardness: 2, Structural: 1, ResinQuality: 0.02, RotResistance: 1, SmokeFactor: 1.0, BarkUses: []string{"framework", "splints"}, Tags: []string{"desert"}},
		{ID: "black_spruce", Name: "Black Spruce", BiomeTags: []string{"boreal", "subarctic", "swamp", "tundra"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.0, HeatFactor: 0.93, BurnFactor: 0.86, SparkEase: 4, Hardness: 3, Structural: 3, ResinQuality: 0.78, RotResistance: 4, SmokeFactor: 1.05, BarkResource: "spruce_root", BarkUses: []string{"sewing root", "pitch"}, Tags: []string{"conifer", "wetland"}},
		{ID: "tamarack", Name: "Tamarack", BiomeTags: []string{"boreal", "wetlands", "subarctic", "mountain"}, WoodType: WoodTypeResinous, GatherMinKg: 0.8, GatherMaxKg: 4.2, HeatFactor: 1.0, BurnFactor: 0.94, SparkEase: 4, Hardness: 3, Structural: 4, ResinQuality: 0.72, RotResistance: 5, SmokeFactor: 0.95, BarkResource: "inner_bark_fiber", BarkUses: []string{"lashing", "pitch"}, Tags: []string{"conifer", "wetland"}},
		{ID: "paperbark", Name: "Paperbark", BiomeTags: []string{"swamp", "wetlands", "coast", "tropical"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.6, GatherMaxKg: 3.0, HeatFactor: 0.78, BurnFactor: 0.72, SparkEase: 4, Hardness: 2, Structural: 2, ResinQuality: 0.25, RotResistance: 3, SmokeFactor: 0.9, BarkResource: "birch_bark", BarkUses: []string{"sheet bark", "containers"}, Tags: []string{"wetland", "tropical"}},
		{ID: "redwood", Name: "Redwood", BiomeTags: []string{"coast", "temperate_rainforest", "forest"}, WoodType: WoodTypeSoftwood, GatherMinKg: 0.9, GatherMaxKg: 5.2, HeatFactor: 0.96, BurnFactor: 0.9, SparkEase: 3, Hardness: 3, Structural: 5, ResinQuality: 0.3, RotResistance: 5, SmokeFactor: 0.85, BarkResource: "cedar_bark", BarkUses: []string{"insulation", "roofing"}, Tags: []string{"conifer", "coastal"}},
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

func (s *RunState) addWoodStockWithWetness(woodType WoodType, kg float64, wetness float64) error {
	if s == nil || kg <= 0 {
		return nil
	}
	if !s.canStoreAtCamp(kg) {
		return fmt.Errorf("camp inventory full (%.1f/%.1fkg)", s.campUsedKg(), s.campCapacityKg())
	}
	wetness = clampFloat(wetness, 0, 1)
	for i := range s.WoodStock {
		if s.WoodStock[i].Type == woodType {
			total := s.WoodStock[i].Kg + kg
			if total > 0 {
				s.WoodStock[i].Wetness = ((s.WoodStock[i].Wetness * s.WoodStock[i].Kg) + (wetness * kg)) / total
			}
			s.WoodStock[i].Kg += kg
			return nil
		}
	}
	s.WoodStock = append(s.WoodStock, WoodStock{Type: woodType, Kg: kg, Wetness: wetness})
	return nil
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
	applySkillEffort(&player.Gathering, int(math.Ceil(dried*8)), true)
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
	player, ok := s.playerByID(playerID)
	if !ok {
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
	applySkillEffort(&player.Gathering, int(math.Round(kg*10)), true)
	bonusPct := float64(player.Gathering)/100.0*0.2 + float64(player.Strength+player.Agility)*0.03 + float64(sumTraitModifier(player.Traits))*0.01
	if bonusPct != 0 {
		kg = math.Max(0.2, kg*(1.0+bonusPct))
	}
	if err := s.addWoodStockWithWetness(tree.WoodType, kg, s.ambientWoodWetness()); err != nil {
		return TreeSpec{}, 0, err
	}
	return tree, kg, nil
}

type BarkStripResult struct {
	Tree        TreeSpec
	Primary     ResourceSpec
	PrimaryQty  float64
	FiberQty    float64
	HoursSpent  float64
	Quality     CraftQuality
	QualityText string
}

func (s *RunState) StripBark(playerID int, treeID string, requestedQty float64) (BarkStripResult, error) {
	if s == nil {
		return BarkStripResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return BarkStripResult{}, fmt.Errorf("player %d not found", playerID)
	}
	trees := TreesForBiome(s.Scenario.Biome)
	if len(trees) == 0 {
		return BarkStripResult{}, fmt.Errorf("no trees available")
	}

	treeID = strings.ToLower(strings.TrimSpace(treeID))
	tree := trees[0]
	if treeID == "" || treeID == "any" {
		rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("stripbark:%s:%d:%d", normalizeBiome(s.Scenario.Biome), s.Day, playerID)))
		tree = trees[rng.IntN(len(trees))]
	} else {
		found := false
		for _, option := range trees {
			if option.ID == treeID {
				tree = option
				found = true
				break
			}
		}
		if !found {
			return BarkStripResult{}, fmt.Errorf("tree not available in biome: %s", treeID)
		}
	}

	primaryID := strings.TrimSpace(tree.BarkResource)
	if primaryID == "" {
		primaryID = "inner_bark_fiber"
		switch tree.ID {
		case "birch":
			primaryID = "birch_bark"
		case "cedar":
			primaryID = "cedar_bark"
		case "willow":
			primaryID = "willow_bark"
		case "palm":
			primaryID = "bast_strip"
		case "spruce":
			primaryID = "spruce_root"
		}
	}
	primary, ok := s.findResourceForBiome(primaryID)
	if !ok {
		return BarkStripResult{}, fmt.Errorf("resource not registered: %s", primaryID)
	}
	fiber, ok := s.findResourceForBiome("inner_bark_fiber")
	if !ok {
		return BarkStripResult{}, fmt.Errorf("resource not registered: inner_bark_fiber")
	}

	if requestedQty <= 0 {
		requestedQty = 1
	}
	if requestedQty > 6 {
		requestedQty = 6
	}
	primaryQty := requestedQty
	fiberQty := max(1, int(math.Round(float64(requestedQty)*0.8)))
	if primary.Unit == "kg" {
		primaryQty = float64(math.Round(float64(requestedQty)*10) / 10)
	}

	if err := s.addResourceStock(primary, primaryQty); err != nil {
		return BarkStripResult{}, err
	}
	if err := s.addResourceStock(fiber, float64(fiberQty)); err != nil {
		_ = s.consumeResourceStock(primary.ID, primaryQty)
		return BarkStripResult{}, err
	}

	score := float64(player.Bushcraft+player.Agility+player.Crafting/25) + float64(sumTraitModifier(player.Traits))/2
	quality := qualityFromScore(score)
	applySkillEffort(&player.Gathering, int(math.Round(float64(primaryQty)*10)), true)
	applySkillEffort(&player.Crafting, int(math.Round(float64(fiberQty)*8)), true)
	hours := clampFloat(0.35+(float64(requestedQty)*0.18)-qualityTimeReduction(quality), 0.2, 2.5)
	player.Energy = clamp(player.Energy-int(math.Ceil(hours*3)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*2)), 0, 100)
	refreshEffectBars(player)

	return BarkStripResult{
		Tree:        tree,
		Primary:     primary,
		PrimaryQty:  primaryQty,
		FiberQty:    float64(fiberQty),
		HoursSpent:  hours,
		Quality:     quality,
		QualityText: string(quality),
	}, nil
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
	ShelterLeanTo           ShelterType = "lean_to"
	ShelterDebrisHut        ShelterType = "debris_hut"
	ShelterTarpAFrame       ShelterType = "tarp_a_frame"
	ShelterSnowCave         ShelterType = "snow_cave"
	ShelterDesertShade      ShelterType = "desert_shade"
	ShelterSwampPlatform    ShelterType = "swamp_platform"
	ShelterBambooHut        ShelterType = "bamboo_hut"
	ShelterRockOverhang     ShelterType = "rock_overhang"
	ShelterAFrame           ShelterType = "a_frame"
	ShelterWattleDaub       ShelterType = "wattle_daub_hut"
	ShelterRaisedPlatform   ShelterType = "raised_platform_shelter"
	ShelterEarthDugout      ShelterType = "earth_sheltered_dugout"
	ShelterLogCabin         ShelterType = "log_cabin"
	ShelterQuinzee          ShelterType = "quinzee"
	ShelterHuntingBlind     ShelterType = "hunting_blind"
	ShelterElevatedCachePod ShelterType = "elevated_cache_pod"
)

type ShelterStageSpec struct {
	ID                 string
	Name               string
	BuildHours         float64
	BuildEnergyCost    int
	BuildHydrationCost int
	BuildMoraleBonus   int
	StorageCapacityKg  float64
	Insulation         int
	RainProtection     int
	WindProtection     int
	InsectProtection   int
	PredatorSafety     int
	Comfort            int
	DrynessProtection  int
	Stealth            int
	DurabilityBonus    int
	RequiresItems      []string
	RequiresResources  []ResourceRequirement
}

type ShelterSpec struct {
	ID                 ShelterType
	Name               string
	BiomeTags          []string
	StorageCapacityKg  float64
	Insulation         int
	RainProtection     int
	WindProtection     int
	InsectProtection   int
	DurabilityPerDay   int
	BuildMoraleBonus   int
	BuildEnergyCost    int
	BuildHydrationCost int
	PredatorSafety     int
	Comfort            int
	DrynessProtection  int
	Stealth            int
	Maintenance        int
	SleepShelter       bool
	UpgradeComponents  []string
	Stages             []ShelterStageSpec
}

func ShelterCatalog() []ShelterSpec {
	return []ShelterSpec{
		{
			ID: ShelterDebrisHut, Name: "Debris Hut", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain"},
			StorageCapacityKg: 34, Insulation: 5, RainProtection: 4, WindProtection: 4, InsectProtection: 2, PredatorSafety: 2, Comfort: 3, DrynessProtection: 4, Stealth: 4,
			DurabilityPerDay: 5, Maintenance: 4, BuildMoraleBonus: 3, BuildEnergyCost: 6, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"bough_mattress", "storm_flap", "drainage_ditch", "reflective_fire_wall"},
			Stages: []ShelterStageSpec{
				{ID: "frame", Name: "Frame", BuildHours: 2.3, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, StorageCapacityKg: 8, WindProtection: 1, Stealth: 1, DurabilityBonus: 15, RequiresResources: []ResourceRequirement{{ID: "cane_stalk", Qty: 1}}},
				{ID: "insulation", Name: "Leaf Insulation", BuildHours: 2.4, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, Insulation: 2, WindProtection: 1, DrynessProtection: 1, StorageCapacityKg: 10, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "dry_leaf_litter", Qty: 2}}},
				{ID: "weatherproofing", Name: "Weatherproofing", BuildHours: 2.0, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, RainProtection: 3, InsectProtection: 1, StorageCapacityKg: 16, DurabilityBonus: 20, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}},
			},
		},
		{
			ID: ShelterLeanTo, Name: "Lean-to", BiomeTags: []string{"forest", "mountain", "boreal", "coast"},
			StorageCapacityKg: 24, Insulation: 3, RainProtection: 3, WindProtection: 3, InsectProtection: 1, PredatorSafety: 1, Comfort: 2, DrynessProtection: 2, Stealth: 2,
			DurabilityPerDay: 6, Maintenance: 4, BuildMoraleBonus: 2, BuildEnergyCost: 4, BuildHydrationCost: 2, SleepShelter: true,
			UpgradeComponents: []string{"storm_flap", "drainage_ditch", "raised_sleeping_platform"},
			Stages: []ShelterStageSpec{
				{ID: "wall", Name: "Back Wall", BuildHours: 1.7, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, WindProtection: 2, StorageCapacityKg: 8, DurabilityBonus: 14},
				{ID: "roof", Name: "Roofing", BuildHours: 1.6, BuildEnergyCost: 2, BuildHydrationCost: 1, BuildMoraleBonus: 1, RainProtection: 2, StorageCapacityKg: 8, DurabilityBonus: 16, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}},
				{ID: "insulation", Name: "Insulation Layer", BuildHours: 1.4, BuildEnergyCost: 2, BuildHydrationCost: 1, BuildMoraleBonus: 1, Insulation: 2, Comfort: 2, StorageCapacityKg: 8, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "spruce_bough", Qty: 1}}},
			},
		},
		{
			ID: ShelterAFrame, Name: "A-Frame Shelter", BiomeTags: []string{"forest", "mountain", "boreal", "coast", "wetlands"},
			StorageCapacityKg: 30, Insulation: 4, RainProtection: 4, WindProtection: 4, InsectProtection: 2, PredatorSafety: 2, Comfort: 3, DrynessProtection: 4, Stealth: 3,
			DurabilityPerDay: 5, Maintenance: 5, BuildMoraleBonus: 3, BuildEnergyCost: 5, BuildHydrationCost: 2, SleepShelter: true,
			UpgradeComponents: []string{"insulated_wall_lining", "storm_flap", "storage_shelves", "reflective_fire_wall"},
			Stages: []ShelterStageSpec{
				{ID: "ridge", Name: "Ridge Pole", BuildHours: 1.6, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, WindProtection: 1, DurabilityBonus: 14, RequiresItems: []string{"ridge_pole_kit"}},
				{ID: "roof", Name: "Roof Panels", BuildHours: 2.0, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, RainProtection: 2, StorageCapacityKg: 10, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}},
				{ID: "insulate", Name: "Insulation", BuildHours: 1.8, BuildEnergyCost: 2, BuildHydrationCost: 1, BuildMoraleBonus: 1, Insulation: 3, Comfort: 2, DrynessProtection: 2, StorageCapacityKg: 20, DurabilityBonus: 20, RequiresResources: []ResourceRequirement{{ID: "dry_leaf_litter", Qty: 1}}},
			},
		},
		{
			ID: ShelterWattleDaub, Name: "Wattle & Daub Hut", BiomeTags: []string{"wetlands", "swamp", "river", "forest", "delta"},
			StorageCapacityKg: 44, Insulation: 5, RainProtection: 5, WindProtection: 5, InsectProtection: 3, PredatorSafety: 3, Comfort: 4, DrynessProtection: 5, Stealth: 2,
			DurabilityPerDay: 3, Maintenance: 3, BuildMoraleBonus: 4, BuildEnergyCost: 7, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"storm_flap", "stone_hearth", "smoke_hole_baffle", "storage_shelves"},
			Stages: []ShelterStageSpec{
				{ID: "wattle", Name: "Wattle Frame", BuildHours: 3.0, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, WindProtection: 2, StorageCapacityKg: 10, DurabilityBonus: 14, RequiresResources: []ResourceRequirement{{ID: "willow_withy", Qty: 2}}},
				{ID: "daub", Name: "Daub Layer", BuildHours: 3.5, BuildEnergyCost: 5, BuildHydrationCost: 2, BuildMoraleBonus: 2, RainProtection: 3, DrynessProtection: 2, StorageCapacityKg: 14, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "mud", Qty: 1.5}}},
				{ID: "roof", Name: "Roof Pack", BuildHours: 2.8, BuildEnergyCost: 4, BuildHydrationCost: 1, BuildMoraleBonus: 1, Insulation: 3, Comfort: 2, StorageCapacityKg: 20, DurabilityBonus: 20, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 2}}},
			},
		},
		{
			ID: ShelterRaisedPlatform, Name: "Raised Platform Shelter", BiomeTags: []string{"swamp", "wetlands", "delta", "jungle", "coast"},
			StorageCapacityKg: 40, Insulation: 3, RainProtection: 4, WindProtection: 3, InsectProtection: 6, PredatorSafety: 4, Comfort: 4, DrynessProtection: 6, Stealth: 2,
			DurabilityPerDay: 4, Maintenance: 4, BuildMoraleBonus: 4, BuildEnergyCost: 7, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"bough_mattress", "storm_flap", "elevated_food_cache", "drainage_ditch"},
			Stages: []ShelterStageSpec{
				{ID: "piles", Name: "Pile Foundation", BuildHours: 2.8, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, DrynessProtection: 2, PredatorSafety: 1, StorageCapacityKg: 8, DurabilityBonus: 14, RequiresResources: []ResourceRequirement{{ID: "drift_log", Qty: 1}}},
				{ID: "deck", Name: "Deck Platform", BuildHours: 2.4, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, Comfort: 2, InsectProtection: 2, StorageCapacityKg: 12, DurabilityBonus: 16, RequiresResources: []ResourceRequirement{{ID: "cane_stalk", Qty: 1}}},
				{ID: "roof", Name: "Roof & Screens", BuildHours: 2.5, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, RainProtection: 3, InsectProtection: 3, DrynessProtection: 2, StorageCapacityKg: 20, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "palm_frond", Qty: 1}}},
			},
		},
		{
			ID: ShelterEarthDugout, Name: "Earth-Sheltered Dugout", BiomeTags: []string{"forest", "mountain", "badlands", "tundra", "boreal"},
			StorageCapacityKg: 52, Insulation: 6, RainProtection: 4, WindProtection: 6, InsectProtection: 3, PredatorSafety: 4, Comfort: 4, DrynessProtection: 4, Stealth: 5,
			DurabilityPerDay: 3, Maintenance: 3, BuildMoraleBonus: 4, BuildEnergyCost: 8, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"stone_hearth", "smoke_hole_baffle", "drainage_ditch", "storage_shelves"},
			Stages: []ShelterStageSpec{
				{ID: "excavate", Name: "Excavation", BuildHours: 3.4, BuildEnergyCost: 6, BuildHydrationCost: 2, BuildMoraleBonus: 1, WindProtection: 2, Stealth: 2, DurabilityBonus: 14, RequiresItems: []string{"digging_stick"}},
				{ID: "timber", Name: "Timber Lining", BuildHours: 3.0, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, Insulation: 2, DrynessProtection: 2, StorageCapacityKg: 16, DurabilityBonus: 18, RequiresResources: []ResourceRequirement{{ID: "gravel", Qty: 1.0}}},
				{ID: "cap", Name: "Roof Cap", BuildHours: 2.7, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 2, RainProtection: 3, Comfort: 2, StorageCapacityKg: 20, DurabilityBonus: 20, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}},
			},
		},
		{
			ID: ShelterLogCabin, Name: "Log Cabin", BiomeTags: []string{"forest", "boreal", "mountain", "subarctic"},
			StorageCapacityKg: 72, Insulation: 7, RainProtection: 7, WindProtection: 7, InsectProtection: 5, PredatorSafety: 6, Comfort: 6, DrynessProtection: 6, Stealth: 2,
			DurabilityPerDay: 2, Maintenance: 2, BuildMoraleBonus: 5, BuildEnergyCost: 10, BuildHydrationCost: 4, SleepShelter: true,
			UpgradeComponents: []string{"stone_hearth", "smoke_hole_baffle", "storage_shelves", "elevated_food_cache", "door_latch"},
			Stages: []ShelterStageSpec{
				{ID: "foundation", Name: "Foundation", BuildHours: 4.0, BuildEnergyCost: 6, BuildHydrationCost: 2, BuildMoraleBonus: 1, WindProtection: 1, DrynessProtection: 1, DurabilityBonus: 10, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 2}}},
				{ID: "walls", Name: "Wall Courses", BuildHours: 6.0, BuildEnergyCost: 8, BuildHydrationCost: 3, BuildMoraleBonus: 1, WindProtection: 2, Insulation: 2, StorageCapacityKg: 12, DurabilityBonus: 14},
				{ID: "roof", Name: "Roof", BuildHours: 5.0, BuildEnergyCost: 7, BuildHydrationCost: 2, BuildMoraleBonus: 1, RainProtection: 3, DrynessProtection: 2, StorageCapacityKg: 12, DurabilityBonus: 15, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 2}}},
				{ID: "chinking", Name: "Chinking", BuildHours: 3.0, BuildEnergyCost: 4, BuildHydrationCost: 1, BuildMoraleBonus: 1, Insulation: 2, WindProtection: 1, DurabilityBonus: 16, RequiresResources: []ResourceRequirement{{ID: "mud", Qty: 1.2}}},
				{ID: "hearth", Name: "Hearth", BuildHours: 2.5, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, Comfort: 2, PredatorSafety: 1, StorageCapacityKg: 12, DurabilityBonus: 14, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 1}}},
				{ID: "door", Name: "Door & Hatch", BuildHours: 2.0, BuildEnergyCost: 2, BuildHydrationCost: 1, BuildMoraleBonus: 1, RainProtection: 1, InsectProtection: 2, StorageCapacityKg: 24, DurabilityBonus: 16},
			},
		},
		{
			ID: ShelterSnowCave, Name: "Snow Cave", BiomeTags: []string{"arctic", "subarctic", "tundra", "winter"},
			StorageCapacityKg: 26, Insulation: 6, RainProtection: 3, WindProtection: 6, InsectProtection: 3, PredatorSafety: 3, Comfort: 3, DrynessProtection: 3, Stealth: 4,
			DurabilityPerDay: 7, Maintenance: 6, BuildMoraleBonus: 2, BuildEnergyCost: 8, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"bough_mattress", "cold_air_trench"},
			Stages: []ShelterStageSpec{
				{ID: "drift", Name: "Snow Drift", BuildHours: 1.5, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 0, WindProtection: 2, DurabilityBonus: 12},
				{ID: "hollow", Name: "Hollow Chamber", BuildHours: 2.2, BuildEnergyCost: 4, BuildHydrationCost: 2, BuildMoraleBonus: 1, Insulation: 3, PredatorSafety: 1, DurabilityBonus: 15},
				{ID: "vent", Name: "Vent & Entrance Baffle", BuildHours: 1.0, BuildEnergyCost: 1, BuildHydrationCost: 1, BuildMoraleBonus: 1, Comfort: 1, DrynessProtection: 1, DurabilityBonus: 10},
			},
		},
		{
			ID: ShelterQuinzee, Name: "Quinzee", BiomeTags: []string{"arctic", "subarctic", "tundra", "winter", "boreal"},
			StorageCapacityKg: 30, Insulation: 7, RainProtection: 3, WindProtection: 7, InsectProtection: 3, PredatorSafety: 3, Comfort: 3, DrynessProtection: 3, Stealth: 4,
			DurabilityPerDay: 6, Maintenance: 5, BuildMoraleBonus: 3, BuildEnergyCost: 8, BuildHydrationCost: 3, SleepShelter: true,
			UpgradeComponents: []string{"cold_air_trench", "bough_mattress"},
			Stages: []ShelterStageSpec{
				{ID: "mound", Name: "Snow Mound", BuildHours: 2.2, BuildEnergyCost: 3, BuildHydrationCost: 1, BuildMoraleBonus: 1, WindProtection: 2, DurabilityBonus: 10},
				{ID: "set", Name: "Set & Cure", BuildHours: 1.2, BuildEnergyCost: 1, BuildHydrationCost: 1, BuildMoraleBonus: 0, DurabilityBonus: 8},
				{ID: "hollow", Name: "Hollow Interior", BuildHours: 2.3, BuildEnergyCost: 4, BuildHydrationCost: 1, BuildMoraleBonus: 1, Insulation: 4, Comfort: 2, DurabilityBonus: 14},
			},
		},
		{
			ID: ShelterDesertShade, Name: "Desert Shade", BiomeTags: []string{"desert", "dry", "savanna", "badlands"},
			StorageCapacityKg: 18, Insulation: 1, RainProtection: 1, WindProtection: 2, InsectProtection: 2, PredatorSafety: 1, Comfort: 2, DrynessProtection: 1, Stealth: 2,
			DurabilityPerDay: 4, Maintenance: 4, BuildMoraleBonus: 2, BuildEnergyCost: 3, BuildHydrationCost: 2, SleepShelter: true,
			UpgradeComponents: []string{"storm_flap", "reflective_fire_wall"},
		},
		{
			ID: ShelterSwampPlatform, Name: "Swamp Platform", BiomeTags: []string{"swamp", "wetlands", "delta", "jungle"},
			StorageCapacityKg: 28, Insulation: 2, RainProtection: 2, WindProtection: 2, InsectProtection: 5, PredatorSafety: 3, Comfort: 3, DrynessProtection: 5, Stealth: 2,
			DurabilityPerDay: 5, Maintenance: 4, BuildMoraleBonus: 3, BuildEnergyCost: 6, BuildHydrationCost: 3, SleepShelter: true,
		},
		{
			ID: ShelterBambooHut, Name: "Bamboo Hut", BiomeTags: []string{"jungle", "tropical", "wetlands", "island"},
			StorageCapacityKg: 40, Insulation: 3, RainProtection: 4, WindProtection: 3, InsectProtection: 4, PredatorSafety: 3, Comfort: 4, DrynessProtection: 4, Stealth: 2,
			DurabilityPerDay: 4, Maintenance: 3, BuildMoraleBonus: 4, BuildEnergyCost: 6, BuildHydrationCost: 3, SleepShelter: true,
		},
		{
			ID: ShelterRockOverhang, Name: "Rock Overhang", BiomeTags: []string{"mountain", "badlands", "desert", "coast"},
			StorageCapacityKg: 24, Insulation: 3, RainProtection: 2, WindProtection: 4, InsectProtection: 1, PredatorSafety: 2, Comfort: 2, DrynessProtection: 2, Stealth: 3,
			DurabilityPerDay: 3, Maintenance: 2, BuildMoraleBonus: 2, BuildEnergyCost: 2, BuildHydrationCost: 1, SleepShelter: true,
		},
		{
			ID: ShelterHuntingBlind, Name: "Hunting Blind", BiomeTags: []string{"forest", "savanna", "wetlands", "badlands", "mountain"},
			StorageCapacityKg: 12, Insulation: 0, RainProtection: 1, WindProtection: 2, InsectProtection: 1, PredatorSafety: 0, Comfort: 1, DrynessProtection: 1, Stealth: 6,
			DurabilityPerDay: 6, Maintenance: 5, BuildMoraleBonus: 1, BuildEnergyCost: 3, BuildHydrationCost: 1, SleepShelter: false,
			UpgradeComponents: []string{"camouflage_screen", "lookout_platform"},
			Stages: []ShelterStageSpec{
				{ID: "hide", Name: "Concealment Frame", BuildHours: 1.4, BuildEnergyCost: 2, BuildHydrationCost: 1, BuildMoraleBonus: 0, WindProtection: 1, Stealth: 3, DurabilityBonus: 12},
				{ID: "screen", Name: "Camouflage Screen", BuildHours: 1.3, BuildEnergyCost: 1, BuildHydrationCost: 1, BuildMoraleBonus: 1, Stealth: 3, RainProtection: 1, DurabilityBonus: 10},
			},
		},
		{
			ID: ShelterElevatedCachePod, Name: "Elevated Cache Pod", BiomeTags: []string{"forest", "wetlands", "swamp", "tundra", "boreal"},
			StorageCapacityKg: 26, Insulation: 0, RainProtection: 2, WindProtection: 2, InsectProtection: 2, PredatorSafety: 5, Comfort: 0, DrynessProtection: 3, Stealth: 3,
			DurabilityPerDay: 4, Maintenance: 3, BuildMoraleBonus: 2, BuildEnergyCost: 3, BuildHydrationCost: 1, SleepShelter: false,
			UpgradeComponents: []string{"elevated_food_cache"},
		},
		{
			ID: ShelterTarpAFrame, Name: "Tarp A-Frame", BiomeTags: []string{"forest", "coast", "wetlands", "jungle", "mountain"},
			StorageCapacityKg: 30, Insulation: 3, RainProtection: 4, WindProtection: 3, InsectProtection: 2, PredatorSafety: 2, Comfort: 3, DrynessProtection: 4, Stealth: 2,
			DurabilityPerDay: 4, Maintenance: 4, BuildMoraleBonus: 3, BuildEnergyCost: 3, BuildHydrationCost: 2, SleepShelter: true,
		},
	}
}

type ShelterState struct {
	Type       ShelterType `json:"type"`
	Durability int         `json:"durability"`
	BuiltDay   int         `json:"built_day"`
	Stage      int         `json:"stage,omitempty"`
	SiteX      int         `json:"site_x,omitempty"`
	SiteY      int         `json:"site_y,omitempty"`
	Upgrades   []string    `json:"upgrades,omitempty"`
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

	if s.Shelter.Type != "" && s.Shelter.Type != chosen.ID && s.Shelter.Durability > 0 {
		return ShelterSpec{}, fmt.Errorf("already maintaining %s; finish or replace deliberately", s.Shelter.Type)
	}

	stages := chosen.Stages
	if len(stages) == 0 {
		stages = defaultShelterStages(chosen)
	}
	nextStage := 1
	if s.Shelter.Type == chosen.ID {
		nextStage = s.Shelter.Stage + 1
	}
	if nextStage > len(stages) {
		return ShelterSpec{}, fmt.Errorf("%s is already fully built", chosen.Name)
	}
	stage := stages[nextStage-1]

	for _, req := range stage.RequiresItems {
		if !slices.Contains(s.CraftedItems, req) {
			return ShelterSpec{}, fmt.Errorf("requires crafted item: %s", req)
		}
	}
	if nextStage > 1 {
		for _, req := range stage.RequiresResources {
			if s.resourceQty(req.ID) < req.Qty {
				return ShelterSpec{}, fmt.Errorf("requires resource: %s %.1f", req.ID, req.Qty)
			}
		}
		for _, req := range stage.RequiresResources {
			_ = s.consumeResourceStock(req.ID, req.Qty)
		}
	} else {
		// Stage one uses immediately available on-site materials to keep first shelter access practical.
		for _, req := range stage.RequiresResources {
			if s.resourceQty(req.ID) >= req.Qty {
				_ = s.consumeResourceStock(req.ID, req.Qty)
			}
		}
	}

	if s.Shelter.Type != chosen.ID {
		s.Shelter = ShelterState{
			Type:       chosen.ID,
			Durability: 30,
			BuiltDay:   s.Day,
			Stage:      0,
			SiteX:      s.Travel.PosX,
			SiteY:      s.Travel.PosY,
		}
	}
	s.Shelter.Stage = nextStage
	s.Shelter.Durability = clamp(s.Shelter.Durability+stage.DurabilityBonus, 10, 100)

	energyCost := max(1, stage.BuildEnergyCost)
	hydrationCost := max(1, stage.BuildHydrationCost)
	moraleBonus := stage.BuildMoraleBonus
	if energyCost == 1 && chosen.BuildEnergyCost > 0 {
		energyCost = chosen.BuildEnergyCost
	}
	if hydrationCost == 1 && chosen.BuildHydrationCost > 0 {
		hydrationCost = chosen.BuildHydrationCost
	}
	if moraleBonus == 0 {
		moraleBonus = chosen.BuildMoraleBonus
	}
	player.Energy = clamp(player.Energy-energyCost, 0, 100)
	player.Hydration = clamp(player.Hydration-hydrationCost, 0, 100)
	player.Morale = clamp(player.Morale+moraleBonus, 0, 100)
	effortHours := stage.BuildHours
	if effortHours <= 0 {
		effortHours = 1.2
	}
	applySkillEffort(&player.Sheltercraft, int(math.Round(effortHours*16)), true)
	if stage.BuildHours > 0 {
		_ = s.AdvanceActionClock(stage.BuildHours)
	}
	refreshEffectBars(player)

	return chosen, nil
}

func defaultShelterStages(spec ShelterSpec) []ShelterStageSpec {
	return []ShelterStageSpec{
		{
			ID:                 "build",
			Name:               "Build",
			BuildHours:         2.0,
			BuildEnergyCost:    max(1, spec.BuildEnergyCost),
			BuildHydrationCost: max(1, spec.BuildHydrationCost),
			BuildMoraleBonus:   max(1, spec.BuildMoraleBonus),
			StorageCapacityKg:  spec.StorageCapacityKg,
			Insulation:         spec.Insulation,
			RainProtection:     spec.RainProtection,
			WindProtection:     spec.WindProtection,
			InsectProtection:   spec.InsectProtection,
			PredatorSafety:     spec.PredatorSafety,
			Comfort:            spec.Comfort,
			DrynessProtection:  spec.DrynessProtection,
			Stealth:            spec.Stealth,
			DurabilityBonus:    70,
		},
	}
}

type shelterMetrics struct {
	StorageCapacityKg float64
	Insulation        int
	RainProtection    int
	WindProtection    int
	InsectProtection  int
	DurabilityPerDay  int
	PredatorSafety    int
	Comfort           int
	DrynessProtection int
	Stealth           int
	Maintenance       int
}

func shelterMetricsForState(spec ShelterSpec, state ShelterState) shelterMetrics {
	metrics := shelterMetrics{
		StorageCapacityKg: spec.StorageCapacityKg,
		Insulation:        spec.Insulation,
		RainProtection:    spec.RainProtection,
		WindProtection:    spec.WindProtection,
		InsectProtection:  spec.InsectProtection,
		DurabilityPerDay:  spec.DurabilityPerDay,
		PredatorSafety:    spec.PredatorSafety,
		Comfort:           spec.Comfort,
		DrynessProtection: spec.DrynessProtection,
		Stealth:           spec.Stealth,
		Maintenance:       max(1, spec.Maintenance),
	}
	if state.Stage <= 0 {
		return metrics
	}
	stages := spec.Stages
	if len(stages) == 0 {
		stages = defaultShelterStages(spec)
	}
	metrics = shelterMetrics{
		DurabilityPerDay: max(1, spec.DurabilityPerDay),
		Maintenance:      max(1, spec.Maintenance),
	}
	for i := 0; i < min(state.Stage, len(stages)); i++ {
		stage := stages[i]
		metrics.StorageCapacityKg += stage.StorageCapacityKg
		metrics.Insulation += stage.Insulation
		metrics.RainProtection += stage.RainProtection
		metrics.WindProtection += stage.WindProtection
		metrics.InsectProtection += stage.InsectProtection
		metrics.PredatorSafety += stage.PredatorSafety
		metrics.Comfort += stage.Comfort
		metrics.DrynessProtection += stage.DrynessProtection
		metrics.Stealth += stage.Stealth
	}
	return metrics
}

func (s *RunState) shelterLocationModifier(state ShelterState) shelterMetrics {
	if s == nil {
		return shelterMetrics{}
	}
	x, y := state.SiteX, state.SiteY
	if x == 0 && y == 0 {
		x, y = s.Travel.PosX, s.Travel.PosY
	}
	cell, ok := s.TopologyCellAt(x, y)
	if !ok {
		return shelterMetrics{}
	}
	mod := shelterMetrics{}
	if cell.Elevation >= 45 {
		mod.Insulation -= 1
		mod.WindProtection -= 2
		mod.InsectProtection += 1
		mod.Stealth -= 1
	}
	if cell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake|TopoFlagCoast) != 0 || s.isNearWater(x, y) {
		mod.InsectProtection -= 2
		mod.DrynessProtection -= 2
		mod.Comfort -= 1
	}
	switch cell.Biome {
	case TopoBiomeForest, TopoBiomeBoreal, TopoBiomeJungle:
		mod.WindProtection += 1
		mod.Stealth += 2
		mod.DrynessProtection -= 1
	case TopoBiomeDesert, TopoBiomeGrassland, TopoBiomeTundra:
		mod.WindProtection -= 1
		mod.RainProtection -= 1
		mod.Stealth -= 1
	case TopoBiomeSwamp, TopoBiomeWetland:
		mod.InsectProtection -= 1
		mod.DrynessProtection -= 2
	}
	return mod
}

func (s *RunState) currentShelterMetrics() (shelterMetrics, bool) {
	if s == nil || s.Shelter.Type == "" || s.Shelter.Durability <= 0 {
		return shelterMetrics{}, false
	}
	spec, ok := shelterByID(s.Shelter.Type)
	if !ok {
		return shelterMetrics{}, false
	}
	metrics := shelterMetricsForState(spec, s.Shelter)
	loc := s.shelterLocationModifier(s.Shelter)
	metrics.Insulation += loc.Insulation
	metrics.RainProtection += loc.RainProtection
	metrics.WindProtection += loc.WindProtection
	metrics.InsectProtection += loc.InsectProtection
	metrics.PredatorSafety += loc.PredatorSafety
	metrics.Comfort += loc.Comfort
	metrics.DrynessProtection += loc.DrynessProtection
	metrics.Stealth += loc.Stealth
	metrics.StorageCapacityKg = maxFloat64(8.0, metrics.StorageCapacityKg+loc.StorageCapacityKg)
	metrics.DurabilityPerDay = max(1, metrics.DurabilityPerDay-loc.Maintenance/2)
	metrics = s.applyShelterUpgradeModifiers(metrics)
	return metrics, true
}

func (s *RunState) applyShelterUpgradeModifiers(metrics shelterMetrics) shelterMetrics {
	if s == nil {
		return metrics
	}
	installed := map[string]bool{}
	for _, id := range s.CraftedItems {
		installed[strings.TrimSpace(strings.ToLower(id))] = true
	}
	for _, id := range s.Shelter.Upgrades {
		installed[strings.TrimSpace(strings.ToLower(id))] = true
	}
	apply := func(id string, fn func()) {
		if installed[id] {
			fn()
		}
	}
	apply("raised_sleeping_platform", func() { metrics.Comfort += 2; metrics.DrynessProtection += 1 })
	apply("bough_mattress", func() { metrics.Comfort += 2; metrics.Insulation += 1 })
	apply("insulated_wall_lining", func() { metrics.Insulation += 2; metrics.WindProtection += 1 })
	apply("storm_flap", func() { metrics.RainProtection += 2; metrics.WindProtection += 1 })
	apply("drainage_ditch", func() { metrics.DrynessProtection += 2; metrics.DurabilityPerDay = max(1, metrics.DurabilityPerDay-1) })
	apply("reflective_fire_wall", func() { metrics.Insulation += 1; metrics.Comfort += 1 })
	apply("stone_hearth", func() { metrics.Comfort += 1; metrics.RainProtection += 1 })
	apply("smoke_hole_baffle", func() { metrics.WindProtection += 1; metrics.RainProtection += 1 })
	apply("storage_shelves", func() { metrics.StorageCapacityKg += 6 })
	apply("elevated_food_cache", func() { metrics.StorageCapacityKg += 8; metrics.PredatorSafety += 2 })
	apply("door_latch", func() { metrics.PredatorSafety += 1; metrics.Stealth += 1 })
	apply("camouflage_screen", func() { metrics.Stealth += 2 })
	apply("lookout_platform", func() { metrics.PredatorSafety += 1; metrics.WindProtection += 1 })
	apply("cold_air_trench", func() { metrics.Insulation += 1; metrics.DrynessProtection += 1 })
	return metrics
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
	applySkillEffort(&player.Firecraft, 12, true)
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
	applySkillEffort(&player.Firecraft, 8, true)
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
		applySkillEffort(&player.Firecraft, created*3, true)
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
		applySkillEffort(&player.Firecraft, created*3, true)
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
		applySkillEffort(&player.Firecraft, created*4, true)
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
	applySkillEffort(&player.Firecraft, 8, success)
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
		applySkillEffort(&player.Firecraft, 6, false)
		refreshEffectBars(player)
		return chance, false, nil
	}

	if err := s.startFireWithMethod(playerID, woodType, kg, FireMethodBowDrill); err != nil {
		return chance, false, err
	}
	applySkillEffort(&player.Firecraft, 12, true)
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
		metrics, ok := s.currentShelterMetrics()
		if !ok {
			metrics = shelterMetrics{DurabilityPerDay: 5}
		}
		loss := max(1, metrics.DurabilityPerDay)
		if isRainyWeather(s.Weather.Type) {
			loss += max(1, 2-metrics.DrynessProtection/3)
		}
		if isSevereWeather(s.Weather.Type) {
			loss += max(1, 3-metrics.WindProtection/3)
		}
		if metrics.Maintenance > 0 {
			loss = max(1, loss-metrics.Maintenance/5)
		}
		s.Shelter.Durability = clamp(s.Shelter.Durability-loss, 0, 100)
		if s.Shelter.Durability == 0 {
			s.Shelter = ShelterState{}
		}
	}

	// Prepared fire materials can degrade in persistent wet weather.
	if isRainyWeather(s.Weather.Type) || s.Weather.Type == WeatherSnow || s.Weather.Type == WeatherBlizzard {
		wetHit := 0.08
		if metrics, ok := s.currentShelterMetrics(); ok {
			wetHit *= clampFloat(0.45-float64(metrics.DrynessProtection)*0.02, 0.2, 0.45)
		}
		s.FirePrep.TinderQuality = clampFloat(s.FirePrep.TinderQuality-wetHit, 0, 1)
		s.FirePrep.KindlingQuality = clampFloat(s.FirePrep.KindlingQuality-wetHit*0.8, 0, 1)
		s.FirePrep.FeatherQuality = clampFloat(s.FirePrep.FeatherQuality-wetHit*0.7, 0, 1)
	}

	if !s.Fire.Lit {
		s.progressTrapsDaily()
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
		s.progressTrapsDaily()
		return
	}
	s.Fire.Intensity = clamp(int(float64(s.Fire.Intensity)*0.86), 8, 100)
	s.Fire.HeatC = clamp(int(float64(s.Fire.HeatC)*0.84), 0, 120)
	s.progressTrapsDaily()
}

func (s *RunState) campImpactForDay() statDelta {
	if s == nil {
		return statDelta{}
	}
	impact := statDelta{}

	if s.Shelter.Type != "" && s.Shelter.Durability > 0 {
		if shelter, ok := s.currentShelterMetrics(); ok {
			// Shelter blunts weather stress and bug pressure.
			impact.Energy += shelter.Insulation / 2
			impact.Morale += shelter.Comfort/2 + shelter.RainProtection/3
			if isRainyWeather(s.Weather.Type) {
				impact.Energy += shelter.RainProtection/2 + shelter.WindProtection/3
				impact.Morale += shelter.DrynessProtection/2 + shelter.RainProtection/3
			}
			if biomeIsTropicalWet(s.Scenario.Biome) {
				impact.Morale += shelter.InsectProtection / 2
			}
			if s.Weather.TemperatureC <= 0 {
				impact.Energy += shelter.Insulation / 3
			}
			impact.Morale += shelter.PredatorSafety / 3
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
	Category          string
	MinBushcraft      int
	RequiresFire      bool
	RequiresShelter   bool
	WoodKg            float64
	WeightKg          float64
	BaseHours         float64
	Portable          bool
	RequiresItems     []string
	RequiresResources []ResourceRequirement
	Effects           statDelta
}

type ResourceRequirement struct {
	ID  string
	Qty float64
}

func CraftableCatalog() []CraftableSpec {
	base := []CraftableSpec{
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

		// Cordage and trap ecosystem enablers.
		{ID: "natural_twine", Name: "Natural Twine", Category: "cordage", BiomeTags: []string{"forest", "boreal", "savanna", "jungle", "wetlands", "desert", "coast"}, Description: "Twisted bark and bast fibers for trap lines and lashings.", MinBushcraft: 0, WeightKg: 0.08, BaseHours: 0.45, Portable: true, RequiresResources: []ResourceRequirement{{ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "heavy_cordage", Name: "Heavy Cordage", Category: "cordage", BiomeTags: []string{"forest", "boreal", "savanna", "jungle", "wetlands", "coast"}, Description: "Multi-strand cord for rafts, shelter, and heavy traps.", MinBushcraft: 1, WeightKg: 0.16, BaseHours: 0.8, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "inner_bark_fiber", Qty: 1}}, Effects: statDelta{Morale: 1}},

		// Watercraft.
		{ID: "brush_raft", Name: "Brush Raft", Category: "watercraft", BiomeTags: []string{"river", "lake", "delta", "wetlands", "coast", "island", "jungle"}, Description: "Improvised raft with buoyant brush and cordage lashings.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 3.2, WeightKg: 6.4, BaseHours: 3.5, Portable: false, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 2}}, Effects: statDelta{Morale: 2}},
		{ID: "reed_coracle", Name: "Reed Coracle", Category: "watercraft", BiomeTags: []string{"delta", "wetlands", "river", "lake", "coast"}, Description: "Round reed coracle sealed with pitch for short crossings.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 2.0, WeightKg: 8.5, BaseHours: 5.2, Portable: false, RequiresItems: []string{"pitch_glue", "heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 4}}, Effects: statDelta{Morale: 3}},
		{ID: "dugout_canoe", Name: "Dugout Canoe", Category: "watercraft", BiomeTags: []string{"river", "lake", "delta", "coast", "wetlands", "forest"}, Description: "Labor-intensive dugout hull for hauling gear and setting fish lines.", MinBushcraft: 3, RequiresFire: true, RequiresShelter: true, WoodKg: 5.4, WeightKg: 14.0, BaseHours: 8.5, Portable: false, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "charcoal", Qty: 2}}, Effects: statDelta{Morale: 4}},

		// Plant and hide clothing.
		{ID: "grass_cape", Name: "Grass Cape", Category: "clothing", BiomeTags: []string{"savanna", "badlands", "wetlands", "coast", "forest"}, Description: "Woven grass cape for wind and drizzle protection.", MinBushcraft: 0, WeightKg: 0.8, BaseHours: 1.5, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "dry_grass", Qty: 2}}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "bast_sandals", Name: "Bast Sandals", Category: "clothing", BiomeTags: []string{"forest", "boreal", "savanna", "desert", "coast"}, Description: "Simple woven sandals from bast and yucca fibers.", MinBushcraft: 0, WeightKg: 0.35, BaseHours: 1.0, Portable: true, RequiresResources: []ResourceRequirement{{ID: "bast_strip", Qty: 1}, {ID: "yucca_fiber", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "woven_tunic", Name: "Woven Tunic", Category: "clothing", BiomeTags: []string{"forest", "boreal", "wetlands", "jungle", "coast"}, Description: "Layered plant-fiber tunic that improves comfort and morale.", MinBushcraft: 1, WeightKg: 1.1, BaseHours: 3.0, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "nettle_fiber", Qty: 2}}, Effects: statDelta{Energy: 1, Morale: 2}},
		{ID: "hide_moccasins", Name: "Hide Moccasins", Category: "clothing", BiomeTags: []string{"forest", "savanna", "badlands", "mountain", "coast"}, Description: "Soft hide footwear with fiber stitching.", MinBushcraft: 1, WeightKg: 0.5, BaseHours: 2.2, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "rawhide_strip", Qty: 1}}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "hide_jacket", Name: "Hide Jacket", Category: "clothing", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain", "badlands"}, Description: "Insulating hide layer for cold and wind exposure.", MinBushcraft: 2, WeightKg: 2.0, BaseHours: 4.6, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "rawhide_strip", Qty: 2}}, Effects: statDelta{Energy: 2, Morale: 1}},
	}
	return append(base, expandedCraftableCatalog()...)
}

func expandedCraftableCatalog() []CraftableSpec {
	return []CraftableSpec{
		// Tooling and fabrication.
		{ID: "bone_needle", Name: "Bone Needle", Category: "tools", BiomeTags: []string{"forest", "boreal", "savanna", "jungle", "coast"}, Description: "Fine needle for hide and plant-fiber stitching.", MinBushcraft: 1, WeightKg: 0.03, BaseHours: 0.6, Portable: true, RequiresResources: []ResourceRequirement{{ID: "shell_fragment", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "bone_awl", Name: "Bone Awl", Category: "tools", BiomeTags: []string{"forest", "boreal", "savanna", "jungle", "coast"}, Description: "Piercing awl for leather and bark work.", MinBushcraft: 1, WeightKg: 0.08, BaseHours: 0.7, Portable: true, RequiresResources: []ResourceRequirement{{ID: "shell_fragment", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "stone_adze", Name: "Stone Adze", Category: "tools", BiomeTags: []string{"forest", "mountain", "badlands", "river", "coast"}, Description: "Heavy cutting adze for shaping beams and planks.", MinBushcraft: 2, WeightKg: 1.1, BaseHours: 2.4, Portable: true, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 1}}, Effects: statDelta{Energy: 2}},
		{ID: "wood_mallet", Name: "Wood Mallet", Category: "tools", BiomeTags: []string{"forest", "mountain", "boreal", "savanna"}, Description: "Mallet used for pegs, wedges, and deadfall setups.", MinBushcraft: 1, WoodKg: 0.9, WeightKg: 0.8, BaseHours: 1.2, Portable: true, Effects: statDelta{Energy: 1}},
		{ID: "wedge_set", Name: "Wedge Set", Category: "tools", BiomeTags: []string{"forest", "mountain", "boreal", "coast", "badlands"}, Description: "Wedges for splitting wood and opening joints.", MinBushcraft: 1, WoodKg: 0.6, WeightKg: 0.4, BaseHours: 1.0, Portable: true, RequiresItems: []string{"wood_mallet"}, Effects: statDelta{Energy: 1}},

		// Fire and cooking.
		{ID: "ember_pot", Name: "Ember Pot", Category: "fire", BiomeTags: []string{"forest", "boreal", "mountain", "coast", "river"}, Description: "Keeps live embers for easier fire restart.", MinBushcraft: 2, RequiresFire: true, WeightKg: 1.2, BaseHours: 2.2, Portable: true, RequiresResources: []ResourceRequirement{{ID: "clay", Qty: 1.0}}, Effects: statDelta{Morale: 1}},
		{ID: "drying_rack", Name: "Drying Rack", Category: "food", BiomeTags: []string{"forest", "coast", "savanna", "jungle", "mountain"}, Description: "Passive drying rack for meat, fish, and herbs.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.4, WeightKg: 4.5, BaseHours: 2.8, Portable: false, RequiresItems: []string{"heavy_cordage"}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "smoking_rack", Name: "Smoking Rack", Category: "food", BiomeTags: []string{"forest", "coast", "savanna", "jungle", "wetlands"}, Description: "Controlled smoking rack for preservation.", MinBushcraft: 2, RequiresFire: true, RequiresShelter: true, WoodKg: 1.8, WeightKg: 5.0, BaseHours: 3.4, Portable: false, RequiresItems: []string{"drying_rack"}, Effects: statDelta{Morale: 2}},
		{ID: "stone_oven", Name: "Stone Oven", Category: "food", BiomeTags: []string{"mountain", "badlands", "river", "coast", "forest"}, Description: "Stone-lined oven for baking and long cooks.", MinBushcraft: 2, RequiresFire: true, RequiresShelter: true, WoodKg: 1.0, WeightKg: 14, BaseHours: 5.0, Portable: false, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 3}, {ID: "mud", Qty: 1}}, Effects: statDelta{Morale: 2}},
		{ID: "clay_oven", Name: "Clay Oven", Category: "food", BiomeTags: []string{"river", "delta", "wetlands", "swamp", "lake", "coast"}, Description: "Compact clay oven improves cook consistency.", MinBushcraft: 2, RequiresFire: true, RequiresShelter: true, WoodKg: 0.8, WeightKg: 10, BaseHours: 4.5, Portable: false, RequiresResources: []ResourceRequirement{{ID: "clay", Qty: 2.2}, {ID: "gravel", Qty: 1}}, Effects: statDelta{Morale: 2}},
		{ID: "char_cloth", Name: "Char Cloth", Category: "fire", BiomeTags: []string{"forest", "boreal", "coast", "savanna", "jungle"}, Description: "Charred tinder cloth for high-reliability fire starts.", MinBushcraft: 1, RequiresFire: true, WeightKg: 0.04, BaseHours: 0.7, Portable: true, RequiresResources: []ResourceRequirement{{ID: "flax_fiber", Qty: 1}, {ID: "charcoal", Qty: 1}}, Effects: statDelta{Morale: 1}},

		// Hunting and projectile systems.
		{ID: "hunting_blind", Name: "Hunting Blind", Category: "structures", BiomeTags: []string{"forest", "savanna", "wetlands", "badlands", "mountain"}, Description: "Concealed blind for patient hunting setups.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.3, WeightKg: 3.5, BaseHours: 2.1, Portable: false, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "camouflage_screen", Name: "Camouflage Screen", Category: "shelter_upgrade", BiomeTags: []string{"forest", "savanna", "wetlands", "jungle", "mountain"}, Description: "Camouflage panel for blinds and camps.", MinBushcraft: 1, RequiresShelter: true, WeightKg: 0.8, BaseHours: 1.3, Portable: false, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "dry_leaf_litter", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "fire_hardened_spear", Name: "Fire-Hardened Spear", Category: "hunting", BiomeTags: []string{"forest", "coast", "river", "jungle", "savanna", "badlands"}, Description: "General-purpose hunting and fishing spear.", MinBushcraft: 1, RequiresFire: true, WoodKg: 0.8, WeightKg: 1.2, BaseHours: 1.4, Portable: true, Effects: statDelta{Energy: 1}},
		{ID: "atlatl", Name: "Atlatl Thrower", Category: "hunting", BiomeTags: []string{"forest", "mountain", "savanna", "badlands", "coast"}, Description: "Spear thrower that increases launch force.", MinBushcraft: 2, WoodKg: 0.6, WeightKg: 0.7, BaseHours: 2.0, Portable: true, RequiresItems: []string{"fire_hardened_spear", "natural_twine"}, Effects: statDelta{Morale: 1}},
		{ID: "short_bow", Name: "Short Bow", Category: "hunting", BiomeTags: []string{"forest", "coast", "savanna", "badlands", "mountain"}, Description: "Compact hunting bow for mixed terrain.", MinBushcraft: 2, WoodKg: 0.8, WeightKg: 0.9, BaseHours: 3.2, Portable: true, RequiresItems: []string{"heavy_cordage"}, Effects: statDelta{Morale: 2}},
		{ID: "long_bow", Name: "Long Bow", Category: "hunting", BiomeTags: []string{"forest", "boreal", "mountain", "coast"}, Description: "Longer draw bow with better range.", MinBushcraft: 3, WoodKg: 1.1, WeightKg: 1.3, BaseHours: 4.1, Portable: true, RequiresItems: []string{"short_bow", "heavy_cordage"}, Effects: statDelta{Morale: 2}},
		{ID: "stone_arrow_bundle", Name: "Stone Arrow Bundle", Category: "hunting", BiomeTags: []string{"forest", "mountain", "badlands", "coast", "savanna"}, Description: "Bundle of stone-tipped arrows.", MinBushcraft: 2, WeightKg: 0.7, BaseHours: 1.8, Portable: true, RequiresItems: []string{"short_bow", "natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "stone_flake", Qty: 2}}, Effects: statDelta{Energy: 1}},
		{ID: "bone_arrow_bundle", Name: "Bone Arrow Bundle", Category: "hunting", BiomeTags: []string{"forest", "boreal", "wetlands", "coast", "jungle"}, Description: "Light arrows tipped with bone/shell.", MinBushcraft: 2, WeightKg: 0.6, BaseHours: 1.7, Portable: true, RequiresItems: []string{"short_bow", "natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "shell_fragment", Qty: 2}}, Effects: statDelta{Energy: 1}},

		// Trapping and fishing infrastructure.
		{ID: "spring_snare_kit", Name: "Spring Snare Kit", Category: "trapping", BiomeTags: []string{"forest", "boreal", "savanna", "mountain", "badlands"}, Description: "Prebuilt spring snare loops and toggles.", MinBushcraft: 2, WeightKg: 0.5, BaseHours: 1.6, Portable: true, RequiresItems: []string{"trap_trigger_set", "heavy_cordage"}, Effects: statDelta{Morale: 1}},
		{ID: "deadfall_kit", Name: "Deadfall Kit", Category: "trapping", BiomeTags: []string{"forest", "boreal", "mountain", "badlands", "desert"}, Description: "Prepared trigger sticks and guide pegs for deadfalls.", MinBushcraft: 2, WeightKg: 0.7, BaseHours: 1.4, Portable: true, RequiresItems: []string{"trap_trigger_set", "wedge_set"}, Effects: statDelta{Energy: 1}},
		{ID: "fish_trap_basket", Name: "Fish Trap Basket", Category: "fishing", BiomeTags: []string{"delta", "river", "lake", "swamp", "coast", "wetlands"}, Description: "Funnel fish basket for passive channel trapping.", MinBushcraft: 1, WeightKg: 1.2, BaseHours: 1.8, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 2}}, Effects: statDelta{Energy: 1}},
		{ID: "fish_weir_stakes", Name: "Fish Weir Stakes", Category: "fishing", BiomeTags: []string{"river", "delta", "wetlands", "coast"}, Description: "Stake set for temporary stream weirs.", MinBushcraft: 2, WeightKg: 2.0, BaseHours: 2.6, Portable: false, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "willow_withy", Qty: 2}}, Effects: statDelta{Energy: 1}},
		{ID: "gill_net", Name: "Gill Net", Category: "fishing", BiomeTags: []string{"coast", "delta", "lake", "river", "wetlands"}, Description: "Net for unattended fish capture in current.", MinBushcraft: 3, WeightKg: 2.2, BaseHours: 4.0, Portable: true, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "hemp_fiber", Qty: 2}}, Effects: statDelta{Morale: 1}},
		{ID: "trotline_set", Name: "Trotline Set", Category: "fishing", BiomeTags: []string{"river", "lake", "delta", "coast"}, Description: "Longline with multiple gorge hooks.", MinBushcraft: 2, WeightKg: 0.9, BaseHours: 2.0, Portable: true, RequiresItems: []string{"fish_gorge_hooks", "heavy_cordage"}, Effects: statDelta{Energy: 1}},

		// Bedding, storage, and shelter upgrades.
		{ID: "raised_sleeping_platform", Name: "Raised Sleeping Platform", Category: "shelter_upgrade", BiomeTags: []string{"forest", "wetlands", "swamp", "jungle", "coast"}, Description: "Keeps sleeping area off wet or cold ground.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 1.6, WeightKg: 5.5, BaseHours: 2.6, Portable: false, RequiresResources: []ResourceRequirement{{ID: "spruce_bough", Qty: 1}}, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "bough_mattress", Name: "Bough Mattress", Category: "shelter_upgrade", BiomeTags: []string{"forest", "boreal", "wetlands", "mountain"}, Description: "Springy bough layer that improves sleep quality.", MinBushcraft: 1, RequiresShelter: true, WeightKg: 2.0, BaseHours: 1.1, Portable: false, RequiresResources: []ResourceRequirement{{ID: "spruce_bough", Qty: 2}}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "insulated_bedding", Name: "Insulated Bedding", Category: "shelter_upgrade", BiomeTags: []string{"forest", "boreal", "tundra", "mountain", "wetlands"}, Description: "Multi-layer bedding for cold-weather rest.", MinBushcraft: 2, RequiresShelter: true, WeightKg: 1.6, BaseHours: 2.2, Portable: false, RequiresResources: []ResourceRequirement{{ID: "dry_moss", Qty: 2}, {ID: "rawhide_strip", Qty: 1}}, Effects: statDelta{Energy: 2, Morale: 1}},
		{ID: "groundsheet", Name: "Groundsheet", Category: "shelter_upgrade", BiomeTags: []string{"forest", "coast", "wetlands", "savanna", "jungle"}, Description: "Water-resistant sheet to keep bedroll dry.", MinBushcraft: 1, RequiresShelter: true, WeightKg: 1.0, BaseHours: 1.4, Portable: false, RequiresResources: []ResourceRequirement{{ID: "bark_pitch", Qty: 1}, {ID: "reed_bundle", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "storm_flap", Name: "Storm Flap", Category: "shelter_upgrade", BiomeTags: []string{"forest", "coast", "wetlands", "mountain", "tundra"}, Description: "Wind/rain flap for shelter entrances.", MinBushcraft: 1, RequiresShelter: true, WeightKg: 0.7, BaseHours: 1.2, Portable: false, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "drainage_ditch", Name: "Drainage Ditch", Category: "shelter_upgrade", BiomeTags: []string{"forest", "wetlands", "swamp", "mountain", "coast"}, Description: "Perimeter ditch to reduce pooling water.", MinBushcraft: 1, RequiresShelter: true, BaseHours: 1.0, Portable: false, RequiresItems: []string{"digging_stick"}, RequiresResources: []ResourceRequirement{{ID: "gravel", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "reflective_fire_wall", Name: "Reflective Fire Wall", Category: "shelter_upgrade", BiomeTags: []string{"forest", "mountain", "boreal", "badlands"}, Description: "Reflector wall that bounces heat into shelter.", MinBushcraft: 2, RequiresFire: true, RequiresShelter: true, WoodKg: 1.2, WeightKg: 2.5, BaseHours: 2.1, Portable: false, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 1}}, Effects: statDelta{Energy: 2}},
		{ID: "stone_hearth", Name: "Stone Hearth", Category: "shelter_upgrade", BiomeTags: []string{"forest", "mountain", "badlands", "coast", "boreal"}, Description: "Contained hearth for safer heat and cooking.", MinBushcraft: 2, RequiresFire: true, RequiresShelter: true, WeightKg: 8, BaseHours: 2.8, Portable: false, RequiresResources: []ResourceRequirement{{ID: "stone_cobble", Qty: 2}, {ID: "mud", Qty: 0.8}}, Effects: statDelta{Morale: 1}},
		{ID: "smoke_hole_baffle", Name: "Smoke-Hole Baffle", Category: "shelter_upgrade", BiomeTags: []string{"forest", "boreal", "mountain", "tundra", "wetlands"}, Description: "Improves smoke exit while reducing wind-driven rain.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 0.7, WeightKg: 1.2, BaseHours: 1.8, Portable: false, RequiresResources: []ResourceRequirement{{ID: "thatch_bundle", Qty: 1}}, Effects: statDelta{Morale: 1}},
		{ID: "storage_shelves", Name: "Storage Shelves", Category: "shelter_upgrade", BiomeTags: []string{"forest", "coast", "mountain", "jungle", "boreal"}, Description: "Raised shelves to keep stores off damp ground.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.1, WeightKg: 3.0, BaseHours: 1.9, Portable: false, Effects: statDelta{Morale: 1}},
		{ID: "elevated_food_cache", Name: "Elevated Food Cache", Category: "storage", BiomeTags: []string{"forest", "boreal", "wetlands", "tundra", "mountain"}, Description: "Suspended cache to protect food from scavengers.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 1.7, WeightKg: 4.2, BaseHours: 2.7, Portable: false, RequiresItems: []string{"heavy_cordage"}, Effects: statDelta{Morale: 1}},
		{ID: "underground_cold_pit", Name: "Underground Cold Pit", Category: "storage", BiomeTags: []string{"forest", "boreal", "mountain", "badlands", "tundra"}, Description: "Cool pit storage for preserving perishables.", MinBushcraft: 2, RequiresShelter: true, BaseHours: 2.5, Portable: false, RequiresItems: []string{"digging_stick"}, RequiresResources: []ResourceRequirement{{ID: "gravel", Qty: 1}, {ID: "clay", Qty: 0.8}}, Effects: statDelta{Energy: 1}},
		{ID: "drying_box", Name: "Drying Box", Category: "storage", BiomeTags: []string{"forest", "coast", "savanna", "jungle", "badlands"}, Description: "Ventilated drying box for jerky and herbs.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 1.2, WeightKg: 3.4, BaseHours: 2.0, Portable: false, RequiresItems: []string{"drying_rack"}, Effects: statDelta{Morale: 1}},

		// Mobility/transport.
		{ID: "hand_sled", Name: "Hand Sled", Category: "transport", BiomeTags: []string{"tundra", "arctic", "subarctic", "boreal", "mountain"}, Description: "Simple drag sled for moving heavy loads.", MinBushcraft: 2, WoodKg: 2.2, WeightKg: 7.5, BaseHours: 3.2, Portable: false, RequiresItems: []string{"wedge_set", "heavy_cordage"}, Effects: statDelta{Energy: 1}},
		{ID: "travois", Name: "Travois", Category: "transport", BiomeTags: []string{"savanna", "badlands", "forest", "mountain", "desert"}, Description: "Pole drag for hauling camp materials.", MinBushcraft: 1, WoodKg: 1.8, WeightKg: 6.2, BaseHours: 2.4, Portable: false, RequiresItems: []string{"heavy_cordage"}, Effects: statDelta{Energy: 1}},
		{ID: "snowshoes", Name: "Snowshoes", Category: "transport", BiomeTags: []string{"tundra", "arctic", "subarctic", "boreal", "mountain"}, Description: "Flotation footwear for deep snow travel.", MinBushcraft: 2, WeightKg: 2.1, BaseHours: 3.3, Portable: true, RequiresItems: []string{"heavy_cordage"}, RequiresResources: []ResourceRequirement{{ID: "willow_withy", Qty: 1}}, Effects: statDelta{Energy: 2}},
		{ID: "skis", Name: "Improvised Skis", Category: "transport", BiomeTags: []string{"tundra", "arctic", "subarctic", "boreal", "mountain"}, Description: "Skis for efficient winter movement.", MinBushcraft: 3, WeightKg: 2.8, BaseHours: 4.2, Portable: true, RequiresItems: []string{"wood_mallet", "wedge_set", "heavy_cordage"}, Effects: statDelta{Energy: 2, Morale: 1}},

		// Structures and defenses.
		{ID: "lookout_platform", Name: "Lookout Platform", Category: "structures", BiomeTags: []string{"forest", "savanna", "wetlands", "badlands", "coast"}, Description: "Raised observation platform to scout terrain.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 2.4, WeightKg: 9, BaseHours: 3.8, Portable: false, RequiresItems: []string{"wedge_set"}, Effects: statDelta{Morale: 2}},
		{ID: "perimeter_deadfall_deterrent", Name: "Perimeter Deadfall Deterrent", Category: "structures", BiomeTags: []string{"forest", "boreal", "mountain", "badlands", "savanna"}, Description: "Non-lethal perimeter obstacle to discourage intrusions.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 1.6, WeightKg: 4.8, BaseHours: 2.2, Portable: false, RequiresItems: []string{"deadfall_kit"}, Effects: statDelta{Morale: 1}},
		{ID: "raised_storage_platform", Name: "Raised Storage Platform", Category: "structures", BiomeTags: []string{"forest", "wetlands", "swamp", "delta", "jungle"}, Description: "Platform for dry gear storage above flood line.", MinBushcraft: 2, RequiresShelter: true, WoodKg: 2.0, WeightKg: 7, BaseHours: 2.9, Portable: false, RequiresItems: []string{"heavy_cordage"}, Effects: statDelta{Morale: 1}},
		{ID: "door_latch", Name: "Door Latch", Category: "shelter_upgrade", BiomeTags: []string{"forest", "boreal", "mountain", "coast", "badlands"}, Description: "Simple latch to secure shelter entry.", MinBushcraft: 1, RequiresShelter: true, WoodKg: 0.4, WeightKg: 0.3, BaseHours: 0.8, Portable: false, RequiresItems: []string{"bone_awl"}, Effects: statDelta{Morale: 1}},
		{ID: "cold_air_trench", Name: "Cold-Air Trench", Category: "shelter_upgrade", BiomeTags: []string{"arctic", "subarctic", "tundra", "boreal", "mountain"}, Description: "Entrance trench that traps heavy cold air.", MinBushcraft: 1, RequiresShelter: true, BaseHours: 1.0, Portable: false, RequiresItems: []string{"digging_stick"}, Effects: statDelta{Energy: 1}},
		{ID: "insulated_wall_lining", Name: "Insulated Wall Lining", Category: "shelter_upgrade", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain", "wetlands"}, Description: "Additional inner wall insulation layer.", MinBushcraft: 2, RequiresShelter: true, WeightKg: 1.7, BaseHours: 2.1, Portable: false, RequiresResources: []ResourceRequirement{{ID: "dry_moss", Qty: 1}, {ID: "thatch_bundle", Qty: 1}}, Effects: statDelta{Energy: 2}},

		// Additional clothing progression.
		{ID: "hide_trousers", Name: "Hide Trousers", Category: "clothing", BiomeTags: []string{"forest", "boreal", "subarctic", "mountain", "badlands"}, Description: "Insulated hide legwear.", MinBushcraft: 2, WeightKg: 1.5, BaseHours: 3.8, Portable: true, RequiresItems: []string{"bone_needle", "natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "rawhide_strip", Qty: 2}}, Effects: statDelta{Energy: 2}},
		{ID: "bark_cloak", Name: "Bark Cloak", Category: "clothing", BiomeTags: []string{"forest", "boreal", "coast", "wetlands", "jungle"}, Description: "Layered bark/fiber cloak for rain and wind.", MinBushcraft: 1, WeightKg: 1.2, BaseHours: 2.6, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "cedar_bark", Qty: 1}, {ID: "bast_strip", Qty: 1}}, Effects: statDelta{Energy: 1, Morale: 1}},
		{ID: "reed_sandals", Name: "Reed Sandals", Category: "clothing", BiomeTags: []string{"wetlands", "swamp", "delta", "coast", "jungle"}, Description: "Woven reed sandals for wet terrain.", MinBushcraft: 0, WeightKg: 0.3, BaseHours: 0.9, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "reed_bundle", Qty: 1}}, Effects: statDelta{Energy: 1}},
		{ID: "fiber_poncho", Name: "Fiber Poncho", Category: "clothing", BiomeTags: []string{"forest", "coast", "wetlands", "jungle", "savanna"}, Description: "Quick rain poncho from woven fibers.", MinBushcraft: 1, WeightKg: 0.9, BaseHours: 1.8, Portable: true, RequiresItems: []string{"natural_twine"}, RequiresResources: []ResourceRequirement{{ID: "hemp_fiber", Qty: 1}}, Effects: statDelta{Morale: 1}},
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

func (s *RunState) CraftItem(playerID int, craftID string) (CraftOutcome, error) {
	if s == nil {
		return CraftOutcome{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return CraftOutcome{}, fmt.Errorf("player %d not found", playerID)
	}

	craftID = strings.ToLower(strings.TrimSpace(craftID))
	if craftID == "" {
		return CraftOutcome{}, fmt.Errorf("craft item id required")
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
		return CraftOutcome{}, fmt.Errorf("craft item not available in biome: %s", craftID)
	}

	effectiveCraft := player.Bushcraft + player.Crafting/25 + player.Agility + positiveTraitModifier(player.Traits)/2 + negativeTraitModifier(player.Traits)/2
	if effectiveCraft < chosen.MinBushcraft {
		return CraftOutcome{}, fmt.Errorf("requires bushcraft %+d", chosen.MinBushcraft)
	}
	if chosen.RequiresFire && !s.Fire.Lit {
		return CraftOutcome{}, fmt.Errorf("requires active fire")
	}
	if chosen.RequiresShelter && (s.Shelter.Type == "" || s.Shelter.Durability <= 0) {
		return CraftOutcome{}, fmt.Errorf("requires active shelter")
	}
	for _, needed := range chosen.RequiresItems {
		if !slices.Contains(s.CraftedItems, needed) {
			return CraftOutcome{}, fmt.Errorf("requires crafted item: %s", needed)
		}
	}
	for _, needed := range chosen.RequiresResources {
		if s.resourceQty(needed.ID) < needed.Qty {
			return CraftOutcome{}, fmt.Errorf("requires resource: %s %.1f", needed.ID, needed.Qty)
		}
	}

	itemWeightKg := chosen.WeightKg
	if itemWeightKg <= 0 {
		itemWeightKg = clampFloat(chosen.WoodKg*0.62+float64(len(chosen.RequiresResources))*0.08+0.16, 0.12, 20)
	}
	baseHours := chosen.BaseHours
	if baseHours <= 0 {
		baseHours = clampFloat(0.35+(chosen.WoodKg*0.72)+float64(len(chosen.RequiresResources))*0.25+float64(len(chosen.RequiresItems))*0.2, 0.25, 8.5)
	}
	portability := chosen.Portable
	if !portability && chosen.Category == "" {
		portability = true
	}
	storeAt := "camp"
	if portability {
		if inventoryWeightKg(player.PersonalItems)+itemWeightKg <= s.playerCarryLimitKg(player)+1e-9 {
			storeAt = "personal"
		} else if !s.canStoreAtCamp(itemWeightKg) {
			return CraftOutcome{}, fmt.Errorf("no storage space (personal carry and camp inventory full)")
		}
	} else if !s.canStoreAtCamp(itemWeightKg) {
		return CraftOutcome{}, fmt.Errorf("camp inventory full (%.1f/%.1fkg)", s.campUsedKg(), s.campCapacityKg())
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
			return CraftOutcome{}, fmt.Errorf("needs %.1fkg wood", chosen.WoodKg)
		}
	}
	for _, needed := range chosen.RequiresResources {
		_ = s.consumeResourceStock(needed.ID, needed.Qty)
	}

	if !slices.Contains(s.CraftedItems, chosen.ID) {
		s.CraftedItems = append(s.CraftedItems, chosen.ID)
	}
	if s.Shelter.Type != "" && s.Shelter.Durability > 0 && (strings.EqualFold(chosen.Category, "shelter_upgrade") || isShelterUpgradeID(chosen.ID)) {
		if !slices.Contains(s.Shelter.Upgrades, chosen.ID) {
			s.Shelter.Upgrades = append(s.Shelter.Upgrades, chosen.ID)
		}
	}
	rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("craft:%s:%d:%d", chosen.ID, s.Day, playerID)))
	qualityScore := float64(effectiveCraft) + float64(player.MentalStrength)/2 + rng.Float64()*2.4 - 1.2
	quality := qualityFromScore(qualityScore)
	hours := clampFloat(baseHours-qualityTimeReduction(quality), 0.2, 14)
	_ = s.AdvanceActionClock(hours)

	item := InventoryItem{
		ID:       chosen.ID,
		Name:     chosen.Name,
		Unit:     "set",
		Qty:      1,
		WeightKg: itemWeightKg,
		Category: chosen.Category,
		Quality:  string(quality),
	}
	if storeAt == "personal" {
		if err := s.AddPersonalInventoryItem(playerID, item); err != nil {
			if err := s.addCampInventoryItem(item); err != nil {
				return CraftOutcome{}, err
			}
			storeAt = "camp"
		}
	} else {
		if err := s.addCampInventoryItem(item); err != nil {
			return CraftOutcome{}, err
		}
	}

	applySkillEffort(&player.Crafting, int(math.Round(hours*18)), true)
	switch strings.ToLower(strings.TrimSpace(chosen.Category)) {
	case "shelter_upgrade", "structures":
		applySkillEffort(&player.Sheltercraft, int(math.Round(hours*14)), true)
	case "fire":
		applySkillEffort(&player.Firecraft, int(math.Round(hours*14)), true)
	case "fishing", "trapping":
		applySkillEffort(&player.Trapping, int(math.Round(hours*12)), true)
	}
	player.Energy = clamp(player.Energy+chosen.Effects.Energy-int(math.Ceil(hours*2))+qualityCraftEffectBonus(quality), 0, 100)
	player.Hydration = clamp(player.Hydration+chosen.Effects.Hydration-int(math.Ceil(hours*1.4)), 0, 100)
	player.Morale = clamp(player.Morale+chosen.Effects.Morale+1+qualityCraftEffectBonus(quality), 0, 100)
	refreshEffectBars(player)
	return CraftOutcome{
		Spec:       chosen,
		Quality:    quality,
		HoursSpent: hours,
		StoredAt:   storeAt,
	}, nil
}

func isShelterUpgradeID(id string) bool {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "raised_sleeping_platform",
		"bough_mattress",
		"insulated_bedding",
		"groundsheet",
		"storm_flap",
		"drainage_ditch",
		"reflective_fire_wall",
		"stone_hearth",
		"smoke_hole_baffle",
		"storage_shelves",
		"elevated_food_cache",
		"door_latch",
		"camouflage_screen",
		"lookout_platform",
		"cold_air_trench",
		"insulated_wall_lining":
		return true
	default:
		return false
	}
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
