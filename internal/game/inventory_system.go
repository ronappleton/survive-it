package game

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

// Discovery summary:
// - Camp storage is currently derived from shelter type plus crafted container bonuses.
// - Inventory stacks are merged/consumed by ID with age-aware depletion, so capacity hooks are lightweight.
// - Shelter-stage storage can be integrated safely by swapping the shelter capacity lookup only.
type InventoryItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Qty      float64 `json:"qty"`
	WeightKg float64 `json:"weight_kg"` // Per-unit weight.
	Category string  `json:"category,omitempty"`
	Quality  string  `json:"quality,omitempty"`
	AgeDays  int     `json:"age_days,omitempty"`
}

func normalizeInventoryQty(unit string, qty float64) float64 {
	if qty <= 0 {
		return 0
	}
	if strings.EqualFold(strings.TrimSpace(unit), "kg") {
		return math.Round(qty*10) / 10
	}
	return math.Round(qty)
}

func formatInventoryQty(unit string, qty float64) string {
	if strings.EqualFold(strings.TrimSpace(unit), "kg") {
		return fmt.Sprintf("%.1f%s", qty, unit)
	}
	return fmt.Sprintf("%.0f%s", qty, unit)
}

func defaultUnitWeightKg(unit string) float64 {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "kg":
		return 1.0
	case "bundle":
		return 0.35
	case "sheet":
		return 0.16
	case "stick":
		return 0.12
	case "chunk":
		return 0.2
	case "lump":
		return 0.08
	case "piece":
		return 0.07
	case "set":
		return 0.3
	case "hook":
		return 0.02
	case "hide":
		return 1.4
	case "roll":
		return 0.12
	default:
		return 0.2
	}
}

func (s *RunState) campCapacityKg() float64 {
	if s == nil {
		return 0
	}
	// Minimal ground cache before shelter.
	capKg := 8.0
	if s.Shelter.Type != "" && s.Shelter.Durability > 0 {
		if metrics, ok := s.currentShelterMetrics(); ok && metrics.StorageCapacityKg > 0 {
			capKg = metrics.StorageCapacityKg
		} else if shelter, ok := shelterByID(s.Shelter.Type); ok && shelter.StorageCapacityKg > 0 {
			capKg = shelter.StorageCapacityKg
		}
	}
	if slices.Contains(s.CraftedItems, "split_basket") {
		capKg += 6
	}
	if slices.Contains(s.CraftedItems, "bark_container") {
		capKg += 3
	}
	return capKg
}

func (s *RunState) campUsedKg() float64 {
	if s == nil {
		return 0
	}
	used := 0.0
	for _, wood := range s.WoodStock {
		used += max(0, wood.Kg)
	}
	for _, stock := range s.ResourceStock {
		spec, ok := s.findResourceForBiome(stock.ID)
		if ok {
			used += max(0, stock.Qty) * defaultUnitWeightKg(spec.Unit)
			continue
		}
		used += max(0, stock.Qty) * defaultUnitWeightKg(stock.Unit)
	}
	for _, item := range s.CampInventory {
		if item.Qty <= 0 {
			continue
		}
		w := item.WeightKg
		if w <= 0 {
			w = defaultUnitWeightKg(item.Unit)
		}
		used += item.Qty * w
	}
	return used
}

func (s *RunState) campFreeKg() float64 {
	return max(0.0, s.campCapacityKg()-s.campUsedKg())
}

func (s *RunState) canStoreAtCamp(weightKg float64) bool {
	if s == nil {
		return false
	}
	if weightKg <= 0 {
		return true
	}
	return s.campUsedKg()+weightKg <= s.campCapacityKg()+1e-9
}

func addOrMergeInventory(items []InventoryItem, item InventoryItem) []InventoryItem {
	item.ID = strings.ToLower(strings.TrimSpace(item.ID))
	if item.ID == "" {
		return items
	}
	item.Qty = normalizeInventoryQty(item.Unit, item.Qty)
	if item.Qty <= 0 {
		return items
	}
	if item.WeightKg <= 0 {
		item.WeightKg = defaultUnitWeightKg(item.Unit)
	}
	for i := range items {
		if items[i].ID != item.ID {
			continue
		}
		if items[i].AgeDays != item.AgeDays {
			continue
		}
		if strings.TrimSpace(items[i].Quality) != strings.TrimSpace(item.Quality) {
			continue
		}
		items[i].Qty = normalizeInventoryQty(items[i].Unit, items[i].Qty+item.Qty)
		if strings.TrimSpace(items[i].Name) == "" {
			items[i].Name = item.Name
		}
		if strings.TrimSpace(items[i].Category) == "" {
			items[i].Category = item.Category
		}
		if strings.TrimSpace(item.Quality) != "" {
			items[i].Quality = item.Quality
		}
		return items
	}
	return append(items, item)
}

func consumeInventory(items []InventoryItem, id string, qty float64) ([]InventoryItem, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" || qty <= 0 {
		return items, false
	}
	indexes := make([]int, 0, len(items))
	available := 0.0
	for i := range items {
		if items[i].ID != id {
			continue
		}
		indexes = append(indexes, i)
		available += items[i].Qty
	}
	if len(indexes) == 0 || available+1e-9 < qty {
		return items, false
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		if items[indexes[i]].AgeDays == items[indexes[j]].AgeDays {
			return indexes[i] < indexes[j]
		}
		// Consume oldest stacks first.
		return items[indexes[i]].AgeDays > items[indexes[j]].AgeDays
	})
	remaining := qty
	for _, idx := range indexes {
		if remaining <= 0 {
			break
		}
		take := min(items[idx].Qty, remaining)
		items[idx].Qty = normalizeInventoryQty(items[idx].Unit, items[idx].Qty-take)
		remaining -= take
	}
	if remaining > 1e-6 {
		return items, false
	}
	filtered := make([]InventoryItem, 0, len(items))
	for i := range items {
		if items[i].Qty <= 0 {
			continue
		}
		filtered = append(filtered, items[i])
	}
	return filtered, true
}

func inventoryTotalQtyByID(items []InventoryItem, id string) float64 {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return 0
	}
	total := 0.0
	for _, item := range items {
		if item.ID != id {
			continue
		}
		total += max(0, item.Qty)
	}
	return total
}

func inventoryItemByID(items []InventoryItem, id string) (InventoryItem, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return InventoryItem{}, false
	}
	found := false
	var best InventoryItem
	for _, item := range items {
		if item.ID != id {
			continue
		}
		if !found || item.AgeDays > best.AgeDays {
			best = item
			found = true
		}
	}
	return best, found
}

func inventoryWeightKg(items []InventoryItem) float64 {
	total := 0.0
	for _, item := range items {
		if item.Qty <= 0 {
			continue
		}
		w := item.WeightKg
		if w <= 0 {
			w = defaultUnitWeightKg(item.Unit)
		}
		total += item.Qty * w
	}
	return total
}

func deriveCarryLimitKg(player PlayerState, hasPackFrame bool) float64 {
	carry := 3.2
	carry += float64(clamp(player.Strength, -3, 3)) * 1.6
	carry += float64(clamp(player.Endurance, -3, 3)) * 0.8
	carry += float64(clamp(player.Agility, -3, 3)) * 0.3
	carry += float64(player.Gathering+player.Crafting) / 120.0
	carry += float64(sumTraitModifier(player.Traits)) * 0.35
	if hasPackFrame {
		carry += 6
	}
	return clampFloat(carry, 2.0, 22.0)
}

func (s *RunState) playerCarryLimitKg(player *PlayerState) float64 {
	if player == nil {
		return 0
	}
	hasPack := slices.Contains(s.CraftedItems, "pack_frame")
	limit := deriveCarryLimitKg(*player, hasPack)
	if player.CarryLimitKg > 0 {
		limit = player.CarryLimitKg
		if hasPack {
			limit += 6
		}
	}
	return clampFloat(limit, 2.0, 28.0)
}

func (s *RunState) AddPersonalInventoryItem(playerID int, item InventoryItem) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return fmt.Errorf("player %d not found", playerID)
	}
	item.Qty = normalizeInventoryQty(item.Unit, item.Qty)
	if item.Qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if item.WeightKg <= 0 {
		item.WeightKg = defaultUnitWeightKg(item.Unit)
	}
	nextWeight := inventoryWeightKg(player.PersonalItems) + item.Qty*item.WeightKg
	limit := s.playerCarryLimitKg(player)
	if nextWeight > limit+1e-9 {
		return fmt.Errorf("carry limit exceeded (%.1f/%.1fkg)", nextWeight, limit)
	}
	player.PersonalItems = addOrMergeInventory(player.PersonalItems, item)
	player.CarryLimitKg = deriveCarryLimitKg(*player, slices.Contains(s.CraftedItems, "pack_frame"))
	return nil
}

func (s *RunState) removePersonalInventoryItem(playerID int, itemID string, qty float64) (InventoryItem, error) {
	player, ok := s.playerByID(playerID)
	if !ok {
		return InventoryItem{}, fmt.Errorf("player %d not found", playerID)
	}
	current, ok := inventoryItemByID(player.PersonalItems, itemID)
	if !ok {
		return InventoryItem{}, fmt.Errorf("item not in personal inventory: %s", itemID)
	}
	totalQty := inventoryTotalQtyByID(player.PersonalItems, itemID)
	if qty <= 0 {
		qty = totalQty
	}
	if qty > totalQty+1e-9 {
		return InventoryItem{}, fmt.Errorf("not enough quantity of %s", itemID)
	}
	var consumed bool
	player.PersonalItems, consumed = consumeInventory(player.PersonalItems, itemID, qty)
	if !consumed {
		return InventoryItem{}, fmt.Errorf("not enough quantity of %s", itemID)
	}
	current.Qty = normalizeInventoryQty(current.Unit, qty)
	return current, nil
}

func (s *RunState) addCampInventoryItem(item InventoryItem) error {
	if s == nil {
		return fmt.Errorf("run state is nil")
	}
	item.Qty = normalizeInventoryQty(item.Unit, item.Qty)
	if item.Qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if item.WeightKg <= 0 {
		item.WeightKg = defaultUnitWeightKg(item.Unit)
	}
	if !s.canStoreAtCamp(item.Qty * item.WeightKg) {
		return fmt.Errorf("camp inventory full (%.1f/%.1fkg)", s.campUsedKg(), s.campCapacityKg())
	}
	s.CampInventory = addOrMergeInventory(s.CampInventory, item)
	return nil
}

func (s *RunState) removeCampInventoryItem(itemID string, qty float64) (InventoryItem, error) {
	if s == nil {
		return InventoryItem{}, fmt.Errorf("run state is nil")
	}
	item, ok := inventoryItemByID(s.CampInventory, itemID)
	if !ok {
		return InventoryItem{}, fmt.Errorf("item not found in camp inventory: %s", itemID)
	}
	totalQty := inventoryTotalQtyByID(s.CampInventory, itemID)
	if qty <= 0 {
		qty = totalQty
	}
	if qty > totalQty+1e-9 {
		return InventoryItem{}, fmt.Errorf("not enough quantity of %s", itemID)
	}
	var consumed bool
	s.CampInventory, consumed = consumeInventory(s.CampInventory, itemID, qty)
	if !consumed {
		return InventoryItem{}, fmt.Errorf("not enough quantity of %s", itemID)
	}
	item.Qty = normalizeInventoryQty(item.Unit, qty)
	return item, nil
}

func (s *RunState) StashPersonalItem(playerID int, itemID string, qty float64) (InventoryItem, error) {
	item, err := s.removePersonalInventoryItem(playerID, itemID, qty)
	if err != nil {
		return InventoryItem{}, err
	}
	if err := s.addCampInventoryItem(item); err != nil {
		_ = s.AddPersonalInventoryItem(playerID, item)
		return InventoryItem{}, err
	}
	return item, nil
}

func (s *RunState) TakeCampItem(playerID int, itemID string, qty float64) (InventoryItem, error) {
	item, err := s.removeCampInventoryItem(itemID, qty)
	if err != nil {
		return InventoryItem{}, err
	}
	if err := s.AddPersonalInventoryItem(playerID, item); err != nil {
		_ = s.addCampInventoryItem(item)
		return InventoryItem{}, err
	}
	return item, nil
}

func formatInventoryList(items []InventoryItem) string {
	if len(items) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		label := item.ID
		if strings.TrimSpace(item.Name) != "" {
			label = item.Name
		}
		qty := formatInventoryQty(item.Unit, item.Qty)
		extra := ""
		if strings.TrimSpace(item.Quality) != "" {
			extra = " " + item.Quality
		}
		if item.AgeDays > 0 {
			extra += fmt.Sprintf(" d%dd", item.AgeDays)
		}
		parts = append(parts, fmt.Sprintf("%s %s%s", item.ID, qty, extra))
		if strings.TrimSpace(item.Name) != "" && item.Name != label {
			parts[len(parts)-1] = fmt.Sprintf("%s(%s)%s", item.ID, qty, extra)
		}
	}
	return strings.Join(parts, ", ")
}

func (s *RunState) CampInventorySummary() string {
	if s == nil {
		return "camp inventory unavailable"
	}
	capKg := s.campCapacityKg()
	used := s.campUsedKg()
	woods := formatWoodStock(s.WoodStock)
	resources := formatResourceStock(s.ResourceStock)
	extras := formatInventoryList(s.CampInventory)
	return fmt.Sprintf("Camp %.1f/%.1fkg | wood: %s | resources: %s | extras: %s", used, capKg, woods, resources, extras)
}

func (s *RunState) PersonalInventorySummary(playerID int) string {
	if s == nil {
		return "personal inventory unavailable"
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return fmt.Sprintf("player %d not found", playerID)
	}
	used := inventoryWeightKg(player.PersonalItems)
	limit := s.playerCarryLimitKg(player)
	return fmt.Sprintf("P%d %.1f/%.1fkg | %s", playerID, used, limit, formatInventoryList(player.PersonalItems))
}

func (s *RunState) AdvanceActionClock(hours float64) int {
	if s == nil || hours <= 0 {
		return 0
	}
	minutes := int(math.Round(hours * 60))
	if minutes <= 0 {
		minutes = 1
	}
	return s.AdvanceMinutes(minutes)
}

func (s *RunState) AdvanceMinutes(minutes int) int {
	if s == nil || minutes <= 0 {
		return 0
	}
	s.EnsurePlayerRuntimeStats()
	daysAdvanced := 0
	remaining := minutes
	for remaining > 0 {
		minUntilMidnight := int(math.Ceil((24.0 - s.ClockHours) * 60.0))
		if minUntilMidnight <= 0 {
			minUntilMidnight = 1
		}
		stepMinutes := remaining
		if stepMinutes > minUntilMidnight {
			stepMinutes = minUntilMidnight
		}

		fraction := float64(stepMinutes) / 1440.0
		for i := range s.Players {
			applyMetabolismFraction(&s.Players[i], fraction)
			applyPhysiologyFraction(&s.Players[i], fraction)
		}
		s.MetabolismProgress = clampFloat(s.MetabolismProgress+fraction, 0, 1)
		s.ClockHours += float64(stepMinutes) / 60.0
		remaining -= stepMinutes

		for s.ClockHours >= 24.0 {
			s.ClockHours -= 24.0
			s.AdvanceDay()
			daysAdvanced++
		}
	}
	return daysAdvanced
}

func formatClockFromHours(hours float64) string {
	if hours < 0 {
		hours = 0
	}
	whole := int(math.Floor(hours)) % 24
	min := int(math.Round((hours - math.Floor(hours)) * 60))
	if min >= 60 {
		min -= 60
		whole = (whole + 1) % 24
	}
	return fmt.Sprintf("%02d:%02d", whole, min)
}
