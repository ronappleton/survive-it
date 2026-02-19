package game

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Discovery summary:
// - Encounter rolls are deterministic and hash-seeded by seed/cell/day/time/action/roll index.
// - Tables are channel-based (mammal/bird/fish/insect) and already modulated by pressure/disturbance.
// - Wiring species IDs to AnimalCatalog keeps encounters consistent with expanded content registries.
type WildlifeEncounter struct {
	Channel        string
	Species        string
	Message        string
	Prey           bool
	Predator       bool
	EnergyDelta    int
	HydrationDelta int
	MoraleDelta    int
}

type encounterSpecies struct {
	AnimalID  string
	Name      string
	Weight    int
	Prey      bool
	Predator  bool
	Scavenger bool
	Nocturnal bool
}

const (
	encounterChannelMammal = 0
	encounterChannelBird   = 1
	encounterChannelFish   = 2
	encounterChannelInsect = 3
)

func deterministicEncounterRoll(seed int64, x, y, day int, block TimeBlock, action string, rollIndex int, salt string) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d:%d:%d:%s:%s:%d:%s", seed, x, y, day, block, action, rollIndex, salt)))
	return float64(h.Sum64()&0x7fffffff) / float64(0x7fffffff)
}

func pickWeightedIndex(seed int64, x, y, day int, block TimeBlock, action string, rollIndex int, salt string, weights []int) int {
	total := 0
	for _, w := range weights {
		if w > 0 {
			total += w
		}
	}
	if total <= 0 {
		return -1
	}
	roll := deterministicEncounterRoll(seed, x, y, day, block, action, rollIndex, salt)
	target := int(roll * float64(total))
	if target >= total {
		target = total - 1
	}
	running := 0
	for i, w := range weights {
		if w <= 0 {
			continue
		}
		running += w
		if target < running {
			return i
		}
	}
	return -1
}

func wildlifeNameByID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return ""
	}
	for _, animal := range AnimalCatalog() {
		if animal.ID == id {
			return animal.Name
		}
	}
	return ""
}

func speciesEntry(id string, weight int, prey bool, predator bool, scavenger bool, nocturnal bool) encounterSpecies {
	name := wildlifeNameByID(id)
	if name == "" {
		name = strings.ReplaceAll(id, "_", " ")
	}
	return encounterSpecies{
		AnimalID:  id,
		Name:      name,
		Weight:    weight,
		Prey:      prey,
		Predator:  predator,
		Scavenger: scavenger,
		Nocturnal: nocturnal,
	}
}

func biomeEncounterList(biome uint8, channel string) []encounterSpecies {
	switch channel {
	case "mammal":
		switch biome {
		case TopoBiomeDesert:
			return []encounterSpecies{
				speciesEntry("rabbit", 24, true, false, false, true),
				speciesEntry("kangaroo", 12, true, false, false, false),
				speciesEntry("camel", 6, true, false, false, false),
				speciesEntry("iguana", 14, true, false, false, true),
				speciesEntry("monitor_lizard", 10, false, true, false, true),
				speciesEntry("fox", 10, false, true, true, true),
				speciesEntry("coyote", 12, false, true, true, true),
				speciesEntry("jackal", 9, false, true, true, true),
			}
		case TopoBiomeJungle, TopoBiomeSwamp:
			return []encounterSpecies{
				speciesEntry("boar", 18, true, false, false, false),
				speciesEntry("deer", 14, true, false, false, false),
				speciesEntry("capybara", 16, true, false, false, false),
				speciesEntry("muskrat", 14, true, false, false, true),
				speciesEntry("alligator", 9, false, true, false, true),
				speciesEntry("caiman", 8, false, true, false, true),
				speciesEntry("crocodile", 6, false, true, false, true),
				speciesEntry("jaguar", 8, false, true, false, true),
				speciesEntry("python", 7, false, true, false, true),
			}
		case TopoBiomeTundra, TopoBiomeBoreal:
			return []encounterSpecies{
				speciesEntry("rabbit", 20, true, false, false, true),
				speciesEntry("caribou", 14, true, false, false, false),
				speciesEntry("moose", 8, true, false, false, false),
				speciesEntry("beaver", 8, true, false, false, true),
				speciesEntry("fox", 11, false, true, true, true),
				speciesEntry("wolf", 12, false, true, false, true),
				speciesEntry("brown_bear", 7, false, true, true, true),
			}
		case TopoBiomeMountain:
			return []encounterSpecies{
				speciesEntry("mountain_goat", 16, true, false, false, false),
				speciesEntry("deer", 14, true, false, false, false),
				speciesEntry("elk", 10, true, false, false, false),
				speciesEntry("rabbit", 16, true, false, false, true),
				speciesEntry("cougar", 10, false, true, false, true),
				speciesEntry("wolf", 8, false, true, false, true),
				speciesEntry("black_bear", 7, false, true, true, true),
			}
		default:
			return []encounterSpecies{
				speciesEntry("rabbit", 22, true, false, false, true),
				speciesEntry("deer", 18, true, false, false, false),
				speciesEntry("boar", 12, true, false, false, false),
				speciesEntry("beaver", 8, true, false, false, true),
				speciesEntry("coyote", 10, false, true, true, true),
				speciesEntry("wolf", 8, false, true, false, true),
				speciesEntry("black_bear", 7, false, true, true, true),
			}
		}
	case "bird":
		switch biome {
		case TopoBiomeWetland, TopoBiomeSwamp:
			return []encounterSpecies{
				speciesEntry("duck", 22, true, false, false, false),
				speciesEntry("heron", 14, true, false, false, false),
				speciesEntry("egret", 12, true, false, false, false),
				speciesEntry("goose", 10, true, false, false, false),
				speciesEntry("crow", 12, false, false, true, false),
				speciesEntry("raven", 8, false, false, true, false),
				speciesEntry("kingfisher", 10, true, false, false, false),
			}
		case TopoBiomeDesert:
			return []encounterSpecies{
				speciesEntry("quail", 24, true, false, false, false),
				speciesEntry("dove", 20, true, false, false, false),
				speciesEntry("partridge", 14, true, false, false, false),
				speciesEntry("raven", 12, false, false, true, false),
				speciesEntry("crow", 9, false, false, true, false),
			}
		default:
			return []encounterSpecies{
				speciesEntry("grouse", 18, true, false, false, false),
				speciesEntry("pheasant", 16, true, false, false, false),
				speciesEntry("wild_turkey", 10, true, false, false, false),
				speciesEntry("quail", 14, true, false, false, false),
				speciesEntry("seagull", 8, true, false, true, false),
				speciesEntry("pelican", 6, true, false, false, false),
				speciesEntry("crow", 10, false, false, true, false),
				speciesEntry("raven", 8, false, false, true, false),
			}
		}
	case "fish":
		switch biome {
		case TopoBiomeSwamp, TopoBiomeWetland:
			return []encounterSpecies{
				speciesEntry("catfish", 20, true, false, false, true),
				speciesEntry("carp", 16, true, false, false, false),
				speciesEntry("tilapia", 16, true, false, false, false),
				speciesEntry("eel", 10, true, false, false, true),
				speciesEntry("bluegill", 14, true, false, false, false),
				speciesEntry("crappie", 12, true, false, false, false),
			}
		case TopoBiomeJungle:
			return []encounterSpecies{
				speciesEntry("tilapia", 18, true, false, false, false),
				speciesEntry("piranha", 14, true, false, false, true),
				speciesEntry("arapaima", 10, true, false, false, false),
				speciesEntry("catla", 14, true, false, false, false),
				speciesEntry("rohu", 14, true, false, false, false),
				speciesEntry("eel", 8, true, false, false, true),
			}
		case TopoBiomeDesert:
			return []encounterSpecies{
				speciesEntry("carp", 18, true, false, false, false),
				speciesEntry("catfish", 16, true, false, false, true),
				speciesEntry("tilapia", 14, true, false, false, false),
				speciesEntry("bluegill", 12, true, false, false, false),
			}
		case TopoBiomeTundra, TopoBiomeBoreal:
			return []encounterSpecies{
				speciesEntry("char", 18, true, false, false, false),
				speciesEntry("trout", 18, true, false, false, false),
				speciesEntry("grayling", 16, true, false, false, false),
				speciesEntry("walleye", 14, true, false, false, false),
				speciesEntry("northern_pike", 10, true, false, false, true),
				speciesEntry("perch", 12, true, false, false, false),
			}
		default:
			return []encounterSpecies{
				speciesEntry("trout", 18, true, false, false, false),
				speciesEntry("perch", 16, true, false, false, false),
				speciesEntry("walleye", 12, true, false, false, false),
				speciesEntry("catfish", 12, true, false, false, true),
				speciesEntry("sturgeon", 8, true, false, false, false),
				speciesEntry("eel", 8, true, false, false, true),
			}
		}
	default: // insect
		switch biome {
		case TopoBiomeSwamp, TopoBiomeWetland, TopoBiomeJungle:
			return []encounterSpecies{
				speciesEntry("mosquito", 30, false, false, false, true),
				speciesEntry("black_fly", 24, false, false, false, true),
				speciesEntry("midge", 20, false, false, false, true),
				speciesEntry("tick", 18, false, false, false, false),
				speciesEntry("leech", 18, false, false, false, true),
				speciesEntry("fire_ant", 12, false, false, false, false),
				speciesEntry("wasp", 10, false, false, false, false),
			}
		case TopoBiomeDesert:
			return []encounterSpecies{
				speciesEntry("fire_ant", 20, false, false, false, false),
				speciesEntry("midge", 16, false, false, false, true),
				speciesEntry("scorpion", 14, false, true, false, true),
				speciesEntry("wolf_spider", 10, false, true, false, true),
				speciesEntry("centipede", 10, false, true, false, true),
			}
		default:
			return []encounterSpecies{
				speciesEntry("midge", 22, false, false, false, true),
				speciesEntry("mosquito", 18, false, false, false, true),
				speciesEntry("tick", 14, false, false, false, false),
				speciesEntry("horsefly", 14, false, false, false, true),
				speciesEntry("bee", 10, false, false, false, false),
			}
		}
	}
}

func baseEncounterChanceForAction(action string) float64 {
	switch action {
	case "forage":
		return 0.18
	case "hunt":
		return 0.31
	case "fish":
		return 0.4
	default:
		return 0.11
	}
}

func biomeChanceModifier(biome uint8, action string) float64 {
	mod := 0.0
	switch biome {
	case TopoBiomeSwamp, TopoBiomeWetland:
		mod += 0.045
	case TopoBiomeJungle:
		mod += 0.035
	case TopoBiomeForest, TopoBiomeGrassland:
		mod += 0.01
	case TopoBiomeDesert:
		mod -= 0.03
	case TopoBiomeMountain:
		mod -= 0.012
	case TopoBiomeTundra, TopoBiomeBoreal:
		mod -= 0.02
	}
	if action == "fish" {
		switch biome {
		case TopoBiomeSwamp, TopoBiomeWetland:
			mod += 0.02
		case TopoBiomeDesert:
			mod -= 0.02
		}
	}
	return mod
}

func timeBlockChanceModifier(block TimeBlock, action string) float64 {
	switch block {
	case TimeBlockDawn:
		if action == "fish" {
			return 0.02
		}
		return 0.018
	case TimeBlockDusk:
		if action == "fish" {
			return 0.022
		}
		return 0.015
	case TimeBlockNight:
		switch action {
		case "hunt":
			return -0.02
		case "forage":
			return -0.025
		case "fish":
			return -0.01
		default:
			return -0.012
		}
	default:
		return 0
	}
}

func channelWeightsForEncounter(biome uint8, action string, block TimeBlock, nearWater bool) []int {
	channels := []int{30, 28, 0, 30}
	switch action {
	case "forage":
		channels = []int{20, 24, 6, 50}
	case "hunt":
		channels = []int{62, 23, 8, 7}
	case "fish":
		channels = []int{8, 12, 72, 8}
	}

	switch biome {
	case TopoBiomeSwamp, TopoBiomeWetland:
		channels[encounterChannelMammal] -= 4
		channels[encounterChannelBird] += 8
		channels[encounterChannelFish] += 10
		channels[encounterChannelInsect] += 22
	case TopoBiomeJungle:
		channels[encounterChannelMammal] += 2
		channels[encounterChannelBird] += 4
		channels[encounterChannelFish] += 6
		channels[encounterChannelInsect] += 18
	case TopoBiomeDesert:
		channels[encounterChannelMammal] -= 6
		channels[encounterChannelBird] -= 4
		channels[encounterChannelFish] -= 6
		channels[encounterChannelInsect] -= 10
	case TopoBiomeMountain:
		channels[encounterChannelMammal] += 4
		channels[encounterChannelBird] += 3
		channels[encounterChannelFish] -= 4
		channels[encounterChannelInsect] -= 10
	case TopoBiomeTundra, TopoBiomeBoreal:
		channels[encounterChannelMammal] += 6
		channels[encounterChannelBird] += 2
		channels[encounterChannelFish] -= 2
		channels[encounterChannelInsect] -= 14
	}

	switch block {
	case TimeBlockDawn:
		channels[encounterChannelMammal] += 8
		channels[encounterChannelBird] += 6
	case TimeBlockDusk:
		channels[encounterChannelMammal] += 7
		channels[encounterChannelBird] += 5
	case TimeBlockNight:
		channels[encounterChannelMammal] -= 4
		channels[encounterChannelBird] -= 12
		channels[encounterChannelFish] += 4
		channels[encounterChannelInsect] += 16
	default:
		channels[encounterChannelInsect] += 4
	}

	if !nearWater {
		channels[encounterChannelFish] = 0
	}
	for i := range channels {
		if channels[i] < 0 {
			channels[i] = 0
		}
	}
	return channels
}

func (s *RunState) isNearWater(x, y int) bool {
	if s == nil {
		return false
	}
	cell, ok := s.TopologyCellAt(x, y)
	if !ok {
		return false
	}
	if cell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake|TopoFlagCoast) != 0 {
		return true
	}
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			if ox == 0 && oy == 0 {
				continue
			}
			near, ok := s.TopologyCellAt(x+ox, y+oy)
			if !ok {
				continue
			}
			if near.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake) != 0 {
				return true
			}
		}
	}
	return false
}

func (s *RunState) RollWildlifeEncounter(playerID, x, y int, action string, rollIndex int) (WildlifeEncounter, bool) {
	if s == nil {
		return WildlifeEncounter{}, false
	}
	s.EnsureTopology()
	idx, ok := s.topoIndex(x, y)
	if !ok {
		return WildlifeEncounter{}, false
	}
	if playerID <= 0 {
		playerID = 1
	}
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		action = "move"
	}
	state := CellState{}
	if idx < len(s.CellStates) {
		state = s.CellStates[idx]
	}
	cell := s.Topology.Cells[idx]
	nearWater := s.isNearWater(x, y)
	season, okSeason := s.CurrentSeason()
	if !okSeason {
		season = SeasonAutumn
	}

	block := s.CurrentTimeBlock()
	baseChance := baseEncounterChanceForAction(action)
	baseChance += biomeChanceModifier(cell.Biome, action)
	baseChance += timeBlockChanceModifier(block, action)
	if nearWater {
		baseChance += 0.03
	}
	if action == "fish" && !nearWater {
		return WildlifeEncounter{}, false
	}
	chance := baseChance
	chance -= float64(state.Disturbance) * 0.0018
	if action == "hunt" || action == "fish" {
		chance -= float64(state.HuntPressure) * 0.0015
	}
	if action == "forage" {
		chance -= float64(state.Depletion) * 0.0015
	}
	if chance < 0.015 {
		chance = 0.015
	}
	if chance > 0.88 {
		chance = 0.88
	}
	triggerRoll := deterministicEncounterRoll(s.Config.Seed, x, y, s.Day, block, action, rollIndex, "trigger")
	if triggerRoll > chance {
		return WildlifeEncounter{}, false
	}

	channels := []string{"mammal", "bird", "fish", "insect"}
	weights := channelWeightsForEncounter(cell.Biome, action, block, nearWater)
	if !insectActivityAllowed(s.ActiveClimateProfile(), season, s.Weather.TemperatureC, cell.Biome) {
		weights[encounterChannelInsect] = 0
	}
	channelIdx := pickWeightedIndex(s.Config.Seed, x, y, s.Day, block, action, rollIndex, "channel", weights)
	if channelIdx < 0 || channelIdx >= len(channels) {
		return WildlifeEncounter{}, false
	}
	channel := channels[channelIdx]
	species := biomeEncounterList(cell.Biome, channel)
	species = filterEncounterSpeciesForClimate(species, channel, s.ActiveClimateProfile(), season, s.Weather.TemperatureC, cell.Biome)
	if len(species) == 0 {
		msg := "It's quiet. No wildlife activity matches these conditions."
		if channel == "insect" {
			msg = quietInsectMessage(s.ActiveClimateProfile(), season)
		}
		return WildlifeEncounter{
			Channel: "ambient",
			Message: msg,
		}, true
	}
	speciesWeights := make([]int, len(species))
	for i, sp := range species {
		w := sp.Weight
		if state.Disturbance > 0 {
			w -= int(state.Disturbance) / 4
		}
		if sp.Prey && state.HuntPressure > 0 {
			w -= int(state.HuntPressure) / 3
		}
		if (sp.Predator || sp.Scavenger) && state.CarcassToken > 0 {
			w += int(state.CarcassToken) * 4
		}
		if channel == "fish" && state.Depletion > 0 {
			w -= int(state.Depletion) / 5
		}
		switch block {
		case TimeBlockDawn, TimeBlockDusk:
			if sp.Prey && (channel == "mammal" || channel == "bird") {
				w += 4
			}
		case TimeBlockNight:
			if channel == "bird" && sp.Prey {
				w -= 8
			}
			if channel == "mammal" && sp.Prey {
				w -= 3
			}
			if sp.Predator {
				w += 3
			}
			if channel == "insect" {
				w += 8
			}
		}
		if sp.Nocturnal {
			if block == TimeBlockNight || block == TimeBlockDusk {
				w += 5
			} else {
				w -= 3
			}
		}
		if cell.Biome == TopoBiomeDesert && channel == "insect" && sp.AnimalID != "fire_ant" && sp.AnimalID != "scorpion" {
			w -= 6
		}
		if (cell.Biome == TopoBiomeSwamp || cell.Biome == TopoBiomeWetland || cell.Biome == TopoBiomeJungle) && channel == "insect" {
			w += 5
		}
		if w < 1 {
			w = 1
		}
		speciesWeights[i] = w
	}
	speciesIdx := pickWeightedIndex(s.Config.Seed, x, y, s.Day, block, action, rollIndex, "species", speciesWeights)
	if speciesIdx < 0 || speciesIdx >= len(species) {
		return WildlifeEncounter{}, false
	}
	selected := species[speciesIdx]

	event := WildlifeEncounter{
		Channel:  channel,
		Species:  selected.Name,
		Prey:     selected.Prey,
		Predator: selected.Predator,
	}
	switch channel {
	case "mammal":
		if selected.Predator {
			event.Message = fmt.Sprintf("Predator sign nearby: %s.", selected.Name)
			event.MoraleDelta = -2
		} else {
			event.Message = fmt.Sprintf("Wildlife sign: %s activity in this area.", selected.Name)
			if action == "hunt" {
				event.MoraleDelta = 1
			}
		}
	case "bird":
		event.Message = fmt.Sprintf("Bird sign: %s around your position.", selected.Name)
		if action == "hunt" && selected.Prey {
			event.MoraleDelta = 1
		}
	case "fish":
		event.Message = fmt.Sprintf("Fish sign: %s in nearby water.", selected.Name)
		if action == "fish" {
			event.MoraleDelta = 1
		}
	default:
		event.Message = fmt.Sprintf("Insects are active: %s.", selected.Name)
		event.MoraleDelta = -1
		if block == TimeBlockNight || block == TimeBlockDusk {
			event.EnergyDelta = -1
			event.HydrationDelta = -1
			event.MoraleDelta = -2
		}
	}
	return event, true
}

func wildlifeEncounterSpeciesIDs() []string {
	biomes := []uint8{
		TopoBiomeForest,
		TopoBiomeGrassland,
		TopoBiomeJungle,
		TopoBiomeWetland,
		TopoBiomeSwamp,
		TopoBiomeDesert,
		TopoBiomeMountain,
		TopoBiomeTundra,
		TopoBiomeBoreal,
	}
	channels := []string{"mammal", "bird", "fish", "insect"}
	seen := map[string]bool{}
	out := make([]string, 0, 128)
	for _, biome := range biomes {
		for _, channel := range channels {
			for _, sp := range biomeEncounterList(biome, channel) {
				id := strings.TrimSpace(strings.ToLower(sp.AnimalID))
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}
