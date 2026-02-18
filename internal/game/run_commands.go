package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type RunCommandResult struct {
	Handled bool
	Message string
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
			Message: "Commands: hunt land|fish|air [raw] [liver] [p#] [grams], actions [p#], use <item> <action> [p#], next, save, load, menu.",
		}
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
	default:
		return RunCommandResult{Handled: false}
	}
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

	if action.Nutrition.CaloriesKcal > 0 || action.Nutrition.ProteinG > 0 || action.Nutrition.FatG > 0 {
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
	if action.Nutrition.CaloriesKcal > 0 || action.Nutrition.ProteinG > 0 || action.Nutrition.FatG > 0 {
		msg += fmt.Sprintf(" | +%dkcal +%dgP +%dgF", action.Nutrition.CaloriesKcal, action.Nutrition.ProteinG, action.Nutrition.FatG)
	}
	msg += specialMsg
	return RunCommandResult{Handled: true, Message: msg}
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
		{ID: "eat_ration", Aliases: []string{"eat", "ration"}, Description: "Eat ration pack for rapid calories.", Nutrition: NutritionTotals{CaloriesKcal: 650, ProteinG: 24, FatG: 26}, MoraleDelta: 2},
	},
	KitDryBag: {
		{ID: "waterproof_cache", Aliases: []string{"protect gear", "dry stash"}, Description: "Keep critical gear dry during storms.", EnergyDelta: 0, MoraleDelta: 2},
	},
}
