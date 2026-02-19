package game

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// Discovery summary:
// - Trap gameplay is table-driven: TrapCatalog defines setup cost/chance/yield and daily resolution is deterministic.
// - Only meaningful state is persisted (condition, armed state, pending catch kg/type), which keeps simulation cheap.
// - Catch output maps through trapCatchItemID/name, so expanding target channels is safe with small ID mapping hooks.
type TrapSpec struct {
	ID                string
	Name              string
	Description       string
	BiomeTags         []string
	Targets           []string
	MinBushcraft      int
	BaseChance        float64
	BaseHours         float64
	ConditionLoss     int
	YieldMinKg        float64
	YieldMaxKg        float64
	NeedsWater        bool
	RequiresCrafted   []string
	RequiresResources []ResourceRequirement
	RequiresKit       []KitItem
}

type PlacedTrap struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	SetByPlayerID    int          `json:"set_by_player_id"`
	Quality          CraftQuality `json:"quality"`
	Effectiveness    float64      `json:"effectiveness"`
	Condition        int          `json:"condition"`
	Targets          []string     `json:"targets,omitempty"`
	Armed            bool         `json:"armed"`
	PendingCatchKg   float64      `json:"pending_catch_kg"`
	PendingCatchType string       `json:"pending_catch_type,omitempty"`
	LastResolvedDay  int          `json:"last_resolved_day"`
	LastCheckedDay   int          `json:"last_checked_day"`
	Successes        int          `json:"successes"`
	Failures         int          `json:"failures"`
	Broken           int          `json:"broken"`
}

type TrapSetResult struct {
	Trap    TrapSpec
	Quality CraftQuality
	Hours   float64
	Chance  float64
}

type TrapCheckResult struct {
	Checked      int
	CollectedKg  float64
	Rearmed      int
	Broken       int
	CampOverflow int
}

func TrapCatalog() []TrapSpec {
	base := []TrapSpec{
		{
			ID: "gorge_hook_line", Name: "Gorge Hook Line",
			Description:  "Carved gorge hook and fiber line for passive fish takes near still edges.",
			BiomeTags:    []string{"river", "lake", "delta", "coast", "wetlands"},
			Targets:      []string{"fish"},
			MinBushcraft: 1, BaseChance: 0.27, BaseHours: 0.6, ConditionLoss: 5, YieldMinKg: 0.12, YieldMaxKg: 0.7, NeedsWater: true,
			RequiresCrafted: []string{"fish_gorge_hooks", "natural_twine"},
		},
		{
			ID: "spring_snare", Name: "Spring Snare",
			Description:  "Bent sapling spring snare for rabbit-sized game trails.",
			BiomeTags:    []string{"forest", "boreal", "savanna", "mountain", "badlands"},
			Targets:      []string{"small_game"},
			MinBushcraft: 2, BaseChance: 0.23, BaseHours: 0.8, ConditionLoss: 8, YieldMinKg: 0.2, YieldMaxKg: 1.1,
			RequiresCrafted: []string{"trap_trigger_set", "natural_twine"},
		},
		{
			ID: "peg_snare", Name: "Peg Snare",
			Description:  "Simple trail snare built with peg anchors and natural cordage.",
			BiomeTags:    []string{"forest", "boreal", "savanna", "badlands", "tundra"},
			Targets:      []string{"small_game"},
			MinBushcraft: 1, BaseChance: 0.19, BaseHours: 0.5, ConditionLoss: 7, YieldMinKg: 0.18, YieldMaxKg: 0.9,
			RequiresCrafted: []string{"natural_twine"},
		},
		{
			ID: "figure4_deadfall", Name: "Figure-4 Deadfall",
			Description:  "Deadfall trigger tuned for rodents and birds.",
			BiomeTags:    []string{"forest", "boreal", "mountain", "savanna", "desert", "badlands"},
			Targets:      []string{"small_game", "bird"},
			MinBushcraft: 2, BaseChance: 0.21, BaseHours: 1.1, ConditionLoss: 9, YieldMinKg: 0.14, YieldMaxKg: 0.85,
			RequiresCrafted: []string{"trap_trigger_set"},
		},
		{
			ID: "paiute_deadfall", Name: "Paiute Deadfall",
			Description:  "Cordage-assisted deadfall with cleaner trigger geometry.",
			BiomeTags:    []string{"forest", "boreal", "mountain", "badlands", "desert"},
			Targets:      []string{"small_game"},
			MinBushcraft: 3, BaseChance: 0.26, BaseHours: 1.3, ConditionLoss: 8, YieldMinKg: 0.18, YieldMaxKg: 1.0,
			RequiresCrafted: []string{"trap_trigger_set", "natural_twine"},
		},
		{
			ID: "funnel_fish_basket", Name: "Funnel Fish Basket",
			Description:  "Baited funnel basket trap for creek narrows and tidal channels.",
			BiomeTags:    []string{"delta", "river", "lake", "swamp", "coast", "wetlands"},
			Targets:      []string{"fish"},
			MinBushcraft: 1, BaseChance: 0.24, BaseHours: 1.2, ConditionLoss: 4, YieldMinKg: 0.2, YieldMaxKg: 1.4, NeedsWater: true,
			RequiresCrafted: []string{"fish_trap", "natural_twine"},
		},
		{
			ID: "bird_noose_perch", Name: "Bird Noose Perch",
			Description:  "Perch noose trap for ground and low-branch birds.",
			BiomeTags:    []string{"forest", "coast", "savanna", "wetlands", "badlands"},
			Targets:      []string{"bird"},
			MinBushcraft: 1, BaseChance: 0.2, BaseHours: 0.7, ConditionLoss: 6, YieldMinKg: 0.1, YieldMaxKg: 0.5,
			RequiresCrafted: []string{"natural_twine"},
		},
		{
			ID: "paracord_twitchup", Name: "Paracord Twitch-Up",
			Description:  "Kit-assisted spring snare using paracord and snare wire.",
			BiomeTags:    []string{"forest", "boreal", "mountain", "savanna", "badlands"},
			Targets:      []string{"small_game"},
			MinBushcraft: 1, BaseChance: 0.31, BaseHours: 0.65, ConditionLoss: 7, YieldMinKg: 0.2, YieldMaxKg: 1.2,
			RequiresCrafted: []string{"trap_trigger_set"},
			RequiresKit:     []KitItem{KitParacord50ft, KitSnareWire},
		},
	}
	return append(base, expandedTrapCatalog()...)
}

func expandedTrapCatalog() []TrapSpec {
	return []TrapSpec{
		{
			ID: "trail_twitchup", Name: "Trail Twitch-Up",
			Description:  "Twitch-up variant tuned for hare and squirrel trails.",
			BiomeTags:    []string{"forest", "boreal", "mountain", "savanna"},
			Targets:      []string{"small_game"},
			MinBushcraft: 2, BaseChance: 0.25, BaseHours: 0.9, ConditionLoss: 8, YieldMinKg: 0.2, YieldMaxKg: 1.0,
			RequiresCrafted: []string{"spring_snare_kit"},
		},
		{
			ID: "rolling_log_deadfall", Name: "Rolling Log Deadfall",
			Description:  "Weighted rolling deadfall for medium trail game.",
			BiomeTags:    []string{"forest", "mountain", "boreal", "badlands"},
			Targets:      []string{"small_game", "medium_game"},
			MinBushcraft: 3, BaseChance: 0.2, BaseHours: 1.6, ConditionLoss: 10, YieldMinKg: 0.5, YieldMaxKg: 3.2,
			RequiresCrafted: []string{"deadfall_kit", "wood_mallet"},
		},
		{
			ID: "snare_fence", Name: "Snare Fence",
			Description:  "Guided fence line with multiple noose points.",
			BiomeTags:    []string{"savanna", "badlands", "forest", "boreal", "mountain"},
			Targets:      []string{"small_game"},
			MinBushcraft: 2, BaseChance: 0.28, BaseHours: 1.4, ConditionLoss: 7, YieldMinKg: 0.25, YieldMaxKg: 1.3,
			RequiresCrafted: []string{"spring_snare_kit", "heavy_cordage"},
		},
		{
			ID: "fish_weir", Name: "Fish Weir",
			Description:  "Stake and funnel weir for shallow channels.",
			BiomeTags:    []string{"river", "delta", "wetlands", "coast"},
			Targets:      []string{"fish"},
			MinBushcraft: 2, BaseChance: 0.33, BaseHours: 2.2, ConditionLoss: 5, YieldMinKg: 0.4, YieldMaxKg: 2.8, NeedsWater: true,
			RequiresCrafted: []string{"fish_weir_stakes"},
		},
		{
			ID: "gill_net_set", Name: "Gill Net Set",
			Description:  "Set gill net in current seam for passive catches.",
			BiomeTags:    []string{"coast", "delta", "river", "lake", "wetlands"},
			Targets:      []string{"fish"},
			MinBushcraft: 3, BaseChance: 0.36, BaseHours: 1.8, ConditionLoss: 6, YieldMinKg: 0.5, YieldMaxKg: 3.5, NeedsWater: true,
			RequiresCrafted: []string{"gill_net"},
		},
		{
			ID: "trotline", Name: "Trotline",
			Description:  "Longline with multiple hooks for overnight fish catches.",
			BiomeTags:    []string{"river", "lake", "delta", "coast"},
			Targets:      []string{"fish"},
			MinBushcraft: 2, BaseChance: 0.31, BaseHours: 1.2, ConditionLoss: 4, YieldMinKg: 0.3, YieldMaxKg: 2.6, NeedsWater: true,
			RequiresCrafted: []string{"trotline_set"},
		},
		{
			ID: "eel_pot", Name: "Eel Pot",
			Description:  "Narrow-mouthed pot trap for eels and small fish.",
			BiomeTags:    []string{"river", "delta", "swamp", "wetlands", "coast"},
			Targets:      []string{"fish"},
			MinBushcraft: 2, BaseChance: 0.27, BaseHours: 1.5, ConditionLoss: 4, YieldMinKg: 0.2, YieldMaxKg: 1.6, NeedsWater: true,
			RequiresCrafted: []string{"fish_trap_basket"},
		},
		{
			ID: "reptile_noose", Name: "Reptile Noose",
			Description:  "Pole-noose setup for lizards and snakes.",
			BiomeTags:    []string{"desert", "dry", "savanna", "jungle", "wetlands"},
			Targets:      []string{"reptile"},
			MinBushcraft: 2, BaseChance: 0.2, BaseHours: 0.9, ConditionLoss: 7, YieldMinKg: 0.15, YieldMaxKg: 1.8,
			RequiresCrafted: []string{"natural_twine"},
		},
		{
			ID: "crab_pot", Name: "Crab Pot",
			Description:  "Weighted pot trap for tidal and estuary shellfish.",
			BiomeTags:    []string{"coast", "delta", "island", "wetlands"},
			Targets:      []string{"fish"},
			MinBushcraft: 1, BaseChance: 0.23, BaseHours: 1.4, ConditionLoss: 4, YieldMinKg: 0.2, YieldMaxKg: 1.4, NeedsWater: true,
			RequiresCrafted: []string{"fish_trap_basket", "heavy_cordage"},
		},
	}
}

func TrapsForBiome(biome string) []TrapSpec {
	norm := normalizeBiome(biome)
	catalog := TrapCatalog()
	out := make([]TrapSpec, 0, len(catalog))
	for _, trap := range catalog {
		for _, tag := range trap.BiomeTags {
			if strings.Contains(norm, normalizeBiome(tag)) {
				out = append(out, trap)
				break
			}
		}
	}
	if len(out) == 0 {
		out = append(out, catalog[0], catalog[1], catalog[3])
	}
	return out
}

func trapByID(biome, id string) (TrapSpec, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return TrapSpec{}, false
	}
	for _, trap := range TrapsForBiome(biome) {
		if trap.ID == id {
			return trap, true
		}
	}
	return TrapSpec{}, false
}

func hasCraftedItem(crafted []string, id string) bool {
	return slices.Contains(crafted, strings.ToLower(strings.TrimSpace(id)))
}

func trapRoll(seed int64, day int, playerID int, trapID string, idx int) float64 {
	rng := seededRNG(seedFromLabel(seed, fmt.Sprintf("trap:%d:%d:%d:%s:%d", day, playerID, idx, trapID, seed)))
	return rng.Float64()
}

func (s *RunState) SetTrap(playerID int, trapID string) (TrapSetResult, error) {
	if s == nil {
		return TrapSetResult{}, fmt.Errorf("run state is nil")
	}
	player, ok := s.playerByID(playerID)
	if !ok {
		return TrapSetResult{}, fmt.Errorf("player %d not found", playerID)
	}
	trap, ok := trapByID(s.Scenario.Biome, trapID)
	if !ok {
		return TrapSetResult{}, fmt.Errorf("trap not available in biome: %s", trapID)
	}
	effective := player.Bushcraft + player.Crafting/20 + player.Hunting/30 + player.Fishing/35 + player.Agility + positiveTraitModifier(player.Traits)/2 + negativeTraitModifier(player.Traits)/2
	if effective < trap.MinBushcraft {
		return TrapSetResult{}, fmt.Errorf("requires bushcraft %+d", trap.MinBushcraft)
	}
	for _, itemID := range trap.RequiresCrafted {
		if !hasCraftedItem(s.CraftedItems, itemID) {
			return TrapSetResult{}, fmt.Errorf("requires crafted item: %s", itemID)
		}
	}
	for _, req := range trap.RequiresResources {
		if s.resourceQty(req.ID) < req.Qty {
			return TrapSetResult{}, fmt.Errorf("requires resource: %s %.1f", req.ID, req.Qty)
		}
	}
	for _, kit := range trap.RequiresKit {
		if !playerHasKitItem(player, s.Config.IssuedKit, kit) {
			return TrapSetResult{}, fmt.Errorf("requires kit item: %s", kit)
		}
	}
	for _, req := range trap.RequiresResources {
		_ = s.consumeResourceStock(req.ID, req.Qty)
	}

	qualityScore := float64(effective) + float64(player.MentalStrength)/2.0 + float64(player.Crafting)/25.0
	if playerHasKitItem(player, s.Config.IssuedKit, KitSnareWire) {
		qualityScore += 0.8
	}
	if playerHasKitItem(player, s.Config.IssuedKit, KitParacord50ft) {
		qualityScore += 0.7
	}
	rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("trapset:%s:%d:%d", trap.ID, s.Day, playerID)))
	qualityScore += rng.Float64()*1.4 - 0.7
	quality := qualityFromScore(qualityScore)
	effectiveness := trap.BaseChance + qualityCatchBonus(quality) + float64(player.Crafting)/400.0 + float64(player.Hunting+player.Fishing)/1200.0
	if trap.NeedsWater && !trapBiomeHasWater(s.Scenario.Biome) {
		effectiveness -= 0.18
	}
	effectiveness = clampFloat(effectiveness, 0.04, 0.9)

	s.PlacedTraps = append(s.PlacedTraps, PlacedTrap{
		ID:            trap.ID,
		Name:          trap.Name,
		SetByPlayerID: playerID,
		Quality:       quality,
		Effectiveness: effectiveness,
		Condition:     100,
		Targets:       append([]string(nil), trap.Targets...),
		Armed:         true,
	})

	applySkillEffort(&player.Crafting, int(math.Round(trap.BaseHours*20)), true)
	if slices.Contains(trap.Targets, "fish") {
		applySkillEffort(&player.Fishing, int(math.Round(trap.BaseHours*14)), true)
	} else {
		applySkillEffort(&player.Hunting, int(math.Round(trap.BaseHours*14)), true)
	}
	applySkillEffort(&player.Trapping, int(math.Round(trap.BaseHours*18)), true)
	player.Energy = clamp(player.Energy-int(math.Ceil(trap.BaseHours*3.2)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(trap.BaseHours*1.6)), 0, 100)
	player.Morale = clamp(player.Morale+1, 0, 100)
	refreshEffectBars(player)

	return TrapSetResult{
		Trap:    trap,
		Quality: quality,
		Hours:   trap.BaseHours,
		Chance:  effectiveness,
	}, nil
}

func trapBiomeHasWater(biome string) bool {
	n := normalizeBiome(biome)
	return strings.Contains(n, "river") ||
		strings.Contains(n, "lake") ||
		strings.Contains(n, "delta") ||
		strings.Contains(n, "coast") ||
		strings.Contains(n, "wetland") ||
		strings.Contains(n, "swamp") ||
		strings.Contains(n, "island")
}

func (s *RunState) progressTrapsDaily() {
	if s == nil || len(s.PlacedTraps) == 0 {
		return
	}
	for i := range s.PlacedTraps {
		trap := &s.PlacedTraps[i]
		if trap.LastResolvedDay == s.Day {
			continue
		}
		spec, ok := trapByID(s.Scenario.Biome, trap.ID)
		if !ok {
			trap.LastResolvedDay = s.Day
			continue
		}
		loss := spec.ConditionLoss
		if isRainyWeather(s.Weather.Type) {
			loss += 1
		}
		if isSevereWeather(s.Weather.Type) {
			loss += 2
		}
		trap.Condition = clamp(trap.Condition-loss, 0, 100)
		if trap.Condition <= 0 {
			trap.Armed = false
			trap.Broken++
			trap.LastResolvedDay = s.Day
			continue
		}
		if !trap.Armed {
			trap.LastResolvedDay = s.Day
			continue
		}

		chance := trap.Effectiveness
		chance += qualityCatchBonus(trap.Quality) * 0.5
		chance *= clampFloat(float64(trap.Condition)/100.0, 0.2, 1.0)
		if spec.NeedsWater && !trapBiomeHasWater(s.Scenario.Biome) {
			chance -= 0.2
		}
		switch s.Weather.Type {
		case WeatherStorm, WeatherBlizzard:
			chance -= 0.08
		case WeatherHeavyRain:
			chance -= 0.05
		case WeatherClear:
			chance += 0.02
		}
		chance = clampFloat(chance, 0.03, 0.93)

		if trapRoll(s.Config.Seed, s.Day, trap.SetByPlayerID, trap.ID, i) <= chance {
			rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("trapyield:%s:%d:%d:%d", trap.ID, s.Day, trap.SetByPlayerID, i)))
			catchKg := spec.YieldMinKg
			if spec.YieldMaxKg > spec.YieldMinKg {
				catchKg += rng.Float64() * (spec.YieldMaxKg - spec.YieldMinKg)
			}
			catchKg = math.Round(catchKg*100) / 100
			target := spec.Targets[rng.IntN(len(spec.Targets))]
			trap.PendingCatchKg += catchKg
			trap.PendingCatchType = target
			trap.Armed = false
			trap.Successes++
		} else {
			trap.Failures++
		}
		trap.LastResolvedDay = s.Day
	}
}

func (s *RunState) CheckTraps() TrapCheckResult {
	result := TrapCheckResult{}
	if s == nil || len(s.PlacedTraps) == 0 {
		return result
	}
	for i := range s.PlacedTraps {
		trap := &s.PlacedTraps[i]
		result.Checked++
		if trap.Condition <= 0 {
			result.Broken++
			continue
		}
		if trap.PendingCatchKg > 0 {
			item := InventoryItem{
				ID:       trapCatchItemID(trap.PendingCatchType),
				Name:     trapCatchName(trap.PendingCatchType),
				Unit:     "kg",
				Qty:      math.Round(trap.PendingCatchKg*10) / 10,
				WeightKg: 1,
				Category: "food",
				Quality:  string(trap.Quality),
			}
			if err := s.addCampInventoryItem(item); err != nil {
				result.CampOverflow++
			} else {
				result.CollectedKg += item.Qty
				if player, ok := s.playerByID(trap.SetByPlayerID); ok {
					effort := int(math.Round(item.Qty * 10))
					if effort < 4 {
						effort = 4
					}
					applySkillEffort(&player.Trapping, effort, true)
				}
			}
			trap.PendingCatchKg = 0
			trap.PendingCatchType = ""
		}
		trap.Armed = true
		trap.LastCheckedDay = s.Day
		result.Rearmed++
	}
	return result
}

func trapCatchItemID(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "fish":
		return "fish_carcass"
	case "bird":
		return "bird_carcass"
	case "medium_game":
		return "medium_game_carcass"
	case "reptile":
		return "reptile_carcass"
	default:
		return "small_game_carcass"
	}
}

func trapCatchName(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "fish":
		return "Fish Carcass"
	case "bird":
		return "Bird Carcass"
	case "medium_game":
		return "Medium Game Carcass"
	case "reptile":
		return "Reptile Carcass"
	default:
		return "Small Game Carcass"
	}
}

func formatTrapStatus(traps []PlacedTrap) string {
	if len(traps) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(traps))
	for i, trap := range traps {
		mode := "armed"
		if !trap.Armed {
			mode = "awaiting_check"
		}
		pending := ""
		if trap.PendingCatchKg > 0 {
			pending = fmt.Sprintf(" catch %.2fkg %s", trap.PendingCatchKg, trap.PendingCatchType)
		}
		parts = append(parts, fmt.Sprintf("#%d %s(%s) %s cond:%d%%%s", i+1, trap.ID, trap.Quality, mode, trap.Condition, pending))
	}
	return strings.Join(parts, " | ")
}
