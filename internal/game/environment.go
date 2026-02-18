package game

import (
	"fmt"
	"hash/fnv"
	"strings"
)

type TemperatureRange struct {
	MinC int
	MaxC int
}

func WildlifeForBiome(biome string) []string {
	b := strings.ToLower(strings.TrimSpace(biome))
	switch {
	case strings.Contains(b, "rainforest"), strings.Contains(b, "temperate_rainforest"), strings.Contains(b, "vancouver"):
		return []string{"Black Bear", "Grizzly Bear", "Wolf", "Cougar", "Elk", "Deer"}
	case strings.Contains(b, "boreal"), strings.Contains(b, "subarctic"), strings.Contains(b, "arctic"), strings.Contains(b, "tundra"):
		return []string{"Brown Bear", "Wolf", "Moose", "Caribou", "Wolverine", "Arctic Fox"}
	case strings.Contains(b, "mountain"), strings.Contains(b, "highland"), strings.Contains(b, "montane"):
		return []string{"Black Bear", "Mountain Lion", "Goat", "Deer", "Wolf", "Boar"}
	case strings.Contains(b, "jungle"), strings.Contains(b, "tropical"), strings.Contains(b, "wetlands"), strings.Contains(b, "swamp"), strings.Contains(b, "island"):
		return []string{"Wild Boar", "Monkey", "Crocodile", "Snake", "Big Cat", "Tapir"}
	case strings.Contains(b, "savanna"), strings.Contains(b, "badlands"):
		return []string{"Hyena", "Leopard", "Buffalo", "Warthog", "Antelope", "Jackal"}
	case strings.Contains(b, "desert"), strings.Contains(b, "dry"):
		return []string{"Coyote", "Fox", "Camel", "Lizard", "Snake", "Scorpion"}
	case strings.Contains(b, "coast"), strings.Contains(b, "delta"), strings.Contains(b, "lake"):
		return []string{"Bear", "Wolf", "Otter", "Seal", "Deer", "Waterfowl"}
	default:
		return []string{"Deer", "Boar", "Wolf", "Bear"}
	}
}

func InsectsForBiome(biome string) []string {
	b := strings.ToLower(strings.TrimSpace(biome))
	switch {
	case strings.Contains(b, "rainforest"), strings.Contains(b, "jungle"), strings.Contains(b, "wet"), strings.Contains(b, "swamp"), strings.Contains(b, "wetlands"), strings.Contains(b, "island"):
		return []string{"Mosquitoes", "Ticks", "Sandflies", "Leeches", "Ants"}
	case strings.Contains(b, "boreal"), strings.Contains(b, "subarctic"), strings.Contains(b, "lake"), strings.Contains(b, "delta"):
		return []string{"Mosquitoes", "Blackflies", "Ticks"}
	case strings.Contains(b, "desert"), strings.Contains(b, "dry"), strings.Contains(b, "savanna"), strings.Contains(b, "badlands"):
		return []string{"Flies", "Ants", "Scorpions", "Beetles"}
	case strings.Contains(b, "arctic"), strings.Contains(b, "tundra"), strings.Contains(b, "winter"):
		return []string{"Biting Midges", "Mosquitoes (seasonal)"}
	default:
		return []string{"Mosquitoes", "Ticks", "Flies"}
	}
}

func TemperatureRangeForBiome(biome string) TemperatureRange {
	b := strings.ToLower(strings.TrimSpace(biome))
	switch {
	case strings.Contains(b, "forest"):
		return TemperatureRange{MinC: 4, MaxC: 22}
	case strings.Contains(b, "rainforest"), strings.Contains(b, "temperate_rainforest"), strings.Contains(b, "vancouver"):
		return TemperatureRange{MinC: 2, MaxC: 18}
	case strings.Contains(b, "boreal"), strings.Contains(b, "subarctic"), strings.Contains(b, "arctic"), strings.Contains(b, "tundra"):
		return TemperatureRange{MinC: -25, MaxC: 5}
	case strings.Contains(b, "mountain"), strings.Contains(b, "highland"), strings.Contains(b, "montane"):
		return TemperatureRange{MinC: -8, MaxC: 16}
	case strings.Contains(b, "jungle"), strings.Contains(b, "tropical"), strings.Contains(b, "wetlands"), strings.Contains(b, "swamp"), strings.Contains(b, "island"):
		return TemperatureRange{MinC: 22, MaxC: 37}
	case strings.Contains(b, "savanna"), strings.Contains(b, "badlands"):
		return TemperatureRange{MinC: 14, MaxC: 36}
	case strings.Contains(b, "desert"), strings.Contains(b, "dry"):
		return TemperatureRange{MinC: 3, MaxC: 45}
	case strings.Contains(b, "coast"), strings.Contains(b, "delta"), strings.Contains(b, "lake"):
		return TemperatureRange{MinC: 4, MaxC: 24}
	default:
		return TemperatureRange{MinC: 8, MaxC: 24}
	}
}

func TemperatureForDayCelsius(biome string, day int) int {
	r := TemperatureRangeForBiome(biome)
	if r.MaxC <= r.MinC {
		return r.MinC
	}
	if day < 1 {
		day = 1
	}
	span := r.MaxC - r.MinC

	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s:%d", strings.ToLower(strings.TrimSpace(biome)), day)))
	offset := int(h.Sum32() % uint32(span+1))
	return r.MinC + offset
}
