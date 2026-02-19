package game

import (
	"fmt"
	"math"
	"strings"
)

type foodItemSpec struct {
	ID              string
	Name            string
	Category        string
	Cooked          bool
	Preserved       bool
	Perishable      bool
	ShelfLifeDays   int
	DecayPerDay     float64
	NutritionPer100 NutritionPer100g
	IllnessRisk     float64
}

var foodItemCatalog = map[string]foodItemSpec{
	"raw_small_game_meat":    {ID: "raw_small_game_meat", Name: "Raw Small Game Meat", Category: "meat", Cooked: false, Perishable: true, ShelfLifeDays: 1, DecayPerDay: 0.58, NutritionPer100: NutritionPer100g{CaloriesKcal: 150, ProteinG: 22, FatG: 6, SugarG: 0}, IllnessRisk: 0.16},
	"raw_bird_meat":          {ID: "raw_bird_meat", Name: "Raw Bird Meat", Category: "meat", Cooked: false, Perishable: true, ShelfLifeDays: 1, DecayPerDay: 0.6, NutritionPer100: NutritionPer100g{CaloriesKcal: 145, ProteinG: 21, FatG: 5, SugarG: 0}, IllnessRisk: 0.18},
	"raw_fish_meat":          {ID: "raw_fish_meat", Name: "Raw Fish Meat", Category: "fish", Cooked: false, Perishable: true, ShelfLifeDays: 1, DecayPerDay: 0.64, NutritionPer100: NutritionPer100g{CaloriesKcal: 120, ProteinG: 20, FatG: 4, SugarG: 0}, IllnessRisk: 0.12},
	"spoiled_meat":           {ID: "spoiled_meat", Name: "Spoiled Meat", Category: "waste", Cooked: false, Perishable: true, ShelfLifeDays: 2, DecayPerDay: 0.36, NutritionPer100: NutritionPer100g{CaloriesKcal: 60, ProteinG: 5, FatG: 2, SugarG: 0}, IllnessRisk: 0.45},
	"cooked_small_game_meat": {ID: "cooked_small_game_meat", Name: "Cooked Small Game Meat", Category: "meat", Cooked: true, Perishable: true, ShelfLifeDays: 2, DecayPerDay: 0.34, NutritionPer100: NutritionPer100g{CaloriesKcal: 205, ProteinG: 28, FatG: 8, SugarG: 0}, IllnessRisk: 0.02},
	"cooked_bird_meat":       {ID: "cooked_bird_meat", Name: "Cooked Bird Meat", Category: "meat", Cooked: true, Perishable: true, ShelfLifeDays: 2, DecayPerDay: 0.36, NutritionPer100: NutritionPer100g{CaloriesKcal: 190, ProteinG: 26, FatG: 7, SugarG: 0}, IllnessRisk: 0.03},
	"cooked_fish_meat":       {ID: "cooked_fish_meat", Name: "Cooked Fish Meat", Category: "fish", Cooked: true, Perishable: true, ShelfLifeDays: 2, DecayPerDay: 0.4, NutritionPer100: NutritionPer100g{CaloriesKcal: 160, ProteinG: 24, FatG: 6, SugarG: 0}, IllnessRisk: 0.01},
	"smoked_small_game_meat": {ID: "smoked_small_game_meat", Name: "Smoked Small Game Meat", Category: "preserved_meat", Cooked: true, Preserved: true, Perishable: true, ShelfLifeDays: 10, DecayPerDay: 0.12, NutritionPer100: NutritionPer100g{CaloriesKcal: 230, ProteinG: 31, FatG: 9, SugarG: 0}, IllnessRisk: 0.015},
	"smoked_bird_meat":       {ID: "smoked_bird_meat", Name: "Smoked Bird Meat", Category: "preserved_meat", Cooked: true, Preserved: true, Perishable: true, ShelfLifeDays: 9, DecayPerDay: 0.13, NutritionPer100: NutritionPer100g{CaloriesKcal: 215, ProteinG: 29, FatG: 8, SugarG: 0}, IllnessRisk: 0.02},
	"smoked_fish_meat":       {ID: "smoked_fish_meat", Name: "Smoked Fish Meat", Category: "preserved_fish", Cooked: true, Preserved: true, Perishable: true, ShelfLifeDays: 8, DecayPerDay: 0.14, NutritionPer100: NutritionPer100g{CaloriesKcal: 195, ProteinG: 28, FatG: 7, SugarG: 0}, IllnessRisk: 0.015},
	"dried_small_game_meat":  {ID: "dried_small_game_meat", Name: "Dried Small Game Meat", Category: "preserved_meat", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 18, DecayPerDay: 0.08, NutritionPer100: NutritionPer100g{CaloriesKcal: 255, ProteinG: 35, FatG: 10, SugarG: 0}, IllnessRisk: 0.03},
	"dried_bird_meat":        {ID: "dried_bird_meat", Name: "Dried Bird Meat", Category: "preserved_meat", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 16, DecayPerDay: 0.09, NutritionPer100: NutritionPer100g{CaloriesKcal: 240, ProteinG: 33, FatG: 9, SugarG: 0}, IllnessRisk: 0.035},
	"dried_fish_meat":        {ID: "dried_fish_meat", Name: "Dried Fish Meat", Category: "preserved_fish", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 14, DecayPerDay: 0.1, NutritionPer100: NutritionPer100g{CaloriesKcal: 220, ProteinG: 34, FatG: 8, SugarG: 0}, IllnessRisk: 0.03},
	"salted_small_game_meat": {ID: "salted_small_game_meat", Name: "Salted Small Game Meat", Category: "preserved_meat", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 24, DecayPerDay: 0.06, NutritionPer100: NutritionPer100g{CaloriesKcal: 210, ProteinG: 30, FatG: 8, SugarG: 0}, IllnessRisk: 0.025},
	"salted_bird_meat":       {ID: "salted_bird_meat", Name: "Salted Bird Meat", Category: "preserved_meat", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 22, DecayPerDay: 0.07, NutritionPer100: NutritionPer100g{CaloriesKcal: 200, ProteinG: 28, FatG: 7, SugarG: 0}, IllnessRisk: 0.03},
	"salted_fish_meat":       {ID: "salted_fish_meat", Name: "Salted Fish Meat", Category: "preserved_fish", Cooked: false, Preserved: true, Perishable: true, ShelfLifeDays: 20, DecayPerDay: 0.08, NutritionPer100: NutritionPer100g{CaloriesKcal: 185, ProteinG: 27, FatG: 6, SugarG: 0}, IllnessRisk: 0.028},
}

type carcassSpec struct {
	ID         string
	Name       string
	MeatID     string
	EdibleBase float64
}

var carcassCatalog = map[string]carcassSpec{
	"small_game_carcass": {ID: "small_game_carcass", Name: "Small Game Carcass", MeatID: "raw_small_game_meat", EdibleBase: 0.6},
	"bird_carcass":       {ID: "bird_carcass", Name: "Bird Carcass", MeatID: "raw_bird_meat", EdibleBase: 0.56},
	"fish_carcass":       {ID: "fish_carcass", Name: "Fish Carcass", MeatID: "raw_fish_meat", EdibleBase: 0.64},
}

type GutResult struct {
	PlayerID    int
	CarcassID   string
	ProcessedKg float64
	MeatID      string
	MeatKg      float64
	SpoiledKg   float64
	InedibleKg  float64
	PiercedGut  bool
	HoursSpent  float64
}

type CookResult struct {
	PlayerID   int
	RawID      string
	CookedID   string
	InputKg    float64
	OutputKg   float64
	HoursSpent float64
}

type PreserveResult struct {
	PlayerID      int
	Method        string
	SourceID      string
	PreservedID   string
	InputKg       float64
	OutputKg      float64
	HoursSpent    float64
	ShelfLifeDays int
}

type EatResult struct {
	PlayerID       int
	ItemID         string
	ConsumedGrams  int
	Nutrition      NutritionTotals
	EnergyDelta    int
	HydrationDelta int
	MoraleDelta    int
	GotIll         bool
}

func (s *RunState) getInventoryQty(playerID int, itemID string) float64 {
	itemID = strings.ToLower(strings.TrimSpace(itemID))
	if itemID == "" {
		return 0
	}
	total := 0.0
	if player, ok := s.playerByID(playerID); ok {
		total += inventoryTotalQtyByID(player.PersonalItems, itemID)
	}
	total += inventoryTotalQtyByID(s.CampInventory, itemID)
	return total
}

func (s *RunState) consumeItemForPlayer(playerID int, itemID string, qty float64, preferPersonal bool) (InventoryItem, string, error) {
	itemID = strings.ToLower(strings.TrimSpace(itemID))
	if qty <= 0 {
		return InventoryItem{}, "", fmt.Errorf("quantity must be positive")
	}
	if preferPersonal {
		if got, err := s.removePersonalInventoryItem(playerID, itemID, qty); err == nil {
			return got, "personal", nil
		}
	}
	if got, err := s.removeCampInventoryItem(itemID, qty); err == nil {
		return got, "camp", nil
	}
	if !preferPersonal {
		if got, err := s.removePersonalInventoryItem(playerID, itemID, qty); err == nil {
			return got, "personal", nil
		}
	}
	return InventoryItem{}, "", fmt.Errorf("item not available: %s", itemID)
}

func (s *RunState) addItemForPlayer(playerID int, source string, item InventoryItem) error {
	if source == "personal" {
		if err := s.AddPersonalInventoryItem(playerID, item); err == nil {
			return nil
		}
	}
	return s.addCampInventoryItem(item)
}

func (s *RunState) GutCarcass(playerID int, carcassID string, kg float64) (GutResult, error) {
	if s == nil {
		return GutResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return GutResult{}, fmt.Errorf("player %d not found", playerID)
	}
	carcassID = strings.ToLower(strings.TrimSpace(carcassID))
	carcass, ok := carcassCatalog[carcassID]
	if !ok {
		return GutResult{}, fmt.Errorf("unknown carcass: %s", carcassID)
	}
	total := s.getInventoryQty(playerID, carcassID)
	if total <= 0 {
		return GutResult{}, fmt.Errorf("no %s available", carcassID)
	}
	if kg <= 0 || kg > total {
		kg = total
	}
	consumed, source, err := s.consumeItemForPlayer(playerID, carcassID, kg, true)
	if err != nil {
		return GutResult{}, err
	}
	_ = consumed

	s.ProcessAttemptCount++
	skill := float64(player.Bushcraft+player.Agility) + float64(player.Crafting)/20.0
	if carcassID == "fish_carcass" {
		skill += float64(player.Fishing) / 30.0
	} else {
		skill += float64(player.Hunting) / 30.0
	}
	toolBonus := 0.0
	if hasAnyKitItem(*player, s.Config.IssuedKit, KitSixInchKnife) || hasAnyKitItem(*player, s.Config.IssuedKit, KitMachete) || hasAnyKitItem(*player, s.Config.IssuedKit, KitMultiTool) {
		toolBonus = 0.08
	}
	pierceChance := 0.23 - (skill * 0.014) - toolBonus
	if pierceChance < 0.02 {
		pierceChance = 0.02
	}
	if pierceChance > 0.48 {
		pierceChance = 0.48
	}
	rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("gut:%s:%d:%d:%d", carcassID, s.Day, playerID, s.ProcessAttemptCount)))
	pierced := rng.Float64() <= pierceChance

	edibleRatio := carcass.EdibleBase + (skill * 0.01)
	if pierced {
		edibleRatio -= 0.35
	}
	edibleRatio = clampFloat(edibleRatio, 0.15, 0.86)
	meatKg := math.Round(kg*edibleRatio*100) / 100
	spoiledKg := 0.0
	if pierced {
		spoiledKg = math.Round(kg*0.22*100) / 100
	} else {
		spoiledKg = math.Round(kg*0.04*100) / 100
	}
	if meatKg+spoiledKg > kg {
		spoiledKg = max(0, kg-meatKg)
	}
	inedibleKg := math.Round((kg-meatKg-spoiledKg)*100) / 100
	if inedibleKg < 0 {
		inedibleKg = 0
	}

	meatItem := InventoryItem{ID: carcass.MeatID, Name: strings.ReplaceAll(carcass.MeatID, "_", " "), Unit: "kg", Qty: meatKg, WeightKg: 1, Category: "food", AgeDays: 0}
	if meatKg > 0 {
		if err := s.addItemForPlayer(playerID, source, meatItem); err != nil {
			// Rollback carcass if no space for output.
			_ = s.addItemForPlayer(playerID, source, InventoryItem{ID: carcassID, Name: carcass.Name, Unit: "kg", Qty: kg, WeightKg: 1.2, Category: "carcass", AgeDays: consumed.AgeDays})
			return GutResult{}, err
		}
	}
	if spoiledKg > 0 {
		_ = s.addItemForPlayer(playerID, source, InventoryItem{ID: "spoiled_meat", Name: "Spoiled Meat", Unit: "kg", Qty: spoiledKg, WeightKg: 1, Category: "food", AgeDays: 0})
	}

	hours := clampFloat(0.3+(kg*0.7)-(skill*0.03), 0.2, 8)
	_ = s.AdvanceActionClock(hours)
	applySkillEffort(&player.Crafting, int(math.Round(hours*18)), true)
	if carcassID == "fish_carcass" {
		applySkillEffort(&player.Fishing, int(math.Round(hours*16)), !pierced)
	} else {
		applySkillEffort(&player.Hunting, int(math.Round(hours*16)), !pierced)
	}
	player.Energy = clamp(player.Energy-int(math.Ceil(hours*3.2)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*2.2)), 0, 100)
	if pierced {
		player.Morale = clamp(player.Morale-2, 0, 100)
	}
	refreshEffectBars(player)

	return GutResult{
		PlayerID:    playerID,
		CarcassID:   carcassID,
		ProcessedKg: kg,
		MeatID:      carcass.MeatID,
		MeatKg:      meatKg,
		SpoiledKg:   spoiledKg,
		InedibleKg:  inedibleKg,
		PiercedGut:  pierced,
		HoursSpent:  hours,
	}, nil
}

func cookedItemID(rawID string) string {
	switch rawID {
	case "raw_small_game_meat":
		return "cooked_small_game_meat"
	case "raw_bird_meat":
		return "cooked_bird_meat"
	case "raw_fish_meat":
		return "cooked_fish_meat"
	default:
		return ""
	}
}

func preserveMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "smoke", "smoked":
		return "smoke"
	case "dry", "dried", "jerky":
		return "dry"
	case "salt", "salted", "cure", "cured":
		return "salt"
	default:
		return ""
	}
}

func preserveItemID(sourceID, method string) string {
	method = preserveMethod(method)
	if method == "" {
		return ""
	}
	meatType := ""
	switch {
	case strings.Contains(sourceID, "small_game"):
		meatType = "small_game"
	case strings.Contains(sourceID, "bird"):
		meatType = "bird"
	case strings.Contains(sourceID, "fish"):
		meatType = "fish"
	}
	if meatType == "" {
		return ""
	}
	switch method {
	case "smoke":
		return "smoked_" + meatType + "_meat"
	case "dry":
		return "dried_" + meatType + "_meat"
	case "salt":
		return "salted_" + meatType + "_meat"
	default:
		return ""
	}
}

func (s *RunState) PreserveFood(playerID int, method string, itemID string, kg float64) (PreserveResult, error) {
	if s == nil {
		return PreserveResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return PreserveResult{}, fmt.Errorf("player %d not found", playerID)
	}
	method = preserveMethod(method)
	if method == "" {
		return PreserveResult{}, fmt.Errorf("unknown preserve method")
	}
	itemID = strings.ToLower(strings.TrimSpace(itemID))
	spec, ok := foodItemCatalog[itemID]
	if !ok || itemID == "spoiled_meat" {
		return PreserveResult{}, fmt.Errorf("item cannot be preserved: %s", itemID)
	}
	if spec.Preserved {
		return PreserveResult{}, fmt.Errorf("item already preserved: %s", itemID)
	}
	preservedID := preserveItemID(itemID, method)
	if preservedID == "" {
		return PreserveResult{}, fmt.Errorf("item cannot be preserved: %s", itemID)
	}
	preservedSpec, ok := foodItemCatalog[preservedID]
	if !ok {
		return PreserveResult{}, fmt.Errorf("no preserved profile for %s", preservedID)
	}
	if method == "smoke" && !s.Fire.Lit {
		return PreserveResult{}, fmt.Errorf("smoking requires active fire")
	}
	if method == "salt" && !hasAnyKitItem(*player, s.Config.IssuedKit, KitSalt) && s.getInventoryQty(playerID, "salt") <= 0 {
		return PreserveResult{}, fmt.Errorf("salting requires salt kit or salt stock")
	}

	total := s.getInventoryQty(playerID, itemID)
	if total <= 0 {
		return PreserveResult{}, fmt.Errorf("no %s available", itemID)
	}
	if kg <= 0 || kg > total {
		kg = min(total, 0.8)
	}
	consumed, source, err := s.consumeItemForPlayer(playerID, itemID, kg, true)
	if err != nil {
		return PreserveResult{}, err
	}

	yield := 0.0
	hours := 0.0
	switch method {
	case "smoke":
		yield = kg * 0.82
		hours = 2.8 + (kg * 2.4)
		if strings.Contains(strings.ToLower(strings.TrimSpace(string(s.Shelter.Type))), "smoke") || containsStringFold(s.CraftedItems, "smoke_rack") {
			yield += kg * 0.04
			hours -= 0.6
		}
		s.Fire.FuelKg = max(0, s.Fire.FuelKg-(kg*0.18))
	case "dry":
		yield = kg * 0.62
		hours = 6.4 + (kg * 7.2)
		if s.Weather.Type == WeatherRain || s.Weather.Type == WeatherStorm {
			hours += 2.2
			yield -= kg * 0.04
		}
	case "salt":
		yield = kg * 0.86
		hours = 1.2 + (kg * 1.5)
	}
	yield = math.Round(clampFloat(yield, 0.1, kg)*100) / 100
	hours = clampFloat(hours-float64(player.Crafting)/180.0, 0.6, 32)

	out := InventoryItem{
		ID:       preservedID,
		Name:     strings.ReplaceAll(preservedID, "_", " "),
		Unit:     "kg",
		Qty:      yield,
		WeightKg: 1,
		Category: "food",
		AgeDays:  0,
	}
	if err := s.addItemForPlayer(playerID, source, out); err != nil {
		_ = s.addItemForPlayer(playerID, source, InventoryItem{
			ID:       itemID,
			Name:     spec.Name,
			Unit:     "kg",
			Qty:      kg,
			WeightKg: 1,
			Category: "food",
			AgeDays:  consumed.AgeDays,
		})
		return PreserveResult{}, err
	}

	_ = s.AdvanceActionClock(hours)
	applySkillEffort(&player.Crafting, int(math.Round(hours*18)), true)
	if method == "smoke" || method == "dry" {
		applySkillEffort(&player.Gathering, int(math.Round(hours*8)), true)
	}
	player.Energy = clamp(player.Energy-int(math.Ceil(hours*1.8)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*1.1)), 0, 100)
	player.Morale = clamp(player.Morale+1, 0, 100)
	refreshEffectBars(player)

	return PreserveResult{
		PlayerID:      playerID,
		Method:        method,
		SourceID:      itemID,
		PreservedID:   preservedID,
		InputKg:       kg,
		OutputKg:      yield,
		HoursSpent:    hours,
		ShelfLifeDays: preservedSpec.ShelfLifeDays,
	}, nil
}

func (s *RunState) CookFood(playerID int, rawID string, kg float64) (CookResult, error) {
	if s == nil {
		return CookResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return CookResult{}, fmt.Errorf("player %d not found", playerID)
	}
	if !s.Fire.Lit {
		return CookResult{}, fmt.Errorf("requires active fire")
	}
	rawID = strings.ToLower(strings.TrimSpace(rawID))
	spec, ok := foodItemCatalog[rawID]
	if !ok || spec.Cooked {
		return CookResult{}, fmt.Errorf("item cannot be cooked: %s", rawID)
	}
	cookedID := cookedItemID(rawID)
	if cookedID == "" {
		return CookResult{}, fmt.Errorf("no cooked variant for %s", rawID)
	}

	total := s.getInventoryQty(playerID, rawID)
	if total <= 0 {
		return CookResult{}, fmt.Errorf("no %s available", rawID)
	}
	if kg <= 0 || kg > total {
		kg = min(total, 0.6)
	}
	consumed, source, err := s.consumeItemForPlayer(playerID, rawID, kg, true)
	if err != nil {
		return CookResult{}, err
	}
	yield := math.Round(kg*0.9*100) / 100
	if hasAnyKitItem(*player, s.Config.IssuedKit, KitCookingPot) || hasAnyKitItem(*player, s.Config.IssuedKit, KitMetalCup) {
		yield = math.Round(kg*0.93*100) / 100
	}
	if yield <= 0 {
		yield = 0.1
	}

	item := InventoryItem{
		ID:       cookedID,
		Name:     strings.ReplaceAll(cookedID, "_", " "),
		Unit:     "kg",
		Qty:      yield,
		WeightKg: 1,
		Category: "food",
		AgeDays:  0,
	}
	if err := s.addItemForPlayer(playerID, source, item); err != nil {
		_ = s.addItemForPlayer(playerID, source, InventoryItem{ID: rawID, Name: spec.Name, Unit: "kg", Qty: kg, WeightKg: 1, Category: "food", AgeDays: consumed.AgeDays})
		return CookResult{}, err
	}

	hours := clampFloat(0.2+(kg*0.5)-float64(player.Crafting)/250.0, 0.12, 4)
	_ = s.AdvanceActionClock(hours)
	applySkillEffort(&player.Crafting, int(math.Round(hours*16)), true)
	player.Energy = clamp(player.Energy-int(math.Ceil(hours*1.2)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*0.8)), 0, 100)
	player.Morale = clamp(player.Morale+1, 0, 100)
	s.Fire.FuelKg = max(0, s.Fire.FuelKg-(kg*0.12))
	if s.Fire.FuelKg <= 0.05 {
		s.ExtinguishFire()
	}
	refreshEffectBars(player)

	return CookResult{
		PlayerID:   playerID,
		RawID:      rawID,
		CookedID:   cookedID,
		InputKg:    kg,
		OutputKg:   yield,
		HoursSpent: hours,
	}, nil
}

func (s *RunState) EatFood(playerID int, itemID string, amount float64) (EatResult, error) {
	if s == nil {
		return EatResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return EatResult{}, fmt.Errorf("player %d not found", playerID)
	}
	itemID = strings.ToLower(strings.TrimSpace(itemID))
	spec, ok := foodItemCatalog[itemID]
	if !ok {
		return EatResult{}, fmt.Errorf("item is not edible profile: %s", itemID)
	}

	totalKg := s.getInventoryQty(playerID, itemID)
	if totalKg <= 0 {
		return EatResult{}, fmt.Errorf("no %s available", itemID)
	}
	consumeKg := 0.25
	if amount > 0 {
		if amount <= 5 {
			consumeKg = amount
		} else {
			consumeKg = amount / 1000.0
		}
	}
	if consumeKg > totalKg {
		consumeKg = totalKg
	}
	if consumeKg <= 0 {
		consumeKg = min(totalKg, 0.2)
	}
	consumed, _, err := s.consumeItemForPlayer(playerID, itemID, consumeKg, true)
	if err != nil {
		return EatResult{}, err
	}

	grams := int(math.Round(consumed.Qty * 1000))
	if grams < 1 {
		grams = 1
	}
	nutrition := nutritionFromPer100g(spec.NutritionPer100, grams)
	applyMealNutritionReserves(player, nutrition)
	energyGain, hydrationGain, moraleGain := nutritionToPlayerEffects(nutrition)
	player.Energy = clamp(player.Energy+energyGain, 0, 100)
	player.Hydration = clamp(player.Hydration+hydrationGain, 0, 100)
	player.Morale = clamp(player.Morale+moraleGain, 0, 100)
	player.Nutrition = player.Nutrition.add(nutrition)

	gotIll := false
	illnessChance := spec.IllnessRisk
	if spec.Perishable && consumed.AgeDays > 0 {
		ageRisk := 0.018 * float64(consumed.AgeDays)
		if spec.Preserved {
			ageRisk = 0.006 * float64(consumed.AgeDays)
		}
		if spec.ShelfLifeDays > 0 && consumed.AgeDays > spec.ShelfLifeDays {
			ageRisk += 0.08
		}
		illnessChance += ageRisk
	}
	if hasAnyKitItem(*player, s.Config.IssuedKit, KitCookingPot) && spec.Cooked {
		illnessChance *= 0.7
	}
	illnessChance = clampFloat(illnessChance, 0, 0.95)
	if illnessChance > 0 {
		s.ProcessAttemptCount++
		rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("eat:%s:%d:%d:%d", itemID, s.Day, playerID, s.ProcessAttemptCount)))
		if rng.Float64() <= illnessChance {
			ailment := Ailment{
				Type:             AilmentFoodPoison,
				Name:             "Food Poisoning",
				DaysRemaining:    2,
				EnergyPenalty:    2,
				HydrationPenalty: 4,
				MoralePenalty:    3,
			}
			if itemID == "spoiled_meat" {
				ailment.DaysRemaining = 3
				ailment.EnergyPenalty = 3
				ailment.HydrationPenalty = 5
				ailment.MoralePenalty = 4
			}
			player.applyAilment(ailment)
			player.Energy = clamp(player.Energy-ailment.EnergyPenalty, 0, 100)
			player.Hydration = clamp(player.Hydration-ailment.HydrationPenalty, 0, 100)
			player.Morale = clamp(player.Morale-ailment.MoralePenalty, 0, 100)
			gotIll = true
		}
	}
	refreshEffectBars(player)

	return EatResult{
		PlayerID:       playerID,
		ItemID:         itemID,
		ConsumedGrams:  grams,
		Nutrition:      nutrition,
		EnergyDelta:    energyGain,
		HydrationDelta: hydrationGain,
		MoraleDelta:    moraleGain,
		GotIll:         gotIll,
	}, nil
}

func containsStringFold(values []string, value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	for _, candidate := range values {
		if strings.TrimSpace(strings.ToLower(candidate)) == value {
			return true
		}
	}
	return false
}

func foodDecayWeatherMultiplier(weather WeatherState) float64 {
	mult := 1.0
	if weather.TemperatureC >= 28 {
		mult += 0.35
	}
	if weather.TemperatureC >= 35 {
		mult += 0.25
	}
	if weather.Type == WeatherRain {
		mult += 0.2
	}
	if weather.Type == WeatherStorm {
		mult += 0.35
	}
	if weather.Type == WeatherSnow || weather.TemperatureC <= -2 {
		mult -= 0.3
	}
	return clampFloat(mult, 0.5, 2.2)
}

func degradeFoodInventory(items []InventoryItem, weather WeatherState) ([]InventoryItem, float64) {
	if len(items) == 0 {
		return items, 0
	}
	weatherMult := foodDecayWeatherMultiplier(weather)
	spoiledKg := 0.0
	for i := range items {
		item := &items[i]
		if item.Qty <= 0 {
			continue
		}
		itemID := strings.ToLower(strings.TrimSpace(item.ID))
		if strings.HasSuffix(itemID, "_carcass") {
			item.AgeDays++
			lossPct := 0.1 * weatherMult
			if item.AgeDays > 1 {
				lossPct += 0.22 * weatherMult
			}
			loss := math.Round(clampFloat(item.Qty*lossPct, 0, item.Qty)*100) / 100
			if loss > 0 {
				item.Qty = normalizeInventoryQty(item.Unit, item.Qty-loss)
				spoiledKg += math.Round(loss*0.55*100) / 100
			}
			continue
		}

		spec, ok := foodItemCatalog[itemID]
		if !ok || !spec.Perishable {
			continue
		}
		item.AgeDays++
		if spec.ShelfLifeDays <= 0 || item.AgeDays <= spec.ShelfLifeDays {
			continue
		}
		lossPct := spec.DecayPerDay * weatherMult
		if item.AgeDays > spec.ShelfLifeDays+3 {
			lossPct += 0.1
		}
		loss := math.Round(clampFloat(item.Qty*lossPct, 0, item.Qty)*100) / 100
		if loss <= 0 {
			continue
		}
		item.Qty = normalizeInventoryQty(item.Unit, item.Qty-loss)
		if itemID != "spoiled_meat" {
			spoiledKg += math.Round(loss*0.9*100) / 100
		}
	}
	filtered := make([]InventoryItem, 0, len(items))
	for _, item := range items {
		if item.Qty <= 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, math.Round(spoiledKg*100) / 100
}

func (s *RunState) advanceFoodDegradation() {
	if s == nil {
		return
	}
	camp, spoiledCamp := degradeFoodInventory(s.CampInventory, s.Weather)
	s.CampInventory = camp
	if spoiledCamp > 0 {
		s.CampInventory = addOrMergeInventory(s.CampInventory, InventoryItem{
			ID:       "spoiled_meat",
			Name:     "Spoiled Meat",
			Unit:     "kg",
			Qty:      spoiledCamp,
			WeightKg: 1,
			Category: "food",
			AgeDays:  0,
		})
	}
	for i := range s.Players {
		player := &s.Players[i]
		personal, spoiledPersonal := degradeFoodInventory(player.PersonalItems, s.Weather)
		player.PersonalItems = personal
		if spoiledPersonal > 0 {
			player.PersonalItems = addOrMergeInventory(player.PersonalItems, InventoryItem{
				ID:       "spoiled_meat",
				Name:     "Spoiled Meat",
				Unit:     "kg",
				Qty:      spoiledPersonal,
				WeightKg: 1,
				Category: "food",
				AgeDays:  0,
			})
		}
	}
}
