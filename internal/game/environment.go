package game

import "strings"

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
