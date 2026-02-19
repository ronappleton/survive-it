package game

import (
	"fmt"
	"math"
	"strings"
)

type TravelState struct {
	Direction     string  `json:"direction,omitempty"`
	TotalKm       float64 `json:"total_km"`
	LastStepKm    float64 `json:"last_step_km"`
	LastStepHours float64 `json:"last_step_hours"`
	LastDay       int     `json:"last_day"`
	PosX          int     `json:"pos_x"`
	PosY          int     `json:"pos_y"`
}

type TravelResult struct {
	PlayerID        int
	Direction       string
	DistanceKm      float64
	HoursSpent      float64
	WatercraftUsed  string
	TravelSpeedKmph float64
	EnergyCost      int
	HydrationCost   int
	MoraleDelta     int
	EncounterLogs   []string
}

func normalizeDirection(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "n", "north":
		return "north"
	case "s", "south":
		return "south"
	case "e", "east":
		return "east"
	case "w", "west":
		return "west"
	default:
		return ""
	}
}

func directionDelta(direction string) (int, int) {
	switch normalizeDirection(direction) {
	case "north":
		return 0, -1
	case "south":
		return 0, 1
	case "east":
		return 1, 0
	case "west":
		return -1, 0
	default:
		return 0, 0
	}
}

func (s *RunState) hasCraftedItem(id string) bool {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return false
	}
	return slicesContainsString(s.CraftedItems, id)
}

func slicesContainsString(items []string, target string) bool {
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item)) == target {
			return true
		}
	}
	return false
}

func (s *RunState) availableWatercraft() (string, float64) {
	// Best first.
	if s.hasCraftedItem("dugout_canoe") {
		return "dugout_canoe", 1.45
	}
	if s.hasCraftedItem("reed_coracle") {
		return "reed_coracle", 1.3
	}
	if s.hasCraftedItem("brush_raft") {
		return "brush_raft", 1.2
	}
	return "", 1.0
}

func (s *RunState) canUseWatercraftInBiome() bool {
	if s == nil {
		return false
	}
	n := normalizeBiome(s.Scenario.Biome + " " + s.Scenario.Name)
	return strings.Contains(n, "river") ||
		strings.Contains(n, "lake") ||
		strings.Contains(n, "delta") ||
		strings.Contains(n, "coast") ||
		strings.Contains(n, "island") ||
		strings.Contains(n, "wetland") ||
		strings.Contains(n, "swamp")
}

func (s *RunState) TravelMove(playerID int, direction string, requestedKm float64) (TravelResult, error) {
	if s == nil {
		return TravelResult{}, fmt.Errorf("run state is nil")
	}
	s.EnsureTopology()
	player, ok := s.playerByID(playerID)
	if !ok {
		return TravelResult{}, fmt.Errorf("player %d not found", playerID)
	}
	direction = normalizeDirection(direction)
	if direction == "" {
		return TravelResult{}, fmt.Errorf("direction must be north/south/east/west")
	}
	if requestedKm <= 0 {
		requestedKm = 2.0
	}
	requestedKm = clampFloat(requestedKm, 0.2, 24.0)

	speed := 2.4 + float64(clamp(player.Agility, -3, 3))*0.24 + float64(clamp(player.Endurance, -3, 3))*0.18
	speed += float64(player.Gathering) / 220.0
	if speed < 1.3 {
		speed = 1.3
	}

	watercraftID := ""
	watercraftBoost := 1.0
	if s.canUseWatercraftInBiome() {
		watercraftID, watercraftBoost = s.availableWatercraft()
		speed *= watercraftBoost
	}
	if slicesContainsKit(player.Kit, KitCompass) || slicesContainsKit(s.Config.IssuedKit, KitCompass) {
		speed *= 1.05
	}
	if slicesContainsKit(player.Kit, KitMap) || slicesContainsKit(s.Config.IssuedKit, KitMap) {
		speed *= 1.04
	}

	dx, dy := directionDelta(direction)
	if dx == 0 && dy == 0 {
		return TravelResult{}, fmt.Errorf("direction must be north/south/east/west")
	}
	steps := max(1, int(math.Round(requestedKm/0.1)))
	posX, posY := s.CurrentMapPosition()
	movedSteps := 0
	terrainEffort := 0.0
	encounterLogs := make([]string, 0, 3)
	for step := 0; step < steps; step++ {
		nextX := posX + dx
		nextY := posY + dy
		if s.Topology.Width > 0 {
			if nextX < 0 {
				nextX = s.Topology.Width - 1
			} else if nextX >= s.Topology.Width {
				nextX = 0
			}
		}
		if s.Topology.Height > 0 {
			if nextY < 0 {
				nextY = s.Topology.Height - 1
			} else if nextY >= s.Topology.Height {
				nextY = 0
			}
		}
		fromCell, okFrom := s.TopologyCellAt(posX, posY)
		toCell, okTo := s.TopologyCellAt(nextX, nextY)
		if !okFrom || !okTo || (nextX == posX && nextY == posY) {
			break
		}
		slope := int(toCell.Elevation) - int(fromCell.Elevation)
		terrainMul := 1.0 + float64(toCell.Roughness)*0.09
		if slope > 0 {
			terrainMul += float64(slope) * 0.03
		} else if slope < -6 {
			terrainMul += 0.04
		}
		if toCell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake) != 0 {
			if watercraftID == "" {
				terrainMul += 0.8
			} else {
				terrainMul += 0.18
			}
		}
		terrainMul = clampFloat(terrainMul, 0.65, 3.2)
		terrainEffort += terrainMul
		posX, posY = nextX, nextY
		movedSteps++
		s.applyCellStateAction(posX, posY, "move")
		s.RevealFog(posX, posY, 1)
		if len(encounterLogs) < 2 {
			event, ok := s.RollWildlifeEncounter(playerID, posX, posY, "move", step)
			if ok {
				encounterLogs = append(encounterLogs, event.Message)
				player.Energy = clamp(player.Energy+event.EnergyDelta, 0, 100)
				player.Hydration = clamp(player.Hydration+event.HydrationDelta, 0, 100)
				player.Morale = clamp(player.Morale+event.MoraleDelta, 0, 100)
			}
		}
	}
	if movedSteps == 0 {
		return TravelResult{}, fmt.Errorf("cannot move further in that direction")
	}
	distance := math.Round(float64(movedSteps)*0.1*10) / 10
	avgEffort := terrainEffort / float64(movedSteps)
	effectiveKm := distance * avgEffort
	hours := clampFloat(effectiveKm/speed, 0.1, 16)
	energyCost := clamp(int(math.Ceil(hours*4.0)), 1, 24)
	hydrationCost := clamp(int(math.Ceil(hours*3.1)), 1, 24)
	moraleDelta := 0
	if watercraftID != "" {
		energyCost = max(1, energyCost-1)
		hydrationCost = max(1, hydrationCost-1)
		moraleDelta = 1
	}

	player.Energy = clamp(player.Energy-energyCost, 0, 100)
	player.Hydration = clamp(player.Hydration-hydrationCost, 0, 100)
	player.Morale = clamp(player.Morale+moraleDelta, 0, 100)
	applySkillEffort(&player.Gathering, int(math.Round(hours*12)), true)
	applySkillEffort(&player.Navigation, int(math.Round(hours*14)), true)
	refreshEffectBars(player)

	s.Travel.PosX = posX
	s.Travel.PosY = posY
	s.Travel.Direction = direction
	s.Travel.TotalKm += distance
	s.Travel.LastStepKm = distance
	s.Travel.LastStepHours = hours
	s.Travel.LastDay = s.Day
	_ = s.AdvanceActionClock(hours)

	return TravelResult{
		PlayerID:        playerID,
		Direction:       direction,
		DistanceKm:      distance,
		HoursSpent:      hours,
		WatercraftUsed:  watercraftID,
		TravelSpeedKmph: speed,
		EnergyCost:      energyCost,
		HydrationCost:   hydrationCost,
		MoraleDelta:     moraleDelta,
		EncounterLogs:   encounterLogs,
	}, nil
}

func slicesContainsKit(items []KitItem, target KitItem) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
