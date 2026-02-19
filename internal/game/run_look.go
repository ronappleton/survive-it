package game

import (
	"fmt"
	"strings"
)

func (s *RunState) executeLookCommand(command string, fields []string) RunCommandResult {
	playerID, relative, detailed, subject := parseLookRequest(fields, command == "inspect" || command == "examine")
	if s == nil {
		return RunCommandResult{Handled: true, Message: "Look unavailable."}
	}
	if _, ok := s.playerByID(playerID); !ok {
		return RunCommandResult{Handled: true, Message: fmt.Sprintf("Player %d not found.", playerID)}
	}
	msg := s.describeDirectionalView(playerID, relative, detailed, subject)
	return RunCommandResult{Handled: true, Message: msg}
}

func parseLookRequest(fields []string, defaultDetailed bool) (playerID int, relative string, detailed bool, subject string) {
	playerID = 1
	relative = "front"
	detailed = defaultDetailed
	subjectParts := make([]string, 0, len(fields))
	for _, raw := range fields {
		token := strings.ToLower(strings.TrimSpace(raw))
		if token == "" {
			continue
		}
		if parsed := parsePlayerToken(token); parsed > 0 {
			playerID = parsed
			continue
		}
		if rel, ok := normalizeLookRelative(token); ok {
			relative = rel
			continue
		}
		switch token {
		case "closer", "close", "closely", "detail", "detailed", "inspect", "examine":
			detailed = true
			continue
		case "at", "to", "the", "a", "an", "toward", "towards", "in", "on", "my":
			continue
		}
		subjectParts = append(subjectParts, token)
	}
	subject = strings.TrimSpace(strings.Join(subjectParts, " "))
	return playerID, relative, detailed, subject
}

func (s *RunState) describeDirectionalView(playerID int, relative string, detailed bool, subject string) string {
	s.EnsureTopology()
	dir := s.absoluteLookDirection(relative)
	x, y := s.CurrentMapPosition()
	tx, ty, inBounds := s.stepInDirection(x, y, dir)
	cell, ok := s.TopologyCellAt(x, y)
	if inBounds {
		if ahead, okAhead := s.TopologyCellAt(tx, ty); okAhead {
			cell = ahead
			ok = true
		}
	}
	if !ok {
		return "You scan the area but cannot make out details from here."
	}
	if detailed {
		return s.describeLookCloser(playerID, relative, dir, subject, cell, tx, ty, inBounds)
	}
	return s.describeLookOverview(relative, dir, cell, tx, ty, inBounds)
}

func (s *RunState) describeLookOverview(relative, dir string, cell TopoCell, tx, ty int, inBounds bool) string {
	posLabel := lookRelativeLabel(relative)
	if !inBounds {
		return fmt.Sprintf("Looking %s (%s), you reach the map boundary. Beyond it the terrain drops out of charted range.", posLabel, dir)
	}

	biome := topoBiomeLabel(cell.Biome)
	trees := TreesForBiome(s.Scenario.Biome)
	plants := s.lookPlantsForSeason()
	insects := InsectsForBiome(s.Scenario.Biome)

	treeSnippet := "a few scattered trees"
	if len(trees) >= 2 {
		t1 := trees[s.lookIndex("tree-ahead-a", len(trees), tx, ty)]
		t2 := trees[s.lookIndex("tree-ahead-b", len(trees), tx, ty)]
		if t2.ID == t1.ID && len(trees) > 1 {
			t2 = trees[(s.lookIndex("tree-ahead-b", len(trees), tx, ty)+1)%len(trees)]
		}
		treeSnippet = fmt.Sprintf("several %s and %s trees", strings.ToLower(t1.Name), strings.ToLower(t2.Name))
	}

	plantSnippet := "wild plants around the ground cover"
	if len(plants) > 0 {
		p := plants[s.lookIndex("plant-ahead", len(plants), tx, ty)]
		plantSnippet = fmt.Sprintf("wild plants nearby, including %s", strings.ToLower(p.Name))
	}

	insectSnippet := "insects stir through the brush"
	if hasInsectType(insects, "ant") {
		insectSnippet = "an ant colony appears to be moving over a fallen log"
	} else if len(insects) > 0 {
		insectSnippet = fmt.Sprintf("%s activity flickers around the undergrowth", strings.ToLower(insects[s.lookIndex("insect-ahead", len(insects), tx, ty)]))
	}

	waterSnippet := ""
	if cell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake|TopoFlagCoast) != 0 {
		waterSnippet = " Water glints through the terrain in that direction."
	}

	return fmt.Sprintf("Looking %s (%s), you see %s terrain. %s; %s; %s.%s",
		posLabel, dir, biome, treeSnippet, insectSnippet, plantSnippet, waterSnippet)
}

func (s *RunState) describeLookCloser(playerID int, relative, dir, subject string, cell TopoCell, tx, ty int, inBounds bool) string {
	if !inBounds {
		return fmt.Sprintf("Looking closer %s (%s), the boundary blocks further detail.", lookRelativeLabel(relative), dir)
	}
	norm := strings.ToLower(strings.TrimSpace(subject))
	if norm == "" || norm == "area" {
		norm = "plants"
	}
	if strings.Contains(norm, "plant") || strings.Contains(norm, "herb") || strings.Contains(norm, "berry") {
		plants := s.lookPlantsForSeason()
		if len(plants) == 0 {
			return fmt.Sprintf("Looking closer at the plants %s, you cannot identify anything edible yet.", lookRelativeLabel(relative))
		}
		p := plants[s.lookIndex("plant-close", len(plants), tx, ty)]
		detail := "edible"
		if p.Category == PlantCategoryMedicinal || p.Medicinal > 0 {
			detail = "medicinal"
		}
		if p.Toxicity >= 3 || p.Category == PlantCategoryToxic {
			detail = "toxic"
		}
		return fmt.Sprintf("Looking closer at the plants %s, you locate %s (%s).", lookRelativeLabel(relative), p.Name, detail)
	}
	if strings.Contains(norm, "tree") || strings.Contains(norm, "wood") || strings.Contains(norm, "log") {
		trees := TreesForBiome(s.Scenario.Biome)
		if len(trees) == 0 {
			return "Looking closer at the trees, visibility is poor."
		}
		t := trees[s.lookIndex("tree-close", len(trees), tx, ty)]
		bark := t.BarkResource
		if bark == "" {
			bark = "usable bark fiber"
		}
		return fmt.Sprintf("Looking closer at the trees %s, you identify %s. Bark resource: %s.", lookRelativeLabel(relative), t.Name, bark)
	}
	if strings.Contains(norm, "insect") || strings.Contains(norm, "ant") || strings.Contains(norm, "bug") {
		insects := InsectsForBiome(s.Scenario.Biome)
		if len(insects) == 0 {
			return "Looking closer, you do not see insect activity."
		}
		pick := insects[s.lookIndex("insect-close", len(insects), tx, ty)]
		return fmt.Sprintf("Looking closer at the insect activity %s, you spot %s.", lookRelativeLabel(relative), strings.ToLower(pick))
	}
	if strings.Contains(norm, "water") || strings.Contains(norm, "river") || strings.Contains(norm, "lake") || strings.Contains(norm, "coast") {
		if cell.Flags&(TopoFlagWater|TopoFlagRiver|TopoFlagLake|TopoFlagCoast) == 0 {
			return fmt.Sprintf("Looking closer %s, you do not see open water from this position.", lookRelativeLabel(relative))
		}
		return fmt.Sprintf("Looking closer %s, you confirm water access nearby.", lookRelativeLabel(relative))
	}
	return fmt.Sprintf("Looking closer %s, you note %s terrain with signs of resources.", lookRelativeLabel(relative), topoBiomeLabel(cell.Biome))
}

func (s *RunState) lookPlantsForSeason() []PlantSpec {
	season, ok := s.CurrentSeason()
	if ok {
		return PlantsForBiomeSeason(s.Scenario.Biome, PlantCategoryAny, season)
	}
	return PlantsForBiome(s.Scenario.Biome, PlantCategoryAny)
}

func (s *RunState) absoluteLookDirection(relative string) string {
	facing := normalizeDirection(s.Travel.Direction)
	if facing == "" {
		facing = "north"
	}
	switch strings.ToLower(strings.TrimSpace(relative)) {
	case "left":
		return rotateDirection(facing, -1)
	case "right":
		return rotateDirection(facing, 1)
	case "back", "behind":
		return rotateDirection(facing, 2)
	default:
		return facing
	}
}

func rotateDirection(direction string, steps int) string {
	order := []string{"north", "east", "south", "west"}
	index := 0
	for i, d := range order {
		if d == normalizeDirection(direction) {
			index = i
			break
		}
	}
	n := len(order)
	next := (index + steps) % n
	if next < 0 {
		next += n
	}
	return order[next]
}

func (s *RunState) stepInDirection(x, y int, direction string) (int, int, bool) {
	dx, dy := directionDelta(direction)
	nx := x + dx
	ny := y + dy
	if s.Topology.Width > 0 && (nx < 0 || nx >= s.Topology.Width) {
		return x, y, false
	}
	if s.Topology.Height > 0 && (ny < 0 || ny >= s.Topology.Height) {
		return x, y, false
	}
	return nx, ny, true
}

func (s *RunState) lookIndex(tag string, n int, x int, y int) int {
	if n <= 1 {
		return 0
	}
	rng := seededRNG(seedFromLabel(s.Config.Seed, fmt.Sprintf("look:%s:%d:%d:%d", tag, s.Day, x, y)))
	return rng.IntN(n)
}

func normalizeLookRelative(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "front", "forward", "ahead":
		return "front", true
	case "left":
		return "left", true
	case "right":
		return "right", true
	case "back", "behind":
		return "back", true
	default:
		return "", false
	}
}

func lookRelativeLabel(relative string) string {
	switch relative {
	case "left":
		return "to your left"
	case "right":
		return "to your right"
	case "back":
		return "behind you"
	default:
		return "in front of you"
	}
}

func topoBiomeLabel(biome uint8) string {
	switch biome {
	case TopoBiomeForest:
		return "forested"
	case TopoBiomeGrassland:
		return "grassland"
	case TopoBiomeJungle:
		return "jungle"
	case TopoBiomeWetland:
		return "wetland"
	case TopoBiomeSwamp:
		return "swamp"
	case TopoBiomeDesert:
		return "dry desert"
	case TopoBiomeMountain:
		return "mountainous"
	case TopoBiomeTundra:
		return "tundra"
	case TopoBiomeBoreal:
		return "boreal forest"
	default:
		return "mixed terrain"
	}
}

func hasInsectType(insects []string, key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, insect := range insects {
		if strings.Contains(strings.ToLower(strings.TrimSpace(insect)), key) {
			return true
		}
	}
	return false
}
