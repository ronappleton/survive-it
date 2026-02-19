package game

import (
	"strconv"
	"strings"
)

// Discovery summary:
// - Scenario.Biome string and per-cell topology biome were previously used independently across systems.
// - Weather, look text, insects, and wildlife selection had no shared climate guardrail.
// - This file provides a lightweight constraint layer so topology/weather/encounters/descriptions agree.

type ClimateProfile struct {
	Name               string
	AllowedBiomes      []uint8
	BaseTempC          int
	TempVarianceC      int
	FrozenWaterBelowC  int
	DisallowTags       []string
	SeasonRules        map[SeasonID]SeasonRule
	DefaultInsectMinC  int
	DefaultInsectQuiet string
}

type SeasonRule struct {
	TempBiasC            int
	AllowedFloraTags     []string
	AllowedFaunaTags     []string
	DisallowedFloraTags  []string
	DisallowedFaunaTags  []string
	InsectActivity       bool
	InsectMinTempC       int
	InsectQuietMessage   string
	DisallowWeatherTypes []WeatherType
}

func (s *RunState) ActiveClimateProfile() *ClimateProfile {
	if s == nil {
		return nil
	}
	return s.Scenario.Climate
}

func (s *RunState) FrozenWaterBelowC() int {
	climate := s.ActiveClimateProfile()
	if climate == nil {
		return 0
	}
	return climate.FrozenWaterBelowC
}

func (s *RunState) IsWaterFrozenAtTemp(tempC int) bool {
	return tempC <= s.FrozenWaterBelowC()
}

func (s *RunState) IsWaterCurrentlyFrozen() bool {
	if s == nil {
		return false
	}
	return s.IsWaterFrozenAtTemp(s.Weather.TemperatureC)
}

func (s *RunState) CoherenceDebugLine() string {
	if s == nil {
		return "Biome: unknown | Temp: ? | Season: ? | Weather: ? | Frozen: ?"
	}
	season, ok := s.CurrentSeason()
	if !ok {
		season = SeasonAutumn
	}
	x, y := s.CurrentMapPosition()
	cell, ok := s.TopologyCellAt(x, y)
	if !ok {
		return "Biome: unknown | Temp: ? | Season: ? | Weather: ? | Frozen: ?"
	}
	biomeTags := strings.Join(topoBiomeTags(cell.Biome), ",")
	return "Biome: " + topoBiomeLabel(cell.Biome) +
		" [" + biomeTags + "] | Temp: " + itoa(s.Weather.TemperatureC) + "C | Season: " + string(season) +
		" | Weather: " + WeatherLabel(s.Weather.Type) + " | Frozen water: " + boolLabel(s.IsWaterCurrentlyFrozen())
}

func (s *RunState) CurrentBiomeQuery() string {
	if s == nil {
		return ""
	}
	x, y := s.CurrentMapPosition()
	return s.BiomeQueryAt(x, y)
}

func (s *RunState) BiomeQueryAt(x, y int) string {
	if s == nil {
		return ""
	}
	cell, ok := s.TopologyCellAt(x, y)
	if !ok {
		return s.Scenario.Biome
	}
	return topoBiomeQuery(cell.Biome)
}

func climateSeasonRule(climate *ClimateProfile, season SeasonID) (SeasonRule, bool) {
	if climate == nil {
		return SeasonRule{}, false
	}
	rule, ok := climate.SeasonRules[season]
	if !ok {
		return SeasonRule{}, false
	}
	return rule, true
}

func climateInsectMinTempC(climate *ClimateProfile, season SeasonID) int {
	const defaultMin = 5
	if climate == nil {
		return defaultMin
	}
	rule, ok := climateSeasonRule(climate, season)
	if ok && rule.InsectMinTempC != 0 {
		return rule.InsectMinTempC
	}
	if climate.DefaultInsectMinC != 0 {
		return climate.DefaultInsectMinC
	}
	return defaultMin
}

func insectActivityAllowed(climate *ClimateProfile, season SeasonID, tempC int, biome uint8) bool {
	minTemp := climateInsectMinTempC(climate, season)
	if tempC < minTemp {
		return false
	}
	if climate == nil {
		return true
	}
	if rule, ok := climateSeasonRule(climate, season); ok && !rule.InsectActivity {
		return false
	}
	if (biome == TopoBiomeTundra || biome == TopoBiomeBoreal) && tempC < minTemp+2 {
		return false
	}
	return true
}

func quietInsectMessage(climate *ClimateProfile, season SeasonID) string {
	if climate == nil {
		return "It's quiet. No insect activity in this cold."
	}
	if rule, ok := climateSeasonRule(climate, season); ok {
		if strings.TrimSpace(rule.InsectQuietMessage) != "" {
			return strings.TrimSpace(rule.InsectQuietMessage)
		}
	}
	if strings.TrimSpace(climate.DefaultInsectQuiet) != "" {
		return strings.TrimSpace(climate.DefaultInsectQuiet)
	}
	return "It's quiet. No insect activity in this cold."
}

func climateAllowsTopoBiome(climate *ClimateProfile, biome uint8) bool {
	if climate == nil {
		return true
	}
	tags := topoBiomeTags(biome)
	if hasAnyTag(tags, climate.DisallowTags) {
		return false
	}
	if len(climate.AllowedBiomes) == 0 {
		return true
	}
	for _, allowed := range climate.AllowedBiomes {
		if allowed == biome {
			return true
		}
	}
	return false
}

func constrainWeatherForClimate(seed int64, day int, season SeasonID, weather WeatherType, climate *ClimateProfile) WeatherType {
	if climate == nil {
		return weather
	}
	rule, hasRule := climateSeasonRule(climate, season)
	if hasRule && weatherTypeDisallowed(weather, rule.DisallowWeatherTypes) {
		return fallbackWeatherForClimate(seed, day, season, weather, climate)
	}

	baseTemp := climate.BaseTempC + rule.TempBiasC
	coldSeason := season == SeasonWinter || baseTemp <= 2
	if coldSeason {
		switch weather {
		case WeatherHeatwave:
			return WeatherSunny
		case WeatherRain, WeatherHeavyRain:
			return WeatherSnow
		case WeatherStorm:
			if baseTemp <= -6 {
				return WeatherBlizzard
			}
			return WeatherSnow
		}
	}
	return weather
}

func fallbackWeatherForClimate(seed int64, day int, season SeasonID, from WeatherType, climate *ClimateProfile) WeatherType {
	roll := deterministicWeatherRoll(seed, climate.Name, season, day)
	candidates := []WeatherType{WeatherCloudy, WeatherSnow, WeatherWindy, WeatherClear}
	rule, _ := climateSeasonRule(climate, season)
	baseTemp := climate.BaseTempC + rule.TempBiasC
	if baseTemp > 6 {
		candidates = []WeatherType{WeatherCloudy, WeatherClear, WeatherSunny, WeatherWindy}
	}
	if season == SeasonWinter || baseTemp <= 2 {
		candidates = []WeatherType{WeatherSnow, WeatherCloudy, WeatherWindy, WeatherClear}
	}
	if from == WeatherHeatwave && len(candidates) > 0 {
		return candidates[roll%len(candidates)]
	}
	return candidates[roll%len(candidates)]
}

func weatherTypeDisallowed(weather WeatherType, disallowed []WeatherType) bool {
	for _, blocked := range disallowed {
		if blocked == weather {
			return true
		}
	}
	return false
}

func TemperatureForDayWithClimate(seed int64, day int, season SeasonID, weather WeatherType, climate *ClimateProfile) int {
	if climate == nil {
		return 0
	}
	rule, _ := climateSeasonRule(climate, season)
	variance := climate.TempVarianceC
	if variance < 1 {
		variance = 1
	}
	roll := deterministicWeatherRoll(seed, climate.Name, season, day+37)
	offset := (roll % (variance*2 + 1)) - variance
	temp := climate.BaseTempC + rule.TempBiasC + offset
	temp += weatherTemperatureDelta(weather) / 2
	minC := climate.BaseTempC + rule.TempBiasC - variance - 6
	maxC := climate.BaseTempC + rule.TempBiasC + variance + 6
	return clamp(temp, minC, maxC)
}

func constrainTopoBiomeForClimate(climate *ClimateProfile, biome uint8, moisture, temp uint8) uint8 {
	if climateAllowsTopoBiome(climate, biome) {
		return biome
	}
	candidates := []uint8{
		TopoBiomeTundra,
		TopoBiomeBoreal,
		TopoBiomeMountain,
		TopoBiomeWetland,
		TopoBiomeSwamp,
		TopoBiomeForest,
		TopoBiomeGrassland,
	}
	if int(temp) > 145 {
		candidates = []uint8{
			TopoBiomeGrassland,
			TopoBiomeForest,
			TopoBiomeWetland,
			TopoBiomeMountain,
			TopoBiomeBoreal,
			TopoBiomeTundra,
		}
	}
	if int(moisture) > 178 {
		candidates = []uint8{
			TopoBiomeWetland,
			TopoBiomeSwamp,
			TopoBiomeBoreal,
			TopoBiomeForest,
			TopoBiomeTundra,
			TopoBiomeMountain,
		}
	}
	for _, candidate := range candidates {
		if climateAllowsTopoBiome(climate, candidate) {
			return candidate
		}
	}
	if climate != nil && len(climate.AllowedBiomes) > 0 {
		return climate.AllowedBiomes[0]
	}
	return biome
}

func climateAllowsFlora(climate *ClimateProfile, season SeasonID, tags []string) bool {
	if climate == nil {
		return true
	}
	if hasAnyTag(tags, climate.DisallowTags) {
		return false
	}
	rule, hasRule := climateSeasonRule(climate, season)
	if hasRule {
		if hasAnyTag(tags, rule.DisallowedFloraTags) {
			return false
		}
		if len(rule.AllowedFloraTags) > 0 && !hasAnyTag(tags, rule.AllowedFloraTags) {
			return false
		}
	}
	return true
}

func climateAllowsFauna(climate *ClimateProfile, season SeasonID, tags []string) bool {
	if climate == nil {
		return true
	}
	if hasAnyTag(tags, climate.DisallowTags) {
		return false
	}
	rule, hasRule := climateSeasonRule(climate, season)
	if hasRule {
		if hasAnyTag(tags, rule.DisallowedFaunaTags) {
			return false
		}
		if len(rule.AllowedFaunaTags) > 0 && !hasAnyTag(tags, rule.AllowedFaunaTags) {
			return false
		}
	}
	return true
}

func filterPlantsForClimate(plants []PlantSpec, climate *ClimateProfile, season SeasonID, tempC int) []PlantSpec {
	if len(plants) == 0 || climate == nil {
		return plants
	}
	out := make([]PlantSpec, 0, len(plants))
	for _, plant := range plants {
		tags := plantClimateTags(plant)
		if tempC <= climate.FrozenWaterBelowC && !hasAnyTag(tags, []string{"winter_ok", "all_season", "arctic", "subarctic", "boreal", "tundra"}) {
			continue
		}
		if !climateAllowsFlora(climate, season, tags) {
			continue
		}
		out = append(out, plant)
	}
	return out
}

func filterInsectsForClimate(names []string, climate *ClimateProfile, season SeasonID, tempC int, biome uint8) []string {
	if len(names) == 0 {
		return nil
	}
	if !insectActivityAllowed(climate, season, tempC, biome) {
		return nil
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		tags := insectNameClimateTags(name)
		if !climateAllowsFauna(climate, season, tags) {
			continue
		}
		if tempC < climateInsectMinTempC(climate, season) && hasAnyTag(tags, []string{"warm_season"}) {
			continue
		}
		out = append(out, name)
	}
	return out
}

func filterEncounterSpeciesForClimate(species []encounterSpecies, channel string, climate *ClimateProfile, season SeasonID, tempC int, biome uint8) []encounterSpecies {
	if len(species) == 0 {
		return nil
	}
	if channel == "insect" && !insectActivityAllowed(climate, season, tempC, biome) {
		return nil
	}
	out := make([]encounterSpecies, 0, len(species))
	for _, sp := range species {
		tags := encounterSpeciesClimateTags(sp.AnimalID, channel)
		if !climateAllowsFauna(climate, season, tags) {
			continue
		}
		if channel == "insect" && tempC < climateInsectMinTempC(climate, season) && hasAnyTag(tags, []string{"warm_season"}) {
			continue
		}
		out = append(out, sp)
	}
	return out
}

func plantClimateTags(plant PlantSpec) []string {
	tags := make([]string, 0, len(plant.BiomeTags)+6)
	tags = append(tags, plant.BiomeTags...)
	tags = append(tags, string(plant.Category))
	if len(plant.SeasonTags) == 0 {
		tags = append(tags, "all_season")
	} else {
		winterOK := false
		for _, season := range plant.SeasonTags {
			tags = append(tags, string(season))
			if season == SeasonWinter {
				winterOK = true
			}
		}
		if winterOK {
			tags = append(tags, "winter_ok")
		} else {
			tags = append(tags, "warm_season")
		}
	}
	if plant.Medicinal > 0 {
		tags = append(tags, "medicinal")
	}
	if plant.Toxicity > 0 {
		tags = append(tags, "toxic")
	}
	return normalizeTags(tags)
}

func insectNameClimateTags(name string) []string {
	norm := normalizeTag(name)
	tags := []string{"insect"}
	switch {
	case strings.Contains(norm, "mosquito"), strings.Contains(norm, "midge"), strings.Contains(norm, "black_fly"),
		strings.Contains(norm, "horsefly"), strings.Contains(norm, "bee"), strings.Contains(norm, "wasp"),
		strings.Contains(norm, "fire_ant"), strings.Contains(norm, "leech"):
		tags = append(tags, "warm_season")
	case strings.Contains(norm, "tick"):
		tags = append(tags, "cool_tolerant")
	}
	return normalizeTags(tags)
}

func encounterSpeciesClimateTags(animalID, channel string) []string {
	normID := normalizeTag(animalID)
	tags := []string{channel}
	if animal, ok := animalSpecByID(normID); ok {
		tags = append(tags, animal.BiomeTags...)
		switch animal.Domain {
		case AnimalDomainLand:
			tags = append(tags, "land")
		case AnimalDomainWater:
			tags = append(tags, "water")
		case AnimalDomainAir:
			tags = append(tags, "air")
		}
	}
	switch normID {
	case "quail", "pheasant", "wild_turkey", "dove", "partridge":
		tags = append(tags, "temperate_warm", "warm_season")
	case "mosquito", "midge", "black_fly", "horsefly", "bee", "wasp", "fire_ant", "leech":
		tags = append(tags, "warm_season", "insect")
	case "tick":
		tags = append(tags, "insect", "cool_tolerant")
	case "caribou", "moose", "char", "grayling", "trout", "wolf", "arctic_fox":
		tags = append(tags, "cold", "winter_ok")
	}
	if strings.Contains(normID, "bear") || strings.Contains(normID, "wolf") || strings.Contains(normID, "cougar") || strings.Contains(normID, "jaguar") {
		tags = append(tags, "predator")
	}
	return normalizeTags(tags)
}

func animalSpecByID(id string) (AnimalSpec, bool) {
	for _, animal := range AnimalCatalog() {
		if normalizeTag(animal.ID) == id {
			return animal, true
		}
	}
	return AnimalSpec{}, false
}

func topoBiomeQuery(biome uint8) string {
	switch biome {
	case TopoBiomeForest:
		return "forest temperate"
	case TopoBiomeGrassland:
		return "grassland"
	case TopoBiomeJungle:
		return "jungle tropical"
	case TopoBiomeWetland:
		return "wetlands"
	case TopoBiomeSwamp:
		return "swamp wetlands"
	case TopoBiomeDesert:
		return "desert dry"
	case TopoBiomeMountain:
		return "mountain alpine"
	case TopoBiomeTundra:
		return "tundra arctic subarctic"
	case TopoBiomeBoreal:
		return "boreal subarctic forest"
	default:
		return "forest"
	}
}

func topoBiomeTags(biome uint8) []string {
	switch biome {
	case TopoBiomeForest:
		return []string{"forest", "temperate", "wooded"}
	case TopoBiomeGrassland:
		return []string{"grassland", "open"}
	case TopoBiomeJungle:
		return []string{"jungle", "tropical", "warm", "humid"}
	case TopoBiomeWetland:
		return []string{"wetlands", "water", "humid"}
	case TopoBiomeSwamp:
		return []string{"swamp", "wetlands", "water", "humid"}
	case TopoBiomeDesert:
		return []string{"desert", "arid", "dry", "hot"}
	case TopoBiomeMountain:
		return []string{"mountain", "alpine", "cold"}
	case TopoBiomeTundra:
		return []string{"tundra", "arctic", "subarctic", "cold", "winter_ok"}
	case TopoBiomeBoreal:
		return []string{"boreal", "subarctic", "cold", "forest", "winter_ok"}
	default:
		return []string{"mixed"}
	}
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := normalizeTag(raw)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	return out
}

func normalizeTag(raw string) string {
	tag := strings.ToLower(strings.TrimSpace(raw))
	tag = strings.ReplaceAll(tag, "-", "_")
	tag = strings.ReplaceAll(tag, " ", "_")
	return tag
}

func hasAnyTag(tags []string, candidates []string) bool {
	if len(tags) == 0 || len(candidates) == 0 {
		return false
	}
	set := map[string]bool{}
	for _, tag := range tags {
		set[normalizeTag(tag)] = true
	}
	for _, candidate := range candidates {
		if set[normalizeTag(candidate)] {
			return true
		}
	}
	return false
}

func boolLabel(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
