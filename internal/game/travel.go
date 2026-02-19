package game

import (
	"fmt"
	"math"
	"strconv"
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
	RequestedKm     float64
	RequestedSteps  int
	StepsMoved      int
	DistanceKm      float64
	HoursSpent      float64
	WatercraftUsed  string
	TravelSpeedKmph float64
	EnergyCost      int
	HydrationCost   int
	MoraleDelta     int
	StartBlock      TimeBlock
	EndBlock        TimeBlock
	BlocksCrossed   int
	StopReason      string
	EncounterLogs   []string
}

const travelTileKm = 0.1

func ParseTravelDistanceInput(raw string) (float64, bool) {
	tokens := strings.Fields(strings.ToLower(strings.TrimSpace(raw)))
	return parseTravelDistanceTokens(tokens)
}

func parseTravelDistanceTokens(tokens []string) (float64, bool) {
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(strings.ToLower(token))
		if token != "" {
			filtered = append(filtered, token)
		}
	}
	if len(filtered) == 0 {
		return 0, false
	}
	if len(filtered) == 1 {
		return parseTravelDistanceToken(filtered[0])
	}
	if len(filtered) == 2 {
		if v, err := strconv.ParseFloat(filtered[0], 64); err == nil {
			if unit := normaliseTravelUnit(filtered[1]); unit != "" {
				return convertTravelDistance(v, unit), v > 0
			}
		}
		if km, ok := parseTravelDistanceToken(filtered[0] + filtered[1]); ok {
			return km, true
		}
	}
	return 0, false
}

func parseTravelDistanceToken(token string) (float64, bool) {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return 0, false
	}
	if v, err := strconv.ParseFloat(token, 64); err == nil {
		return v, v > 0
	}
	units := []string{
		"kilometers", "kilometer", "kms", "km",
		"meters", "meter", "metres", "metre", "m",
		"tiles", "tile", "steps", "step",
	}
	for _, suffix := range units {
		if !strings.HasSuffix(token, suffix) {
			continue
		}
		raw := strings.TrimSpace(strings.TrimSuffix(token, suffix))
		if raw == "" {
			return 0, false
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v <= 0 {
			return 0, false
		}
		return convertTravelDistance(v, normaliseTravelUnit(suffix)), true
	}
	return 0, false
}

func normaliseTravelUnit(unit string) string {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "km", "kms", "kilometer", "kilometers":
		return "km"
	case "m", "meter", "meters", "metre", "metres":
		return "m"
	case "tile", "tiles", "step", "steps":
		return "tile"
	default:
		return ""
	}
}

func convertTravelDistance(value float64, unit string) float64 {
	switch normaliseTravelUnit(unit) {
	case "m":
		return value / 1000.0
	case "tile":
		return value * travelTileKm
	default:
		return value
	}
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
	if player.Energy <= 1 || player.Hydration <= 1 {
		return TravelResult{}, fmt.Errorf("too exhausted to travel")
	}
	if requestedKm <= 0 {
		requestedKm = 2.0
	}
	requestedKm = clampFloat(requestedKm, 0.1, 24.0)

	watercraftID := ""
	watercraftBoost := 1.0
	if s.canUseWatercraftInBiome() {
		watercraftID, watercraftBoost = s.availableWatercraft()
	}

	dx, dy := directionDelta(direction)
	if dx == 0 && dy == 0 {
		return TravelResult{}, fmt.Errorf("direction must be north/south/east/west")
	}
	steps := max(1, int(math.Round(requestedKm/travelTileKm)))
	posX, posY := s.CurrentMapPosition()
	movedSteps := 0
	totalMinutes := 0
	energyCost := 0
	hydrationCost := 0
	moraleDelta := 0
	startBlock := s.CurrentTimeBlock()
	endBlock := startBlock
	blocksCrossed := 0
	encounterLogs := make([]string, 0, 3)
	stopReason := ""
	for step := 0; step < steps; step++ {
		if player.Energy <= 1 || player.Hydration <= 1 {
			stopReason = "Too exhausted"
			break
		}
		nextX := posX + dx
		nextY := posY + dy
		if s.Topology.Width > 0 && (nextX < 0 || nextX >= s.Topology.Width) {
			stopReason = "Reached boundary"
			break
		}
		if s.Topology.Height > 0 && (nextY < 0 || nextY >= s.Topology.Height) {
			stopReason = "Reached boundary"
			break
		}
		fromCell, okFrom := s.TopologyCellAt(posX, posY)
		toCell, okTo := s.TopologyCellAt(nextX, nextY)
		if !okFrom || !okTo || (nextX == posX && nextY == posY) {
			break
		}
		_ = fromCell

		stepMinutes := TravelMinutesForStep(s, posX, posY, nextX, nextY, player)
		if toCell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake) != 0 {
			if watercraftID != "" {
				stepMinutes = max(1, int(math.Round(float64(stepMinutes)*0.55*watercraftBoost)))
			} else {
				stepMinutes = max(1, int(math.Round(float64(stepMinutes)*1.35)))
			}
		}
		if slicesContainsKit(player.Kit, KitCompass) || slicesContainsKit(s.Config.IssuedKit, KitCompass) {
			stepMinutes = max(1, int(math.Round(float64(stepMinutes)*0.96)))
		}
		if slicesContainsKit(player.Kit, KitMap) || slicesContainsKit(s.Config.IssuedKit, KitMap) {
			stepMinutes = max(1, int(math.Round(float64(stepMinutes)*0.97)))
		}

		prevBlock := s.CurrentTimeBlock()
		s.AdvanceMinutes(stepMinutes)
		currBlock := s.CurrentTimeBlock()
		if currBlock != prevBlock {
			blocksCrossed++
		}
		endBlock = currBlock

		stepHours := float64(stepMinutes) / 60.0
		stepEnergy := max(1, int(math.Ceil(stepHours*4.0)))
		stepHydration := max(1, int(math.Ceil(stepHours*3.1)))
		if watercraftID != "" {
			stepEnergy = max(1, stepEnergy-1)
			stepHydration = max(1, stepHydration-1)
			moraleDelta++
		}
		player.Energy = clamp(player.Energy-stepEnergy, 0, 100)
		player.Hydration = clamp(player.Hydration-stepHydration, 0, 100)
		energyCost += stepEnergy
		hydrationCost += stepHydration
		totalMinutes += stepMinutes

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
		if player.Energy <= 1 || player.Hydration <= 1 {
			stopReason = "Too exhausted"
			break
		}
	}
	if movedSteps == 0 {
		if stopReason != "" {
			return TravelResult{
				PlayerID:        playerID,
				Direction:       direction,
				RequestedKm:     requestedKm,
				RequestedSteps:  steps,
				StepsMoved:      0,
				DistanceKm:      0,
				HoursSpent:      0,
				WatercraftUsed:  watercraftID,
				TravelSpeedKmph: 0,
				EnergyCost:      0,
				HydrationCost:   0,
				MoraleDelta:     0,
				StartBlock:      startBlock,
				EndBlock:        startBlock,
				BlocksCrossed:   0,
				StopReason:      stopReason,
				EncounterLogs:   nil,
			}, nil
		}
		return TravelResult{}, fmt.Errorf("cannot move further in that direction")
	}
	distance := math.Round(float64(movedSteps)*travelTileKm*10) / 10
	hours := float64(totalMinutes) / 60.0
	speed := 0.0
	if hours > 0 {
		speed = distance / hours
	}
	if watercraftID != "" && movedSteps > 0 {
		moraleDelta += 1
	}
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
		RequestedKm:     requestedKm,
		RequestedSteps:  steps,
		StepsMoved:      movedSteps,
		DistanceKm:      distance,
		HoursSpent:      hours,
		WatercraftUsed:  watercraftID,
		TravelSpeedKmph: speed,
		EnergyCost:      energyCost,
		HydrationCost:   hydrationCost,
		MoraleDelta:     moraleDelta,
		StartBlock:      startBlock,
		EndBlock:        endBlock,
		BlocksCrossed:   blocksCrossed,
		StopReason:      stopReason,
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
