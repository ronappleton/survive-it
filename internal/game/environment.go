package game

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Discovery summary:
// - Biome-to-weather and biome-to-insect defaults live here and feed day/weather + look text helpers.
// - Insect lists are now neutral descriptors; temperature/season gating is enforced in climate filters.
// - This keeps base biome catalogs reusable while avoiding season-only contradictions.

type TemperatureRange struct {
	MinC int
	MaxC int
}

type WeatherType string

const (
	WeatherSunny     WeatherType = "sunny"
	WeatherClear     WeatherType = "clear"
	WeatherCloudy    WeatherType = "cloudy"
	WeatherRain      WeatherType = "rain"
	WeatherHeavyRain WeatherType = "heavy_rain"
	WeatherStorm     WeatherType = "storm"
	WeatherSnow      WeatherType = "snow"
	WeatherBlizzard  WeatherType = "blizzard"
	WeatherWindy     WeatherType = "windy"
	WeatherHeatwave  WeatherType = "heatwave"
)

type WeatherState struct {
	Day          int         `json:"day"`
	Type         WeatherType `json:"type"`
	TemperatureC int         `json:"temperature_c"`
	StreakDays   int         `json:"streak_days"`
}

type weightedWeather struct {
	Type   WeatherType
	Weight int
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
		return []string{"Biting Midges", "Blackflies", "Ticks"}
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

func WeatherLabel(weather WeatherType) string {
	switch weather {
	case WeatherSunny:
		return "Sunny"
	case WeatherClear:
		return "Clear"
	case WeatherCloudy:
		return "Cloudy"
	case WeatherRain:
		return "Rain"
	case WeatherHeavyRain:
		return "Heavy Rain"
	case WeatherStorm:
		return "Storm"
	case WeatherSnow:
		return "Snow"
	case WeatherBlizzard:
		return "Blizzard"
	case WeatherWindy:
		return "Windy"
	case WeatherHeatwave:
		return "Heatwave"
	default:
		return "Unknown"
	}
}

func WeatherTypesForBiomeSeason(biome string, season SeasonID) []WeatherType {
	weights := weatherWeightsForBiomeSeason(biome, season)
	out := make([]WeatherType, 0, len(weights))
	seen := map[WeatherType]bool{}
	for _, entry := range weights {
		if entry.Weight <= 0 || seen[entry.Type] {
			continue
		}
		seen[entry.Type] = true
		out = append(out, entry.Type)
	}
	return out
}

func WeatherForDay(seed int64, biome string, season SeasonID, day int) WeatherType {
	if day < 1 {
		day = 1
	}

	weights := weatherWeightsForBiomeSeason(biome, season)
	if len(weights) == 0 {
		return WeatherClear
	}

	totalWeight := 0
	for _, entry := range weights {
		if entry.Weight > 0 {
			totalWeight += entry.Weight
		}
	}

	if totalWeight <= 0 {
		return WeatherClear
	}

	roll := deterministicWeatherRoll(seed, biome, season, day) % totalWeight
	cumulative := 0
	for _, entry := range weights {
		if entry.Weight <= 0 {
			continue
		}
		cumulative += entry.Weight
		if roll < cumulative {
			return entry.Type
		}
	}

	return weights[len(weights)-1].Type
}

func TemperatureForDayWithWeatherCelsius(seed int64, biome string, season SeasonID, day int, weather WeatherType) int {
	base := TemperatureForDayCelsius(biome, day)

	switch season {
	case SeasonWinter:
		base -= 4
	case SeasonWet:
		base -= 1
	case SeasonDry:
		base += 2
	}

	adjusted := base + weatherTemperatureDelta(weather)
	rangeForBiome := TemperatureRangeForBiome(biome)
	return clamp(adjusted, rangeForBiome.MinC-12, rangeForBiome.MaxC+12)
}

func weatherTemperatureDelta(weather WeatherType) int {
	switch weather {
	case WeatherSunny:
		return 2
	case WeatherClear:
		return 1
	case WeatherCloudy:
		return -1
	case WeatherRain:
		return -2
	case WeatherHeavyRain:
		return -4
	case WeatherStorm:
		return -5
	case WeatherSnow:
		return -6
	case WeatherBlizzard:
		return -10
	case WeatherWindy:
		return -2
	case WeatherHeatwave:
		return 6
	default:
		return 0
	}
}

func weatherWeightsForBiomeSeason(biome string, season SeasonID) []weightedWeather {
	switch {
	case biomeIsArctic(biome):
		return arcticWeatherWeights(season)
	case biomeIsDesertOrDry(biome):
		return desertWeatherWeights(season)
	case biomeIsTropicalWet(biome):
		return tropicalWetWeatherWeights(season)
	case strings.Contains(normalizeBiome(biome), "savanna"), strings.Contains(normalizeBiome(biome), "badlands"):
		return savannaWeatherWeights(season)
	case strings.Contains(normalizeBiome(biome), "mountain"), strings.Contains(normalizeBiome(biome), "highland"), strings.Contains(normalizeBiome(biome), "montane"), strings.Contains(normalizeBiome(biome), "steppe"):
		return mountainWeatherWeights(season)
	case strings.Contains(normalizeBiome(biome), "coast"), strings.Contains(normalizeBiome(biome), "delta"), strings.Contains(normalizeBiome(biome), "lake"):
		return coastWeatherWeights(season)
	default:
		return temperateWeatherWeights(season)
	}
}

func arcticWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonWinter {
		return []weightedWeather{
			{Type: WeatherCloudy, Weight: 28},
			{Type: WeatherSnow, Weight: 30},
			{Type: WeatherBlizzard, Weight: 16},
			{Type: WeatherWindy, Weight: 6},
			{Type: WeatherSunny, Weight: 10},
			{Type: WeatherClear, Weight: 10},
		}
	}

	return []weightedWeather{
		{Type: WeatherCloudy, Weight: 26},
		{Type: WeatherSnow, Weight: 20},
		{Type: WeatherBlizzard, Weight: 6},
		{Type: WeatherWindy, Weight: 14},
		{Type: WeatherRain, Weight: 8},
		{Type: WeatherSunny, Weight: 10},
		{Type: WeatherClear, Weight: 16},
	}
}

func desertWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonDry {
		return []weightedWeather{
			{Type: WeatherSunny, Weight: 34},
			{Type: WeatherClear, Weight: 24},
			{Type: WeatherWindy, Weight: 16},
			{Type: WeatherHeatwave, Weight: 14},
			{Type: WeatherRain, Weight: 8},
			{Type: WeatherStorm, Weight: 4},
		}
	}

	return []weightedWeather{
		{Type: WeatherSunny, Weight: 26},
		{Type: WeatherClear, Weight: 20},
		{Type: WeatherWindy, Weight: 18},
		{Type: WeatherHeatwave, Weight: 8},
		{Type: WeatherRain, Weight: 18},
		{Type: WeatherStorm, Weight: 8},
		{Type: WeatherCloudy, Weight: 2},
	}
}

func tropicalWetWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonDry {
		return []weightedWeather{
			{Type: WeatherSunny, Weight: 18},
			{Type: WeatherClear, Weight: 12},
			{Type: WeatherCloudy, Weight: 18},
			{Type: WeatherRain, Weight: 24},
			{Type: WeatherHeavyRain, Weight: 20},
			{Type: WeatherStorm, Weight: 8},
		}
	}

	return []weightedWeather{
		{Type: WeatherCloudy, Weight: 18},
		{Type: WeatherRain, Weight: 32},
		{Type: WeatherHeavyRain, Weight: 23},
		{Type: WeatherStorm, Weight: 14},
		{Type: WeatherSunny, Weight: 6},
		{Type: WeatherClear, Weight: 4},
		{Type: WeatherWindy, Weight: 3},
	}
}

func savannaWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonWet {
		return []weightedWeather{
			{Type: WeatherSunny, Weight: 14},
			{Type: WeatherClear, Weight: 14},
			{Type: WeatherCloudy, Weight: 16},
			{Type: WeatherRain, Weight: 26},
			{Type: WeatherHeavyRain, Weight: 12},
			{Type: WeatherStorm, Weight: 10},
			{Type: WeatherWindy, Weight: 8},
		}
	}

	return []weightedWeather{
		{Type: WeatherSunny, Weight: 32},
		{Type: WeatherClear, Weight: 22},
		{Type: WeatherWindy, Weight: 18},
		{Type: WeatherHeatwave, Weight: 12},
		{Type: WeatherRain, Weight: 10},
		{Type: WeatherStorm, Weight: 6},
	}
}

func mountainWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonWinter {
		return []weightedWeather{
			{Type: WeatherClear, Weight: 18},
			{Type: WeatherCloudy, Weight: 24},
			{Type: WeatherSnow, Weight: 24},
			{Type: WeatherBlizzard, Weight: 10},
			{Type: WeatherWindy, Weight: 12},
			{Type: WeatherRain, Weight: 6},
			{Type: WeatherSunny, Weight: 6},
		}
	}

	return []weightedWeather{
		{Type: WeatherSunny, Weight: 10},
		{Type: WeatherClear, Weight: 24},
		{Type: WeatherCloudy, Weight: 24},
		{Type: WeatherRain, Weight: 18},
		{Type: WeatherStorm, Weight: 12},
		{Type: WeatherWindy, Weight: 12},
	}
}

func coastWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonWinter {
		return []weightedWeather{
			{Type: WeatherCloudy, Weight: 28},
			{Type: WeatherRain, Weight: 22},
			{Type: WeatherStorm, Weight: 14},
			{Type: WeatherWindy, Weight: 12},
			{Type: WeatherSnow, Weight: 8},
			{Type: WeatherClear, Weight: 10},
			{Type: WeatherSunny, Weight: 6},
		}
	}

	return []weightedWeather{
		{Type: WeatherCloudy, Weight: 24},
		{Type: WeatherRain, Weight: 28},
		{Type: WeatherHeavyRain, Weight: 12},
		{Type: WeatherStorm, Weight: 10},
		{Type: WeatherWindy, Weight: 10},
		{Type: WeatherClear, Weight: 10},
		{Type: WeatherSunny, Weight: 6},
	}
}

func temperateWeatherWeights(season SeasonID) []weightedWeather {
	if season == SeasonWinter {
		return []weightedWeather{
			{Type: WeatherCloudy, Weight: 26},
			{Type: WeatherRain, Weight: 20},
			{Type: WeatherSnow, Weight: 14},
			{Type: WeatherWindy, Weight: 10},
			{Type: WeatherStorm, Weight: 8},
			{Type: WeatherClear, Weight: 12},
			{Type: WeatherSunny, Weight: 10},
		}
	}
	if season == SeasonWet {
		return []weightedWeather{
			{Type: WeatherCloudy, Weight: 24},
			{Type: WeatherRain, Weight: 28},
			{Type: WeatherHeavyRain, Weight: 12},
			{Type: WeatherStorm, Weight: 10},
			{Type: WeatherWindy, Weight: 8},
			{Type: WeatherClear, Weight: 10},
			{Type: WeatherSunny, Weight: 8},
		}
	}
	if season == SeasonDry {
		return []weightedWeather{
			{Type: WeatherSunny, Weight: 24},
			{Type: WeatherClear, Weight: 22},
			{Type: WeatherCloudy, Weight: 18},
			{Type: WeatherRain, Weight: 16},
			{Type: WeatherStorm, Weight: 8},
			{Type: WeatherWindy, Weight: 12},
		}
	}

	return []weightedWeather{
		{Type: WeatherSunny, Weight: 16},
		{Type: WeatherClear, Weight: 18},
		{Type: WeatherCloudy, Weight: 24},
		{Type: WeatherRain, Weight: 22},
		{Type: WeatherHeavyRain, Weight: 8},
		{Type: WeatherStorm, Weight: 6},
		{Type: WeatherWindy, Weight: 6},
	}
}

func deterministicWeatherRoll(seed int64, biome string, season SeasonID, day int) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s:%s:%d:weather", seed, normalizeBiome(biome), season, day)))
	return int(h.Sum64() & 0x7fffffff)
}

func normalizeBiome(biome string) string {
	return strings.ToLower(strings.TrimSpace(biome))
}

func biomeIsArctic(biome string) bool {
	b := normalizeBiome(biome)
	return strings.Contains(b, "boreal") ||
		strings.Contains(b, "subarctic") ||
		strings.Contains(b, "arctic") ||
		strings.Contains(b, "tundra")
}

func biomeIsDesertOrDry(biome string) bool {
	b := normalizeBiome(biome)
	return strings.Contains(b, "desert") ||
		strings.Contains(b, "dry") ||
		strings.Contains(b, "steppe")
}

func biomeIsTropicalWet(biome string) bool {
	b := normalizeBiome(biome)
	return strings.Contains(b, "rainforest") ||
		strings.Contains(b, "jungle") ||
		strings.Contains(b, "tropical") ||
		strings.Contains(b, "wetlands") ||
		strings.Contains(b, "swamp") ||
		strings.Contains(b, "island")
}

func biomeIsHot(biome string) bool {
	return biomeIsTropicalWet(biome) || biomeIsDesertOrDry(biome) ||
		strings.Contains(normalizeBiome(biome), "savanna") ||
		strings.Contains(normalizeBiome(biome), "badlands")
}
