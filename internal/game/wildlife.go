package game

import (
	"fmt"
	"hash/fnv"
	"strings"
)

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
	Name      string
	Weight    int
	Prey      bool
	Predator  bool
	Scavenger bool
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

func biomeEncounterList(biome uint8, channel string) []encounterSpecies {
	switch channel {
	case "mammal":
		switch biome {
		case TopoBiomeDesert:
			return []encounterSpecies{
				{Name: "rabbit", Weight: 30, Prey: true},
				{Name: "lizard", Weight: 28, Prey: true},
				{Name: "fox", Weight: 14, Predator: true},
				{Name: "coyote", Weight: 10, Predator: true},
			}
		case TopoBiomeJungle, TopoBiomeSwamp:
			return []encounterSpecies{
				{Name: "boar", Weight: 22, Prey: true},
				{Name: "small deer", Weight: 18, Prey: true},
				{Name: "monkey", Weight: 14, Prey: true},
				{Name: "jaguar sign", Weight: 9, Predator: true},
				{Name: "wild dog pack", Weight: 8, Predator: true},
			}
		case TopoBiomeTundra, TopoBiomeBoreal:
			return []encounterSpecies{
				{Name: "hare", Weight: 30, Prey: true},
				{Name: "caribou sign", Weight: 16, Prey: true},
				{Name: "fox", Weight: 13, Predator: true},
				{Name: "wolf", Weight: 11, Predator: true},
			}
		case TopoBiomeMountain:
			return []encounterSpecies{
				{Name: "goat sign", Weight: 20, Prey: true},
				{Name: "hare", Weight: 20, Prey: true},
				{Name: "cougar sign", Weight: 10, Predator: true},
				{Name: "wolf", Weight: 8, Predator: true},
			}
		default:
			return []encounterSpecies{
				{Name: "rabbit", Weight: 28, Prey: true},
				{Name: "deer sign", Weight: 22, Prey: true},
				{Name: "wild pig", Weight: 14, Prey: true},
				{Name: "coyote", Weight: 10, Predator: true},
				{Name: "wolf", Weight: 8, Predator: true},
			}
		}
	case "bird":
		switch biome {
		case TopoBiomeWetland, TopoBiomeSwamp:
			return []encounterSpecies{
				{Name: "duck", Weight: 26, Prey: true},
				{Name: "heron", Weight: 18, Prey: true},
				{Name: "crow", Weight: 16, Scavenger: true},
				{Name: "vulture", Weight: 10, Scavenger: true},
			}
		case TopoBiomeDesert:
			return []encounterSpecies{
				{Name: "quail", Weight: 24, Prey: true},
				{Name: "dove", Weight: 20, Prey: true},
				{Name: "raven", Weight: 14, Scavenger: true},
				{Name: "hawk", Weight: 9, Predator: true},
			}
		default:
			return []encounterSpecies{
				{Name: "grouse", Weight: 22, Prey: true},
				{Name: "songbird", Weight: 24, Prey: true},
				{Name: "crow", Weight: 15, Scavenger: true},
				{Name: "hawk", Weight: 9, Predator: true},
			}
		}
	case "fish":
		return []encounterSpecies{
			{Name: "trout", Weight: 28, Prey: true},
			{Name: "perch", Weight: 24, Prey: true},
			{Name: "catfish", Weight: 18, Prey: true},
			{Name: "eel", Weight: 10, Prey: true},
		}
	default: // insect
		switch biome {
		case TopoBiomeSwamp, TopoBiomeWetland, TopoBiomeJungle:
			return []encounterSpecies{
				{Name: "mosquitoes", Weight: 34},
				{Name: "biting flies", Weight: 24},
				{Name: "ticks", Weight: 16},
				{Name: "ants", Weight: 12},
			}
		case TopoBiomeDesert:
			return []encounterSpecies{
				{Name: "ants", Weight: 26},
				{Name: "gnats", Weight: 22},
				{Name: "scorpions", Weight: 10},
			}
		default:
			return []encounterSpecies{
				{Name: "gnats", Weight: 24},
				{Name: "mosquitoes", Weight: 18},
				{Name: "ticks", Weight: 14},
				{Name: "flies", Weight: 20},
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
	channelIdx := pickWeightedIndex(s.Config.Seed, x, y, s.Day, block, action, rollIndex, "channel", weights)
	if channelIdx < 0 || channelIdx >= len(channels) {
		return WildlifeEncounter{}, false
	}
	channel := channels[channelIdx]
	species := biomeEncounterList(cell.Biome, channel)
	if len(species) == 0 {
		return WildlifeEncounter{}, false
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
		if cell.Biome == TopoBiomeDesert && channel == "insect" && sp.Name != "ants" && sp.Name != "scorpions" {
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
