package game

import "math"

const (
	baseTravelMinutesPerTile = 10.0
	minTravelMinutesPerStep  = 1
)

// TravelMinutesForStep returns terrain-aware travel time for one tile.
// Tile size is 100m, so a base of 10 minutes per tile maps to ~0.6km/h before modifiers.
func TravelMinutesForStep(state *RunState, fromX, fromY, toX, toY int, player *PlayerState) int {
	if state == nil {
		return minTravelMinutesPerStep
	}
	fromCell, okFrom := state.TopologyCellAt(fromX, fromY)
	toCell, okTo := state.TopologyCellAt(toX, toY)
	if !okFrom || !okTo {
		return int(baseTravelMinutesPerTile)
	}

	terrainMultiplier := 1.0
	slopeDelta := int(toCell.Elevation) - int(fromCell.Elevation)
	switch {
	case slopeDelta > 0:
		terrainMultiplier += math.Min(1.0, float64(slopeDelta)*0.05)
	case slopeDelta < 0:
		terrainMultiplier += math.Min(0.30, float64(-slopeDelta)*0.015)
	}

	// Roughness is generated as 1..9 and represents movement difficulty.
	if toCell.Roughness > 1 {
		terrainMultiplier += float64(toCell.Roughness-1) * 0.08
	}
	switch toCell.Biome {
	case TopoBiomeWetland, TopoBiomeSwamp, TopoBiomeJungle:
		terrainMultiplier += 0.30
	case TopoBiomeMountain:
		terrainMultiplier += 0.24
	case TopoBiomeTundra, TopoBiomeBoreal:
		terrainMultiplier += 0.15
	case TopoBiomeDesert:
		terrainMultiplier += 0.10
	case TopoBiomeGrassland:
		terrainMultiplier += 0.04
	}
	if toCell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake) != 0 {
		terrainMultiplier += 0.35
	}
	switch state.Weather.Type {
	case WeatherStorm, WeatherBlizzard:
		terrainMultiplier += 0.45
	case WeatherHeavyRain, WeatherSnow:
		terrainMultiplier += 0.25
	case WeatherWindy:
		terrainMultiplier += 0.12
	}

	paceMultiplier := 1.0
	if player != nil {
		paceMultiplier -= float64(clamp(player.Agility, -3, 3)) * 0.04
		paceMultiplier -= float64(clamp(player.Endurance, -3, 3)) * 0.03
		paceMultiplier -= float64(clamp(player.Navigation, 0, 100)) / 750.0
		paceMultiplier -= float64(clamp(player.Gathering, 0, 100)) / 900.0
		if limit := state.playerCarryLimitKg(player); limit > 0 {
			ratio := inventoryWeightKg(player.PersonalItems) / limit
			if ratio > 0.7 {
				paceMultiplier += (ratio - 0.7) * 0.55
			}
		}
	}

	terrainMultiplier = clampFloat(terrainMultiplier, 0.55, 3.4)
	paceMultiplier = clampFloat(paceMultiplier, 0.55, 1.75)

	minutes := int(math.Round(baseTravelMinutesPerTile * terrainMultiplier * paceMultiplier))
	if minutes < minTravelMinutesPerStep {
		minutes = minTravelMinutesPerStep
	}
	return minutes
}
