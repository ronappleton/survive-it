package game

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Discovery summary:
// - Run command handlers are thin adapters over RunState systems (forage/craft/trap/travel/fire/shelter).
// - Shelter status/build output is generated here, so staged shelter progress is surfaced without loop refactors.
// - Keeping command signatures stable preserves parser/UI integration while extending simulation depth.
type RunCommandResult struct {
	Handled       bool
	Message       string
	HoursAdvanced float64
}

type equipmentAction struct {
	ID          string
	Aliases     []string
	Description string
	EnergyDelta int
	Hydration   int
	MoraleDelta int
	Nutrition   NutritionTotals
	Special     string
}

const (
	specialTreatAilment = "treat_ailment"
)

func (s *RunState) ExecuteRunCommand(raw string) RunCommandResult {
	command := strings.TrimSpace(strings.ToLower(raw))
	if command == "" {
		return RunCommandResult{Handled: false}
	}
	s.EnsurePlayerRuntimeStats()
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return RunCommandResult{Handled: false}
	}

	switch fields[0] {
	case "commands", "help":
		return RunCommandResult{
			Handled: true,
			Message: "Commands: look [left|right|front|back], look closer at <plants|trees|insects|water>, hunt land|fish|air [p#], fish [p#], forage [roots|berries|fruits|vegetables|any] [p#] [grams], trees, plants, wood gather|dry|stock [kg] [p#], resources, collect <resource|any> [qty] [p#], bark strip [tree|any] [qty] [p#], inventory camp|personal|stash|take|add|drop [..], trap list|set|status|check [..], gut <carcass> [kg] [p#], cook <raw_meat> [kg] [p#], preserve <smoke|dry|salt> <meat> [kg] [p#], eat <food_item> [grams|kg] [p#], go <n|s|e|w> [km] [p#], fire status|methods|prep|ember|ignite|build|tend|out, shelter list|build|status, craft list|make|inventory, actions [p#], use <item> <action> [p#], ask <player> <task>, next, save, load, menu.",
		}
	case "look", "inspect", "examine":
		return s.executeLookCommand(fields[0], fields[1:])
	case "hunt", "catch":
		return s.executeHuntCommand(fields[1:])
	case "fish":
		return s.executeFishCommand(fields[1:])
	case "forage":
		return s.executeForageCommand(fields[1:])
	case "trees":
		return s.executeTreesCommand()
	case "plants":
		return s.executePlantsCommand()
	case "resources":
		return s.executeResourcesCommand()
	case "collect":
		return s.executeCollectCommand(fields[1:])
	case "bark":
		return s.executeBarkCommand(fields[1:])
	case "inventory":
		return s.executeInventoryCommand(fields[1:])
	case "wood":
		return s.executeWoodCommand(fields[1:])
	case "trap":
		return s.executeTrapCommand(fields[1:])
	case "gut":
		return s.executeGutCommand(fields[1:])
	case "cook":
		return s.executeCookCommand(fields[1:])
	case "preserve":
		return s.executePreserveCommand(fields[1:])
	case "smoke", "dry", "salt":
		return s.executePreserveCommand(fields)
	case "eat":
		return s.executeEatCommand(fields[1:])
	case "go":
		return s.executeGoCommand(fields[1:])
	case "fire":
		return s.executeFireCommand(fields[1:])
	case "shelter":
		return s.executeShelterCommand(fields[1:])
	case "craft":
		return s.executeCraftCommand(fields[1:])
	case "actions":
		playerID := 1
		if len(fields) > 1 {
			if parsed := parsePlayerToken(fields[1]); parsed > 0 {
				playerID = parsed
			}
		}
		return s.listActionsForPlayer(playerID)
	case "use":
		return s.executeUseCommand(fields)
	case "ask":
		return s.executeAskCommand(fields[1:])
	default:
		return RunCommandResult{Handled: false}
	}
}

func (s *RunState) executeAskCommand(args []string) RunCommandResult {
	if len(args) < 2 {
		return RunCommandResult{Handled: true, Message: "Usage: ask <player> <task>"}
	}
	target := strings.TrimSpace(args[0])
	task := strings.TrimSpace(strings.Join(args[1:], " "))
	if task == "" {
		return RunCommandResult{Handled: true, Message: "Usage: ask <player> <task>"}
	}
	name := target
	if pid := parsePlayerToken(target); pid > 0 {
		if player, ok := s.playerByID(pid); ok {
			name = player.Name
		}
	}
	return RunCommandResult{Handled: true, Message: fmt.Sprintf("%s isn't ready for delegated tasks yet (%s).", name, task)}
}

func (s *RunState) listActionsForPlayer(playerID int) RunCommandResult {
	player, ok := s.playerByID(playerID)
	if !ok {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Player %d not found.", playerID)}
	}

	kit := uniqueKitItems(player.Kit, s.Config.IssuedKit)
	if len(kit) == 0 {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d has no kit actions available.", playerID)}
	}

	sort.Slice(kit, func(i, j int) bool { return string(kit[i]) < string(kit[j]) })

	parts := make([]string, 0, len(kit))
	for _, item := range kit {
		actions := actionsForItem(item)
		if len(actions) == 0 {
			continue
		}
		actionNames := make([]string, 0, len(actions))
		for _, action := range actions {
			actionNames = append(actionNames, action.ID)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", itemCommandLabel(item), strings.Join(actionNames, ",")))
	}

	if len(parts) == 0 {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d has no registered equipment actions.", playerID)}
	}

	return RunCommandResult{
		Handled: true,
		Message: fmt.Sprintf("P%d actions -> %s", playerID, strings.Join(parts, " | ")),
	}
}

func (s *RunState) executeUseCommand(fields []string) RunCommandResult {
	if len(fields) < 3 {
		return RunCommandResult{Handled: true, Message: "Usage: use <item> <action> [p#]"}
	}

	playerID, tokens := extractPlayerID(fields[1:])
	if len(tokens) < 2 {
		return RunCommandResult{Handled: true, Message: "Usage: use <item> <action> [p#]"}
	}

	item, actionInput, ok := parseItemAndAction(tokens)
	if !ok {
		return RunCommandResult{Handled: true, Message: "Could not parse item/action. Use: actions [p#] to list options."}
	}

	player, ok := s.playerByID(playerID)
	if !ok {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Player %d not found.", playerID)}
	}
	if !playerHasKitItem(player, s.Config.IssuedKit, item) {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("%s not in player %d kit.", item, playerID)}
	}

	action, ok := findAction(item, actionInput)
	if !ok {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Unknown action for %s. Use: actions p%d", itemCommandLabel(item), playerID)}
	}

	player.Energy = clamp(player.Energy+action.EnergyDelta, 0, 100)
	player.Hydration = clamp(player.Hydration+action.Hydration, 0, 100)
	player.Morale = clamp(player.Morale+action.MoraleDelta, 0, 100)

	totalEnergyDelta := action.EnergyDelta
	totalHydrationDelta := action.Hydration
	totalMoraleDelta := action.MoraleDelta

	if action.Nutrition.CaloriesKcal > 0 || action.Nutrition.ProteinG > 0 || action.Nutrition.FatG > 0 || action.Nutrition.SugarG > 0 {
		player.Nutrition = player.Nutrition.add(action.Nutrition)
		applyMealNutritionReserves(player, action.Nutrition)
		energyBonus, hydrationBonus, moraleBonus := nutritionToPlayerEffects(action.Nutrition)
		player.Energy = clamp(player.Energy+energyBonus, 0, 100)
		player.Hydration = clamp(player.Hydration+hydrationBonus, 0, 100)
		player.Morale = clamp(player.Morale+moraleBonus, 0, 100)
		totalEnergyDelta += energyBonus
		totalHydrationDelta += hydrationBonus
		totalMoraleDelta += moraleBonus
	}

	specialMsg := ""
	if action.Special == specialTreatAilment {
		if len(player.Ailments) > 0 {
			removed := player.Ailments[0]
			player.Ailments = append([]Ailment{}, player.Ailments[1:]...)
			player.Morale = clamp(player.Morale+1, 0, 100)
			totalMoraleDelta++
			specialMsg = fmt.Sprintf(" | treated: %s", removed.Name)
		} else {
			specialMsg = " | no active ailments to treat"
		}
	}

	msg := fmt.Sprintf("P%d used %s -> %s. %+dE %+dH2O %+dM",
		playerID, itemCommandLabel(item), action.ID, totalEnergyDelta, totalHydrationDelta, totalMoraleDelta)
	if action.Nutrition.CaloriesKcal > 0 || action.Nutrition.ProteinG > 0 || action.Nutrition.FatG > 0 || action.Nutrition.SugarG > 0 {
		msg += fmt.Sprintf(" | +%dkcal +%dgP +%dgF +%dgS", action.Nutrition.CaloriesKcal, action.Nutrition.ProteinG, action.Nutrition.FatG, action.Nutrition.SugarG)
	}
	msg += specialMsg
	return RunCommandResult{Handled: true, Message: msg}
}

func (s *RunState) executeHuntCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: hunt <land|fish|air> [p#]"}
	}
	playerID := 1
	domain := AnimalDomainLand
	foundDomain := false
	for _, token := range fields {
		if parsed := parsePlayerToken(token); parsed > 0 {
			playerID = parsed
			continue
		}
		switch strings.ToLower(strings.TrimSpace(token)) {
		case "land":
			domain = AnimalDomainLand
			foundDomain = true
		case "fish", "water":
			domain = AnimalDomainWater
			foundDomain = true
		case "air", "bird":
			domain = AnimalDomainAir
			foundDomain = true
		case "raw", "liver":
			// Legacy tokens from auto-consume flow; intentionally ignored.
			continue
		default:
			if _, err := strconv.ParseFloat(token, 64); err == nil {
				continue
			}
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Unknown hunt option: %s", token)}
		}
	}
	if !foundDomain {
		return RunCommandResult{Handled: true, Message: "Usage: hunt <land|fish|air> [p#]"}
	}
	actionType := "hunt"
	if domain == AnimalDomainWater {
		actionType = "fish"
	}
	result, err := s.HuntAndCollectCarcass(playerID, domain, actionType)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Hunt failed: %v", err)}
	}
	encounterText := ""
	if len(result.EncounterLogs) > 0 {
		encounterText = " | " + strings.Join(result.EncounterLogs, " ")
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("P%d hunted %s (%.1fkg) -> %.1fkg %s stored in %s. Process with: gut %s [kg] p%d",
			playerID, result.AnimalName, float64(result.WeightGrams)/1000.0, result.CarcassKg, result.CarcassID, result.StoredAt, result.CarcassID, playerID) + encounterText,
	}
}

func (s *RunState) executeFishCommand(fields []string) RunCommandResult {
	playerID := 1
	for _, field := range fields {
		if parsed := parsePlayerToken(field); parsed > 0 {
			playerID = parsed
		}
	}
	result, err := s.HuntAndCollectCarcass(playerID, AnimalDomainWater, "fish")
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fish failed: %v", err)}
	}
	encounterText := ""
	if len(result.EncounterLogs) > 0 {
		encounterText = " | " + strings.Join(result.EncounterLogs, " ")
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("P%d fished %s (%.1fkg) -> %.1fkg %s stored in %s. Process with: gut %s [kg] p%d",
			playerID, result.AnimalName, float64(result.WeightGrams)/1000.0, result.CarcassKg, result.CarcassID, result.StoredAt, result.CarcassID, playerID) + encounterText,
	}
}

func (s *RunState) executeForageCommand(fields []string) RunCommandResult {
	playerID := 1
	category := PlantCategoryAny
	grams := 0

	for _, field := range fields {
		if parsed := parsePlayerToken(field); parsed > 0 {
			playerID = parsed
			continue
		}
		if parsedGrams, err := strconv.Atoi(field); err == nil && parsedGrams > 0 {
			grams = parsedGrams
			continue
		}
		category = ParsePlantCategory(field)
	}

	result, err := s.ForageAndConsume(playerID, category, grams)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Forage failed: %v", err)}
	}
	x, y := s.CurrentMapPosition()
	s.applyCellStateAction(x, y, "forage")
	encounterMsg := ""
	if event, ok := s.RollWildlifeEncounter(playerID, x, y, "forage", 0); ok {
		if player, ok := s.playerByID(playerID); ok {
			player.Energy = clamp(player.Energy+event.EnergyDelta, 0, 100)
			player.Hydration = clamp(player.Hydration+event.HydrationDelta, 0, 100)
			player.Morale = clamp(player.Morale+event.MoraleDelta, 0, 100)
			refreshEffectBars(player)
		}
		encounterMsg = " | " + event.Message
	}

	return RunCommandResult{
		Handled: true,
		Message: fmt.Sprintf("P%d foraged %dg %s: %dkcal %dgP %dgF %dgS",
			playerID,
			result.HarvestGrams,
			result.Plant.Name,
			result.Nutrition.CaloriesKcal,
			result.Nutrition.ProteinG,
			result.Nutrition.FatG,
			result.Nutrition.SugarG,
		) + encounterMsg,
	}
}

func (s *RunState) executeTreesCommand() RunCommandResult {
	biome := s.CurrentBiomeQuery()
	if strings.TrimSpace(biome) == "" {
		biome = s.Scenario.Biome
	}
	trees := TreesForBiome(biome)
	if len(trees) == 0 {
		return RunCommandResult{Handled: true, Message: "No tree resources registered for this biome."}
	}
	parts := make([]string, 0, len(trees))
	for _, tree := range trees {
		parts = append(parts, fmt.Sprintf("%s(%s)", tree.Name, tree.WoodType))
	}
	return RunCommandResult{Handled: true, Message: "Trees -> " + strings.Join(parts, ", ")}
}

func (s *RunState) executePlantsCommand() RunCommandResult {
	biome := s.CurrentBiomeQuery()
	if strings.TrimSpace(biome) == "" {
		biome = s.Scenario.Biome
	}
	available := ResourcesForBiome(biome)
	if len(available) == 0 {
		return RunCommandResult{Handled: true, Message: "No utility plant resources available."}
	}
	parts := make([]string, 0, len(available))
	for _, resource := range available {
		if len(resource.Uses) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s[%s]", resource.ID, strings.Join(resource.Uses, "/")))
	}
	if len(parts) == 0 {
		return RunCommandResult{Handled: true, Message: "No utility plant resources available."}
	}
	return RunCommandResult{Handled: true, Message: "Utility plants/materials -> " + strings.Join(parts, ", ")}
}

func (s *RunState) executeResourcesCommand() RunCommandResult {
	biome := s.CurrentBiomeQuery()
	if strings.TrimSpace(biome) == "" {
		biome = s.Scenario.Biome
	}
	available := ResourcesForBiome(biome)
	if len(available) == 0 {
		return RunCommandResult{Handled: true, Message: "No biome resources available."}
	}
	parts := make([]string, 0, len(available))
	for _, resource := range available {
		if len(resource.Uses) == 0 {
			parts = append(parts, resource.ID)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s{%s}", resource.ID, strings.Join(resource.Uses, "/")))
	}
	return RunCommandResult{
		Handled: true,
		Message: fmt.Sprintf("Resources -> %s | Stock: %s", strings.Join(parts, ", "), s.CampInventorySummary()),
	}
}

func (s *RunState) executeBarkCommand(fields []string) RunCommandResult {
	if len(fields) == 0 || fields[0] != "strip" {
		return RunCommandResult{Handled: true, Message: "Usage: bark strip [tree|any] [qty] [p#]"}
	}
	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
	if !hasAmount {
		amount = 1
	}
	treeID := "any"
	if len(rest) > 0 {
		treeID = rest[0]
	}
	result, err := s.StripBark(playerID, treeID, amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Bark stripping failed: %v", err)}
	}
	s.AdvanceActionClock(result.HoursSpent)
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("P%d stripped %s: +%.0f %s +%.0f inner_bark_fiber (%s quality, %.1fh). Stock: %s",
			playerID, result.Tree.Name, result.PrimaryQty, result.Primary.ID, result.FiberQty, result.Quality, result.HoursSpent, s.CampInventorySummary()),
	}
}

func (s *RunState) executeInventoryCommand(fields []string) RunCommandResult {
	if len(fields) == 0 || fields[0] == "status" {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("%s | %s", s.CampInventorySummary(), s.PersonalInventorySummary(1))}
	}
	switch fields[0] {
	case "camp":
		return RunCommandResult{Handled: true, Message: s.CampInventorySummary()}
	case "personal":
		playerID := 1
		if len(fields) > 1 {
			if parsed := parsePlayerToken(fields[1]); parsed > 0 {
				playerID = parsed
			}
		}
		return RunCommandResult{Handled: true, Message: s.PersonalInventorySummary(playerID)}
	case "stash":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory stash <item_id> [qty] [p#]"}
		}
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if len(rest) == 0 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory stash <item_id> [qty] [p#]"}
		}
		if !hasAmount {
			amount = 1
		}
		item, err := s.StashPersonalItem(playerID, rest[0], amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Inventory stash failed: %v", err)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d stashed %s %s. %s", playerID, item.ID, formatInventoryQty(item.Unit, item.Qty), s.CampInventorySummary())}
	case "take":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory take <item_id> [qty] [p#]"}
		}
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if len(rest) == 0 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory take <item_id> [qty] [p#]"}
		}
		if !hasAmount {
			amount = 1
		}
		item, err := s.TakeCampItem(playerID, rest[0], amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Inventory take failed: %v", err)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d took %s %s. %s", playerID, item.ID, formatInventoryQty(item.Unit, item.Qty), s.PersonalInventorySummary(playerID))}
	case "add":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory add <item_id> [qty] [p#]"}
		}
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if len(rest) == 0 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory add <item_id> [qty] [p#]"}
		}
		if !hasAmount {
			amount = 1
		}
		itemID := strings.ToLower(strings.TrimSpace(rest[0]))
		unit := "piece"
		weight := defaultUnitWeightKg(unit)
		if spec, ok := s.findResourceForBiome(itemID); ok {
			unit = spec.Unit
			weight = defaultUnitWeightKg(spec.Unit)
		}
		item := InventoryItem{
			ID:       itemID,
			Name:     strings.ReplaceAll(itemID, "_", " "),
			Unit:     unit,
			Qty:      amount,
			WeightKg: weight,
			Category: "field",
		}
		if err := s.AddPersonalInventoryItem(playerID, item); err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Inventory add failed: %v", err)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d added %s %s to personal inventory. %s", playerID, item.ID, formatInventoryQty(item.Unit, item.Qty), s.PersonalInventorySummary(playerID))}
	case "drop":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory drop <item_id> [qty] [p#]"}
		}
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if len(rest) == 0 {
			return RunCommandResult{Handled: true, Message: "Usage: inventory drop <item_id> [qty] [p#]"}
		}
		if !hasAmount {
			amount = 1
		}
		item, err := s.removePersonalInventoryItem(playerID, rest[0], amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Inventory drop failed: %v", err)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d dropped %s %s. %s", playerID, item.ID, formatInventoryQty(item.Unit, item.Qty), s.PersonalInventorySummary(playerID))}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: inventory camp|personal [p#] | inventory stash <item_id> [qty] [p#] | inventory take <item_id> [qty] [p#] | inventory add <item_id> [qty] [p#] | inventory drop <item_id> [qty] [p#]"}
	}
}

func (s *RunState) executeTrapCommand(fields []string) RunCommandResult {
	if len(fields) == 0 || fields[0] == "status" {
		return RunCommandResult{Handled: true, Message: "Traps: " + formatTrapStatus(s.PlacedTraps)}
	}
	switch fields[0] {
	case "list":
		options := TrapsForBiome(s.Scenario.Biome)
		parts := make([]string, 0, len(options))
		for _, trap := range options {
			parts = append(parts, fmt.Sprintf("%s(%s %.0f%%)", trap.Name, trap.ID, trap.BaseChance*100))
		}
		return RunCommandResult{Handled: true, Message: "Trap options -> " + strings.Join(parts, ", ")}
	case "set":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: trap set <id> [p#]"}
		}
		playerID := 1
		for _, token := range fields[2:] {
			if parsed := parsePlayerToken(token); parsed > 0 {
				playerID = parsed
			}
		}
		setResult, err := s.SetTrap(playerID, fields[1])
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Trap set failed: %v", err)}
		}
		s.AdvanceActionClock(setResult.Hours)
		return RunCommandResult{
			Handled:       true,
			HoursAdvanced: setResult.Hours,
			Message: fmt.Sprintf("P%d set %s (%s quality). Base catch %.0f%%, setup %.1fh. Status: %s",
				playerID, setResult.Trap.Name, setResult.Quality, setResult.Chance*100, setResult.Hours, formatTrapStatus(s.PlacedTraps)),
		}
	case "check":
		check := s.CheckTraps()
		msg := fmt.Sprintf("Checked %d traps, collected %.1fkg, rearmed %d, broken %d.", check.Checked, check.CollectedKg, check.Rearmed, check.Broken)
		if check.CampOverflow > 0 {
			msg += fmt.Sprintf(" %d catches lost: camp storage full.", check.CampOverflow)
		}
		msg += " " + s.CampInventorySummary()
		return RunCommandResult{Handled: true, Message: msg}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: trap list | trap set <id> [p#] | trap status | trap check"}
	}
}

func (s *RunState) executeGutCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: gut <small_game_carcass|bird_carcass|fish_carcass> [kg] [p#]"}
	}
	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields)
	if len(rest) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: gut <small_game_carcass|bird_carcass|fish_carcass> [kg] [p#]"}
	}
	if !hasAmount {
		amount = 0
	}
	result, err := s.GutCarcass(playerID, rest[0], amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Gutting failed: %v", err)}
	}
	msg := fmt.Sprintf("P%d gutted %.2fkg %s -> %.2fkg %s, %.2fkg spoiled, %.2fkg inedible (%.1fh).",
		playerID, result.ProcessedKg, result.CarcassID, result.MeatKg, result.MeatID, result.SpoiledKg, result.InedibleKg, result.HoursSpent)
	if result.PiercedGut {
		msg += " Intestines pierced: contamination increased and some meat became inedible."
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message:       msg + " " + s.CampInventorySummary(),
	}
}

func (s *RunState) executeCookCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: cook <raw_small_game_meat|raw_bird_meat|raw_fish_meat> [kg] [p#]"}
	}
	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields)
	if len(rest) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: cook <raw_small_game_meat|raw_bird_meat|raw_fish_meat> [kg] [p#]"}
	}
	if !hasAmount {
		amount = 0
	}
	result, err := s.CookFood(playerID, rest[0], amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Cook failed: %v", err)}
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("P%d cooked %.2fkg %s -> %.2fkg %s (%.1fh). %s",
			playerID, result.InputKg, result.RawID, result.OutputKg, result.CookedID, result.HoursSpent, s.CampInventorySummary()),
	}
}

func (s *RunState) executePreserveCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: preserve <smoke|dry|salt> <raw_small_game_meat|raw_bird_meat|raw_fish_meat|cooked_...> [kg] [p#]"}
	}
	method := fields[0]
	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
	if len(rest) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: preserve <smoke|dry|salt> <raw_small_game_meat|raw_bird_meat|raw_fish_meat|cooked_...> [kg] [p#]"}
	}
	if !hasAmount {
		amount = 0
	}
	result, err := s.PreserveFood(playerID, method, rest[0], amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Preserve failed: %v", err)}
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("P%d preserved %.2fkg %s via %s -> %.2fkg %s (%.1fh, shelf-life ~%d days). %s",
			playerID, result.InputKg, result.SourceID, result.Method, result.OutputKg, result.PreservedID, result.HoursSpent, result.ShelfLifeDays, s.CampInventorySummary()),
	}
}

func (s *RunState) executeEatCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: eat <food_item> [grams|kg] [p#]"}
	}
	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields)
	if len(rest) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: eat <food_item> [grams|kg] [p#]"}
	}
	if !hasAmount {
		amount = 0
	}
	result, err := s.EatFood(playerID, rest[0], amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Eat failed: %v", err)}
	}
	msg := fmt.Sprintf("P%d ate %dg %s: %dkcal %dgP %dgF %dgS | %+dE %+dH2O %+dM",
		playerID,
		result.ConsumedGrams,
		result.ItemID,
		result.Nutrition.CaloriesKcal, result.Nutrition.ProteinG, result.Nutrition.FatG, result.Nutrition.SugarG,
		result.EnergyDelta, result.HydrationDelta, result.MoraleDelta,
	)
	if result.GotIll {
		msg += " | illness triggered (food poisoning)"
	}
	return RunCommandResult{Handled: true, Message: msg}
}

func (s *RunState) executeGoCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: go <north|south|east|west|n|s|e|w> <distance> [p#]"}
	}

	playerID := 1
	direction := ""
	distanceTokens := make([]string, 0, 2)
	for _, field := range fields {
		token := strings.ToLower(strings.TrimSpace(field))
		if token == "" {
			continue
		}
		if parsed := parsePlayerToken(token); parsed > 0 {
			playerID = parsed
			continue
		}
		if direction == "" {
			if mapped := normalizeDirection(token); mapped != "" {
				direction = mapped
				continue
			}
		}
		distanceTokens = append(distanceTokens, token)
	}

	if direction == "" {
		return RunCommandResult{Handled: true, Message: "Direction must be north/south/east/west."}
	}
	if len(distanceTokens) == 0 {
		return RunCommandResult{Handled: true, Message: "How far? (e.g. 'go north 500m' or 'go north 3km' or 'go north 5 tiles')"}
	}
	amount, ok := parseTravelDistanceTokens(distanceTokens)
	if !ok || amount <= 0 {
		return RunCommandResult{Handled: true, Message: "Distance must be > 0. Use meters (m), kilometers (km), or tiles."}
	}

	result, err := s.TravelMove(playerID, direction, amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Travel failed: %v", err)}
	}
	craftText := "on foot"
	if result.WatercraftUsed != "" {
		craftText = "using " + result.WatercraftUsed
	}
	encounterText := ""
	if len(result.EncounterLogs) > 0 {
		encounterText = " | Encounters: " + strings.Join(result.EncounterLogs, " ")
	}
	stopText := ""
	if strings.TrimSpace(result.StopReason) != "" {
		stopText = " Stopped: " + result.StopReason + "."
	}
	blockText := fmt.Sprintf(" Time block: %s -> %s", result.StartBlock, result.EndBlock)
	if result.BlocksCrossed > 0 {
		blockText += fmt.Sprintf(" (crossed %d)", result.BlocksCrossed)
	}
	return RunCommandResult{
		Handled:       true,
		HoursAdvanced: result.HoursSpent,
		Message: fmt.Sprintf("Travelling %s %.1fkm... (%d steps). P%d traveled %.1fkm %s at %.1fkm/h (%.1fh). Cost: -%dE -%dH2O %+dM. Total travel %.1fkm.",
			result.Direction, result.RequestedKm, result.RequestedSteps,
			playerID, result.DistanceKm, craftText, result.TravelSpeedKmph, result.HoursSpent, result.EnergyCost, result.HydrationCost, result.MoraleDelta, s.Travel.TotalKm) +
			fmt.Sprintf(" Position: (%d,%d).", s.Travel.PosX, s.Travel.PosY) + blockText + "." + stopText + encounterText + " " + s.describeDirectionalView(playerID, "front", false, ""),
	}
}

func (s *RunState) executeCollectCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: collect <resource|any> [qty] [p#]"}
	}

	playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields)
	resourceID := "any"
	if len(rest) > 0 {
		resourceID = rest[0]
	}
	if !hasAmount {
		amount = 0
	}

	resource, qty, err := s.CollectResource(playerID, resourceID, amount)
	if err != nil {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Collect failed: %v", err)}
	}
	unit := resource.Unit
	if unit == "kg" {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d collected %.1f %s of %s. %s", playerID, qty, unit, resource.Name, s.CampInventorySummary())}
	}
	return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d collected %.0f %s of %s. %s", playerID, qty, unit, resource.Name, s.CampInventorySummary())}
}

func (s *RunState) executeWoodCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: wood gather [kg] [p#] | wood dry [kg] [p#] | wood stock"}
	}

	switch fields[0] {
	case "stock", "status":
		return RunCommandResult{Handled: true, Message: s.CampInventorySummary()}
	case "gather":
		playerID, amount, hasAmount, _ := parseOptionalPlayerAndNumber(fields[1:])
		if !hasAmount {
			amount = 0
		}
		tree, kg, err := s.GatherWood(playerID, amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Wood gather failed: %v", err)}
		}
		return RunCommandResult{
			Handled: true,
			Message: fmt.Sprintf("P%d gathered %.1fkg from %s (%s). %s",
				playerID, kg, tree.Name, tree.WoodType, s.CampInventorySummary()),
		}
	case "dry":
		playerID, amount, hasAmount, _ := parseOptionalPlayerAndNumber(fields[1:])
		if !hasAmount {
			amount = 1.0
		}
		dried, err := s.DryWood(playerID, amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Wood dry failed: %v", err)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d dried %.1fkg wood. %s", playerID, dried, s.CampInventorySummary())}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: wood gather [kg] [p#] | wood dry [kg] [p#] | wood stock"}
	}
}

func (s *RunState) executeFireCommand(fields []string) RunCommandResult {
	if len(fields) == 0 || fields[0] == "status" {
		if !s.Fire.Lit {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fire: out | Prep: %s | Wood stock: %s", formatFirePrep(s.FirePrep), formatWoodStock(s.WoodStock))}
		}
		return RunCommandResult{
			Handled: true,
			Message: fmt.Sprintf("Fire: lit (%s) intensity %d heat %dC fuel %.1fkg method:%s | Prep: %s | Wood stock: %s",
				s.Fire.WoodType, s.Fire.Intensity, s.Fire.HeatC, s.Fire.FuelKg, s.Fire.LastMethod, formatFirePrep(s.FirePrep), formatWoodStock(s.WoodStock)),
		}
	}

	switch fields[0] {
	case "methods":
		return RunCommandResult{Handled: true, Message: "Fire methods -> ferro, bow_drill, hand_drill"}
	case "prep":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: fire prep tinder|kindling|feather [count] [p#]"}
		}
		playerID, amount, hasAmount, _ := parseOptionalPlayerAndNumber(fields[2:])
		count := 1
		if hasAmount {
			count = clamp(int(math.Round(amount)), 1, 10)
		}
		created, err := s.PrepareFireMaterial(playerID, fields[1], count)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fire prep failed: %v", err)}
		}
		return RunCommandResult{
			Handled: true,
			Message: fmt.Sprintf("P%d prepared %d %s bundle(s). Prep: %s", playerID, created, fields[1], formatFirePrep(s.FirePrep)),
		}
	case "ember":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: fire ember bow|hand [woodtype] [p#]"}
		}
		method := ParseFireMethod(fields[1])
		if method != FireMethodBowDrill && method != FireMethodHandDrill {
			return RunCommandResult{Handled: true, Message: "Ember method must be bow or hand."}
		}
		playerID := 1
		woodType := WoodType("")
		for _, token := range fields[2:] {
			if parsed := parsePlayerToken(token); parsed > 0 {
				playerID = parsed
				continue
			}
			if parsed := parseWoodType(token); parsed != "" {
				woodType = parsed
			}
		}
		chance, success, err := s.TryCreateEmber(playerID, method, woodType)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fire ember failed: %v", err)}
		}
		if success {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d created ember with %s (chance %.0f%%). Prep: %s", playerID, method, chance*100, formatFirePrep(s.FirePrep))}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d failed ember with %s (chance %.0f%%). Prep: %s", playerID, method, chance*100, formatFirePrep(s.FirePrep))}
	case "ignite":
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if !hasAmount {
			amount = 1.0
		}
		woodType := WoodType("")
		for _, token := range rest {
			if parsed := parseWoodType(token); parsed != "" {
				woodType = parsed
				break
			}
		}
		chance, success, err := s.IgniteFromEmber(playerID, woodType, amount)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Ignite failed: %v", err)}
		}
		if success {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d ignited fire from ember (chance %.0f%%). Fire heat %dC intensity %d.", playerID, chance*100, s.Fire.HeatC, s.Fire.Intensity)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d failed to ignite from ember (chance %.0f%%). Prep: %s", playerID, chance*100, formatFirePrep(s.FirePrep))}
	case "out", "extinguish":
		s.ExtinguishFire()
		return RunCommandResult{Handled: true, Message: "Fire extinguished."}
	case "build", "start":
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if !hasAmount {
			amount = 1.0
		}
		woodType := s.Fire.WoodType
		if woodType == "" {
			if len(s.WoodStock) > 0 {
				woodType = s.WoodStock[0].Type
			} else {
				woodType = WoodTypeHardwood
			}
		}
		for _, token := range rest {
			if parsed := parseWoodType(token); parsed != "" {
				woodType = parsed
				break
			}
		}
		if err := s.startFireWithMethod(playerID, woodType, amount, FireMethodFerro); err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fire start failed: %v", err)}
		}
		return RunCommandResult{
			Handled: true,
			Message: fmt.Sprintf("P%d started fire with %.1fkg %s using ferro method. Intensity %d, heat %dC.",
				playerID, amount, woodType, s.Fire.Intensity, s.Fire.HeatC),
		}
	case "tend":
		playerID, amount, hasAmount, rest := parseOptionalPlayerAndNumber(fields[1:])
		if !hasAmount {
			amount = 0.8
		}
		woodType := s.Fire.WoodType
		for _, token := range rest {
			if parsed := parseWoodType(token); parsed != "" {
				woodType = parsed
				break
			}
		}
		if err := s.TendFire(playerID, amount, woodType); err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Fire tend failed: %v", err)}
		}
		return RunCommandResult{
			Handled: true,
			Message: fmt.Sprintf("P%d tended fire with %.1fkg %s. Intensity %d, heat %dC, fuel %.1fkg.",
				playerID, amount, woodType, s.Fire.Intensity, s.Fire.HeatC, s.Fire.FuelKg),
		}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: fire status | fire methods | fire prep tinder|kindling|feather [count] [p#] | fire ember bow|hand [woodtype] [p#] | fire ignite [woodtype] [kg] [p#] | fire build [woodtype] [kg] [p#] | fire tend [woodtype] [kg] [p#] | fire out"}
	}
}

func (s *RunState) executeShelterCommand(fields []string) RunCommandResult {
	if len(fields) == 0 || fields[0] == "status" {
		if s.Shelter.Type == "" || s.Shelter.Durability <= 0 {
			return RunCommandResult{Handled: true, Message: "Shelter: none"}
		}
		spec, ok := shelterByID(s.Shelter.Type)
		if !ok {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Shelter: %s stage %d durability %d%%", s.Shelter.Type, s.Shelter.Stage, s.Shelter.Durability)}
		}
		stageCount := len(spec.Stages)
		if stageCount == 0 {
			stageCount = 1
		}
		if metrics, ok := s.currentShelterMetrics(); ok {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Shelter: %s stage %d/%d durability %d%% | ins:%d rain:%d wind:%d insect:%d storage:%.0fkg", spec.Name, max(1, s.Shelter.Stage), stageCount, s.Shelter.Durability, metrics.Insulation, metrics.RainProtection, metrics.WindProtection, metrics.InsectProtection, metrics.StorageCapacityKg)}
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Shelter: %s stage %d/%d durability %d%%", spec.Name, max(1, s.Shelter.Stage), stageCount, s.Shelter.Durability)}
	}

	switch fields[0] {
	case "list":
		options := SheltersForBiome(s.Scenario.Biome)
		parts := make([]string, 0, len(options))
		for _, option := range options {
			parts = append(parts, fmt.Sprintf("%s(%s)", option.Name, option.ID))
		}
		return RunCommandResult{Handled: true, Message: "Shelters -> " + strings.Join(parts, ", ")}
	case "build":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: shelter build <id> [p#]"}
		}
		playerID := 1
		shelterID := fields[1]
		if len(fields) > 2 {
			for _, token := range fields[2:] {
				if parsed := parsePlayerToken(token); parsed > 0 {
					playerID = parsed
				}
			}
		}
		shelter, err := s.BuildShelter(playerID, shelterID)
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Shelter build failed: %v", err)}
		}
		stageCount := len(shelter.Stages)
		if stageCount == 0 {
			stageCount = 1
		}
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("P%d built %s stage %d/%d. Durability %d%%.", playerID, shelter.Name, max(1, s.Shelter.Stage), stageCount, s.Shelter.Durability)}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: shelter list | shelter build <id> [p#] | shelter status"}
	}
}

func (s *RunState) executeCraftCommand(fields []string) RunCommandResult {
	if len(fields) == 0 {
		return RunCommandResult{Handled: true, Message: "Usage: craft list | craft make <id> [p#] | craft inventory"}
	}
	switch fields[0] {
	case "inventory":
		if len(s.CraftedItems) == 0 {
			return RunCommandResult{Handled: true, Message: "Crafted: none"}
		}
		return RunCommandResult{Handled: true, Message: "Crafted: " + strings.Join(s.CraftedItems, ", ")}
	case "list":
		options := CraftablesForBiome(s.Scenario.Biome)
		parts := make([]string, 0, len(options))
		for _, option := range options {
			parts = append(parts, fmt.Sprintf("%s(%s)", option.Name, option.ID))
		}
		return RunCommandResult{Handled: true, Message: "Craftables -> " + strings.Join(parts, ", ")}
	case "make":
		if len(fields) < 2 {
			return RunCommandResult{Handled: true, Message: "Usage: craft make <id> [p#]"}
		}
		playerID := 1
		for _, token := range fields[2:] {
			if parsed := parsePlayerToken(token); parsed > 0 {
				playerID = parsed
			}
		}
		outcome, err := s.CraftItem(playerID, fields[1])
		if err != nil {
			return RunCommandResult{Handled: true, Message: fmt.Sprintf("Craft failed: %v", err)}
		}
		return RunCommandResult{
			Handled:       true,
			HoursAdvanced: outcome.HoursSpent,
			Message: fmt.Sprintf("P%d crafted %s (%s, %.1fh, stored %s).",
				playerID, outcome.Spec.Name, outcome.Quality, outcome.HoursSpent, outcome.StoredAt),
		}
	default:
		return RunCommandResult{Handled: true, Message: "Usage: craft list | craft make <id> [p#] | craft inventory"}
	}
}

func parseItemAndAction(tokens []string) (KitItem, string, bool) {
	for i := len(tokens) - 1; i >= 1; i-- {
		itemCandidate := strings.Join(tokens[:i], " ")
		item, ok := parseKitAlias(itemCandidate)
		if !ok {
			continue
		}
		actionInput := strings.Join(tokens[i:], " ")
		if strings.TrimSpace(actionInput) == "" {
			return "", "", false
		}
		return item, actionInput, true
	}
	return "", "", false
}

func extractPlayerID(tokens []string) (int, []string) {
	playerID := 1
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if parsed := parsePlayerToken(token); parsed > 0 {
			playerID = parsed
			continue
		}
		filtered = append(filtered, token)
	}
	return playerID, filtered
}

func parsePlayerToken(raw string) int {
	value := strings.TrimSpace(strings.ToLower(raw))
	if !strings.HasPrefix(value, "p") || len(value) < 2 {
		return 0
	}
	parsed, err := strconv.Atoi(strings.TrimPrefix(value, "p"))
	if err != nil || parsed < 1 {
		return 0
	}
	return parsed
}

func findAction(item KitItem, actionInput string) (equipmentAction, bool) {
	actions := actionsForItem(item)
	if len(actions) == 0 {
		return equipmentAction{}, false
	}

	input := normalizeCommandToken(actionInput)
	best := equipmentAction{}
	bestScore := -1
	for _, action := range actions {
		candidates := append([]string{action.ID}, action.Aliases...)
		for _, candidate := range candidates {
			normCandidate := normalizeCommandToken(candidate)
			if normCandidate == "" {
				continue
			}
			if input == normCandidate {
				if len(normCandidate) > bestScore {
					best = action
					bestScore = len(normCandidate)
				}
				continue
			}
			if strings.Contains(input, normCandidate) || strings.Contains(normCandidate, input) {
				if len(normCandidate) > bestScore {
					best = action
					bestScore = len(normCandidate)
				}
			}
		}
	}
	if bestScore < 0 {
		return equipmentAction{}, false
	}
	return best, true
}

func (s *RunState) playerByID(playerID int) (*PlayerState, bool) {
	for i := range s.Players {
		if s.Players[i].ID == playerID {
			return &s.Players[i], true
		}
	}
	return nil, false
}

func playerHasKitItem(player *PlayerState, issued []KitItem, item KitItem) bool {
	if player == nil {
		return false
	}
	for _, own := range player.Kit {
		if own == item {
			return true
		}
	}
	for _, own := range issued {
		if own == item {
			return true
		}
	}
	return false
}

func uniqueKitItems(personal []KitItem, issued []KitItem) []KitItem {
	seen := map[KitItem]bool{}
	out := make([]KitItem, 0, len(personal)+len(issued))
	for _, item := range personal {
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	for _, item := range issued {
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func parseKitAlias(raw string) (KitItem, bool) {
	norm := normalizeCommandToken(raw)
	item, ok := kitAliasMap()[norm]
	return item, ok
}

func normalizeCommandToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func kitAliasMap() map[string]KitItem {
	m := map[string]KitItem{}
	add := func(item KitItem, aliases ...string) {
		for _, alias := range aliases {
			norm := normalizeCommandToken(alias)
			if norm != "" {
				m[norm] = item
			}
		}
	}

	for _, item := range AllKitItems() {
		add(item, string(item))
	}

	add(KitHatchet, "hatchet", "axe")
	add(KitSixInchKnife, "knife", "six inch knife", "6 inch knife")
	add(KitMachete, "machete")
	add(KitFoldingSaw, "saw", "folding saw")
	add(KitParacord50ft, "paracord", "cord")
	add(KitFerroRod, "ferro", "ferro rod", "firesteel")
	add(KitFirePlunger, "fire plunger", "plunger")
	add(KitMagnifyingLens, "magnifying lens", "lens")
	add(KitCookingPot, "pot", "cooking pot")
	add(KitMetalCup, "cup", "metal cup")
	add(KitCanteen, "canteen")
	add(KitWaterFilter, "filter", "water filter")
	add(KitPurificationTablets, "tabs", "tablets", "purification tablets")
	add(KitFishingLineHooks, "fishing line", "hooks", "line hooks")
	add(KitGillNet, "gill net", "net")
	add(KitSpear, "spear", "fishing spear")
	add(KitSnareWire, "snare wire", "snare")
	add(KitBowArrows, "bow", "arrows", "bow arrows")
	add(KitTarp, "tarp")
	add(KitSleepingBag, "sleeping bag", "bag")
	add(KitWoolBlanket, "wool blanket", "blanket")
	add(KitThermalLayer, "thermal layer", "thermal")
	add(KitRainJacket, "rain jacket", "jacket")
	add(KitMosquitoNet, "mosquito net", "bug net")
	add(KitInsectRepellent, "repellent", "insect repellent")
	add(KitCompass, "compass")
	add(KitMap, "map")
	add(KitHeadlamp, "headlamp", "lamp")
	add(KitSignalMirror, "signal mirror", "mirror")
	add(KitWhistle, "whistle")
	add(KitMultiTool, "multitool", "multi tool", "tool")
	add(KitDuctTape, "duct tape", "tape")
	add(KitSewingKit, "sewing kit", "sewing")
	add(KitShovel, "shovel")
	add(KitClimbingRope, "climbing rope", "rope")
	add(KitCarabiners, "carabiners", "carabiner")
	add(KitFirstAidKit, "first aid", "first aid kit", "med kit")
	add(KitSalt, "salt")
	add(KitEmergencyRations, "rations", "emergency rations")
	add(KitDryBag, "dry bag", "drybag")

	return m
}

func itemCommandLabel(item KitItem) string {
	switch item {
	case KitSixInchKnife:
		return "knife"
	case KitParacord50ft:
		return "paracord"
	case KitFerroRod:
		return "ferro"
	case KitFishingLineHooks:
		return "fishingline"
	case KitPurificationTablets:
		return "tablets"
	case KitBowArrows:
		return "bow"
	case KitWoolBlanket:
		return "blanket"
	case KitThermalLayer:
		return "thermal"
	case KitRainJacket:
		return "jacket"
	case KitInsectRepellent:
		return "repellent"
	case KitSignalMirror:
		return "mirror"
	case KitMultiTool:
		return "multitool"
	case KitFirstAidKit:
		return "firstaid"
	case KitEmergencyRations:
		return "rations"
	default:
		return normalizeCommandToken(string(item))
	}
}

func actionsForItem(item KitItem) []equipmentAction {
	if actions, ok := equipmentActionLibrary[item]; ok {
		return actions
	}

	return []equipmentAction{
		{
			ID:          "improvise",
			Aliases:     []string{"improvise use"},
			Description: "Use gear creatively to stabilize camp workflow.",
			EnergyDelta: -1,
			Hydration:   0,
			MoraleDelta: 1,
		},
	}
}

var equipmentActionLibrary = map[KitItem][]equipmentAction{
	KitHatchet: {
		{ID: "split_kindling", Aliases: []string{"split wood", "kindling"}, Description: "Split dry wood for steady fire fuel.", EnergyDelta: -1, Hydration: -1, MoraleDelta: 2},
		{ID: "shape_poles", Aliases: []string{"shape poles", "notch poles"}, Description: "Shape poles for shelter framing.", EnergyDelta: -2, Hydration: -1, MoraleDelta: 2},
	},
	KitSixInchKnife: {
		{ID: "carve_stakes", Aliases: []string{"carve stakes", "stakes"}, Description: "Carve stakes and trap triggers.", EnergyDelta: -1, MoraleDelta: 1},
		{ID: "prepare_game", Aliases: []string{"butcher", "prepare meat"}, Description: "Process game into safer portions.", EnergyDelta: -1, MoraleDelta: 1},
	},
	KitMachete: {
		{ID: "clear_brush", Aliases: []string{"cut brush", "clear path"}, Description: "Clear overgrowth for travel and shelter.", EnergyDelta: -2, Hydration: -1, MoraleDelta: 2},
	},
	KitFoldingSaw: {
		{ID: "cut_poles", Aliases: []string{"saw poles", "cut branches"}, Description: "Cut straight poles for structure work.", EnergyDelta: -1, Hydration: -1, MoraleDelta: 2},
	},
	KitParacord50ft: {
		{ID: "tie_sticks_together", Aliases: []string{"tie sticks", "tie sticks together", "lash poles"}, Description: "Lash sticks into stronger shelter frames.", EnergyDelta: -1, MoraleDelta: 3},
		{ID: "rig_tripline", Aliases: []string{"trip line", "rig line"}, Description: "Set warning or trap line around camp.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitFerroRod: {
		{ID: "spark_tinder", Aliases: []string{"start fire", "ferro spark"}, Description: "Strike sparks to ignite dry tinder.", EnergyDelta: -1, MoraleDelta: 3},
	},
	KitFirePlunger: {
		{ID: "compress_ember", Aliases: []string{"plunger ember", "ember"}, Description: "Create ember by compression ignition.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitMagnifyingLens: {
		{ID: "solar_ignite", Aliases: []string{"solar fire", "sun ignite"}, Description: "Use sunlight to light charred tinder.", EnergyDelta: 0, MoraleDelta: 2},
	},
	KitCookingPot: {
		{ID: "boil_stew", Aliases: []string{"boil water", "stew"}, Description: "Boil water and cook food safely.", EnergyDelta: 1, Hydration: 2, MoraleDelta: 2},
	},
	KitMetalCup: {
		{ID: "boil_small_batch", Aliases: []string{"boil cup", "small boil"}, Description: "Quickly boil a small water batch.", EnergyDelta: 0, Hydration: 1, MoraleDelta: 1},
	},
	KitCanteen: {
		{ID: "carry_water", Aliases: []string{"fill canteen", "water carry"}, Description: "Carry reserve water for travel.", EnergyDelta: 0, Hydration: 2, MoraleDelta: 1},
	},
	KitWaterFilter: {
		{ID: "filter_stream", Aliases: []string{"filter water", "purify water"}, Description: "Filter stream water to lower contamination risk.", EnergyDelta: 0, Hydration: 3, MoraleDelta: 1},
	},
	KitPurificationTablets: {
		{ID: "purify_batch", Aliases: []string{"use tablets", "purify"}, Description: "Chemically treat collected water.", EnergyDelta: 0, Hydration: 3, MoraleDelta: 1},
	},
	KitFishingLineHooks: {
		{ID: "set_hookline", Aliases: []string{"hook line", "set line"}, Description: "Set passive hook line near feeding routes.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitGillNet: {
		{ID: "deploy_gill_net", Aliases: []string{"deploy net", "set net"}, Description: "Deploy net for passive fish capture.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitSpear: {
		{ID: "spear_fish", Aliases: []string{"fish spear", "spear hunt"}, Description: "Hunt fish in shallows by thrusting.", EnergyDelta: -2, Hydration: -1, MoraleDelta: 2},
	},
	KitSnareWire: {
		{ID: "set_snare", Aliases: []string{"snare", "wire snare"}, Description: "Set wire snare on active game trails.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitBowArrows: {
		{ID: "stalk_shot", Aliases: []string{"hunt with bow", "bow shot"}, Description: "Take a controlled bow shot on game.", EnergyDelta: -2, Hydration: -1, MoraleDelta: 3},
	},
	KitTarp: {
		{ID: "pitch_tarp", Aliases: []string{"setup tarp", "tarp shelter"}, Description: "Pitch tarp to reduce exposure and rain soak.", EnergyDelta: -1, MoraleDelta: 3},
	},
	KitSleepingBag: {
		{ID: "insulated_sleep", Aliases: []string{"rest warm", "sleep"}, Description: "Recover with insulated rest.", EnergyDelta: 2, MoraleDelta: 2},
	},
	KitWoolBlanket: {
		{ID: "wrap_warmth", Aliases: []string{"stay warm", "wrap"}, Description: "Retain heat during cold rest periods.", EnergyDelta: 1, MoraleDelta: 2},
	},
	KitThermalLayer: {
		{ID: "layer_up", Aliases: []string{"dress warm", "thermal up"}, Description: "Reduce cold-weather stress while active.", EnergyDelta: 1, MoraleDelta: 1},
	},
	KitRainJacket: {
		{ID: "weatherproof", Aliases: []string{"rainproof", "wear jacket"}, Description: "Cut rain exposure during tasks.", EnergyDelta: 1, MoraleDelta: 1},
	},
	KitMosquitoNet: {
		{ID: "bug_barrier", Aliases: []string{"net sleep", "mosquito barrier"}, Description: "Block insects while sleeping.", EnergyDelta: 1, MoraleDelta: 2},
	},
	KitInsectRepellent: {
		{ID: "apply_repellent", Aliases: []string{"repel bugs", "apply bug spray"}, Description: "Reduce insect harassment for several hours.", EnergyDelta: 0, MoraleDelta: 2},
	},
	KitCompass: {
		{ID: "orient_course", Aliases: []string{"navigate", "set bearing"}, Description: "Set reliable travel bearing.", EnergyDelta: 0, MoraleDelta: 1},
	},
	KitMap: {
		{ID: "plot_route", Aliases: []string{"plan route", "route"}, Description: "Plan route to avoid unnecessary detours.", EnergyDelta: 0, MoraleDelta: 1},
	},
	KitHeadlamp: {
		{ID: "night_task", Aliases: []string{"work at night", "night"}, Description: "Complete controlled tasks after dark.", EnergyDelta: -1, MoraleDelta: 1},
	},
	KitSignalMirror: {
		{ID: "signal_pass", Aliases: []string{"signal", "flash mirror"}, Description: "Signal aircraft or boats in clear weather.", MoraleDelta: 2},
	},
	KitWhistle: {
		{ID: "emergency_signal", Aliases: []string{"blow whistle", "whistle signal"}, Description: "Issue loud emergency signal blasts.", MoraleDelta: 1},
	},
	KitMultiTool: {
		{ID: "repair_gear", Aliases: []string{"repair", "fix tool"}, Description: "Repair worn gear and fittings.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitDuctTape: {
		{ID: "patch_shelter", Aliases: []string{"patch tarp", "patch gear"}, Description: "Seal leaks and reinforce stress points.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitSewingKit: {
		{ID: "mend_clothes", Aliases: []string{"stitch clothes", "mend"}, Description: "Mend tears to retain warmth and comfort.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitShovel: {
		{ID: "dig_drainage", Aliases: []string{"dig trench", "drainage"}, Description: "Dig drainage and improve camp footing.", EnergyDelta: -2, Hydration: -1, MoraleDelta: 2},
	},
	KitClimbingRope: {
		{ID: "rig_line", Aliases: []string{"rope line", "belay"}, Description: "Rig safe line for steep access.", EnergyDelta: -1, MoraleDelta: 2},
	},
	KitCarabiners: {
		{ID: "anchor_line", Aliases: []string{"clip anchor", "anchor"}, Description: "Secure rope anchor points quickly.", EnergyDelta: 0, MoraleDelta: 1},
	},
	KitFirstAidKit: {
		{ID: "treat_wound", Aliases: []string{"first aid", "treat injury", "medicate"}, Description: "Treat injuries and reduce symptom burden.", EnergyDelta: 1, MoraleDelta: 2, Special: specialTreatAilment},
	},
	KitSalt: {
		{ID: "preserve_meat", Aliases: []string{"salt meat", "cure food"}, Description: "Preserve meat to stretch food stores.", EnergyDelta: 0, MoraleDelta: 1},
	},
	KitEmergencyRations: {
		{ID: "eat_ration", Aliases: []string{"eat", "ration"}, Description: "Eat ration pack for rapid calories.", Nutrition: NutritionTotals{CaloriesKcal: 650, ProteinG: 24, FatG: 26, SugarG: 28}, MoraleDelta: 2},
	},
	KitDryBag: {
		{ID: "waterproof_cache", Aliases: []string{"protect gear", "dry stash"}, Description: "Keep critical gear dry during storms.", EnergyDelta: 0, MoraleDelta: 2},
	},
}
