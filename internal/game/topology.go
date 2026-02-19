package game

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

const (
	TopoFlagWater uint8 = 1 << iota
	TopoFlagRiver
	TopoFlagLake
	TopoFlagCoast
)

const (
	TopoBiomeUnknown uint8 = iota
	TopoBiomeForest
	TopoBiomeGrassland
	TopoBiomeJungle
	TopoBiomeWetland
	TopoBiomeSwamp
	TopoBiomeDesert
	TopoBiomeMountain
	TopoBiomeTundra
	TopoBiomeBoreal
)

type TopoCell struct {
	Elevation   int8  `json:"elevation"`
	Moisture    uint8 `json:"moisture"`
	Temperature uint8 `json:"temperature"`
	Biome       uint8 `json:"biome"`
	Flags       uint8 `json:"flags"`
	Roughness   uint8 `json:"roughness"`
}

type WorldTopology struct {
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Cells  []TopoCell `json:"cells"`
}

type CellState struct {
	HuntPressure uint8 `json:"hunt_pressure"`
	Disturbance  uint8 `json:"disturbance"`
	Depletion    uint8 `json:"depletion"`
	CarcassToken uint8 `json:"carcass_token,omitempty"`
}

type TimeBlock string

const (
	TimeBlockDawn  TimeBlock = "dawn"
	TimeBlockDay   TimeBlock = "day"
	TimeBlockDusk  TimeBlock = "dusk"
	TimeBlockNight TimeBlock = "night"
)

func (s *RunState) CurrentTimeBlock() TimeBlock {
	if s == nil {
		return TimeBlockDay
	}
	h := s.ClockHours
	switch {
	case h >= 5 && h < 8:
		return TimeBlockDawn
	case h >= 8 && h < 18:
		return TimeBlockDay
	case h >= 18 && h < 21:
		return TimeBlockDusk
	default:
		return TimeBlockNight
	}
}

func defaultTopologySizeForMode(mode GameMode) (int, int) {
	switch mode {
	case ModeAlone:
		return 36, 36
	case ModeNakedAndAfraid:
		return 100, 100
	case ModeNakedAndAfraidXL:
		return 125, 125
	default:
		return 72, 72
	}
}

func clampTopologySize(mode GameMode, width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return defaultTopologySizeForMode(mode)
	}
	switch mode {
	case ModeAlone:
		width = clamp(width, 28, 46)
		height = clamp(height, 28, 46)
	case ModeNakedAndAfraid:
		width = clamp(width, 88, 125)
		height = clamp(height, 88, 125)
	case ModeNakedAndAfraidXL:
		width = clamp(width, 100, 150)
		height = clamp(height, 100, 150)
	default:
		width = clamp(width, 50, 140)
		height = clamp(height, 50, 140)
	}
	return width, height
}

func topologySizeForScenario(mode GameMode, scenario Scenario) (int, int) {
	if scenario.MapWidthCells > 0 && scenario.MapHeightCells > 0 {
		return clampTopologySize(mode, scenario.MapWidthCells, scenario.MapHeightCells)
	}
	return defaultTopologySizeForMode(mode)
}

func (s *RunState) EnsureTopology() {
	if s == nil {
		return
	}
	if s.Topology.Width > 0 && s.Topology.Height > 0 && len(s.Topology.Cells) == s.Topology.Width*s.Topology.Height {
		if len(s.CellStates) != len(s.Topology.Cells) {
			s.CellStates = make([]CellState, len(s.Topology.Cells))
		}
		if len(s.FogMask) != len(s.Topology.Cells) {
			s.FogMask = make([]bool, len(s.Topology.Cells))
			if s.Config.Mode != ModeAlone {
				for i := range s.FogMask {
					s.FogMask[i] = true
				}
			} else {
				s.RevealFog(s.Travel.PosX, s.Travel.PosY, 1)
			}
		}
		return
	}
	s.initTopology()
}

func (s *RunState) initTopology() {
	if s == nil {
		return
	}
	w, h := topologySizeForScenario(s.Config.Mode, s.Scenario)
	topology := GenerateWorldTopology(s.Config.Seed, s.Scenario.Biome, w, h)
	s.Topology = topology
	s.CellStates = make([]CellState, len(topology.Cells))
	s.FogMask = make([]bool, len(topology.Cells))
	if s.Config.Mode != ModeAlone {
		for i := range s.FogMask {
			s.FogMask[i] = true
		}
	}
	startX, startY := pickTopologyStartCell(topology)
	s.Travel.PosX = startX
	s.Travel.PosY = startY
	s.RevealFog(startX, startY, 1)
}

func pickTopologyStartCell(topology WorldTopology) (int, int) {
	if topology.Width <= 0 || topology.Height <= 0 {
		return 0, 0
	}
	cx := topology.Width / 2
	cy := topology.Height / 2
	idx := cy*topology.Width + cx
	if idx >= 0 && idx < len(topology.Cells) && topology.Cells[idx].Flags&TopoFlagWater == 0 {
		return cx, cy
	}
	bestX, bestY := cx, cy
	bestDist := math.MaxFloat64
	for y := 0; y < topology.Height; y++ {
		for x := 0; x < topology.Width; x++ {
			i := y*topology.Width + x
			if i < 0 || i >= len(topology.Cells) {
				continue
			}
			if topology.Cells[i].Flags&TopoFlagWater != 0 {
				continue
			}
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < bestDist {
				bestDist = dist
				bestX, bestY = x, y
			}
		}
	}
	return bestX, bestY
}

func (s *RunState) topoIndex(x, y int) (int, bool) {
	if s == nil {
		return 0, false
	}
	if x < 0 || y < 0 || x >= s.Topology.Width || y >= s.Topology.Height {
		return 0, false
	}
	idx := y*s.Topology.Width + x
	if idx < 0 || idx >= len(s.Topology.Cells) {
		return 0, false
	}
	return idx, true
}

func (s *RunState) TopologyCellAt(x, y int) (TopoCell, bool) {
	idx, ok := s.topoIndex(x, y)
	if !ok {
		return TopoCell{}, false
	}
	return s.Topology.Cells[idx], true
}

func (s *RunState) CurrentMapPosition() (int, int) {
	if s == nil {
		return 0, 0
	}
	return s.Travel.PosX, s.Travel.PosY
}

func (s *RunState) IsRevealed(x, y int) bool {
	if s == nil {
		return false
	}
	if s.Config.Mode != ModeAlone {
		return true
	}
	idx, ok := s.topoIndex(x, y)
	if !ok {
		return false
	}
	if len(s.FogMask) != len(s.Topology.Cells) {
		return false
	}
	return s.FogMask[idx]
}

func (s *RunState) RevealFog(x, y, radius int) {
	if s == nil {
		return
	}
	if s.Config.Mode != ModeAlone {
		return
	}
	if len(s.FogMask) != len(s.Topology.Cells) {
		return
	}
	if radius < 0 {
		radius = 0
	}
	for oy := -radius; oy <= radius; oy++ {
		for ox := -radius; ox <= radius; ox++ {
			if int(math.Abs(float64(ox))+math.Abs(float64(oy))) > radius+1 {
				continue
			}
			idx, ok := s.topoIndex(x+ox, y+oy)
			if !ok {
				continue
			}
			s.FogMask[idx] = true
		}
	}
}

func (s *RunState) decayCellStates() {
	if s == nil || len(s.CellStates) == 0 {
		return
	}
	for i := range s.CellStates {
		cs := &s.CellStates[i]
		if cs.HuntPressure > 0 {
			cs.HuntPressure = uint8(max(0, int(cs.HuntPressure)-2))
		}
		if cs.Disturbance > 0 {
			cs.Disturbance = uint8(max(0, int(cs.Disturbance)-5))
		}
		if cs.Depletion > 0 {
			cs.Depletion = uint8(max(0, int(cs.Depletion)-1))
		}
		if cs.CarcassToken > 0 {
			cs.CarcassToken = uint8(max(0, int(cs.CarcassToken)-1))
		}
	}
}

func (s *RunState) applyCellStateAction(x, y int, action string) {
	if s == nil {
		return
	}
	idx, ok := s.topoIndex(x, y)
	if !ok || len(s.CellStates) <= idx {
		return
	}
	cs := &s.CellStates[idx]
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "move":
		cs.Disturbance = uint8(min(255, int(cs.Disturbance)+1))
	case "forage":
		cs.Disturbance = uint8(min(255, int(cs.Disturbance)+2))
		cs.Depletion = uint8(min(255, int(cs.Depletion)+10))
	case "hunt":
		cs.Disturbance = uint8(min(255, int(cs.Disturbance)+7))
		cs.HuntPressure = uint8(min(255, int(cs.HuntPressure)+14))
		cs.CarcassToken = uint8(min(255, int(cs.CarcassToken)+2))
	case "fish":
		cs.Disturbance = uint8(min(255, int(cs.Disturbance)+5))
		cs.HuntPressure = uint8(min(255, int(cs.HuntPressure)+9))
		cs.Depletion = uint8(min(255, int(cs.Depletion)+4))
		cs.CarcassToken = uint8(min(255, int(cs.CarcassToken)+1))
	}
}

func topologyHash(seed int64, x, y int, salt string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d:%d:%s", seed, x, y, salt)))
	return h.Sum64()
}

func hashUnitFloat(seed int64, x, y int, salt string) float64 {
	return float64(topologyHash(seed, x, y, salt)&0xfffffff) / float64(0xfffffff)
}

func smoothstep(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}

func valueNoise2D(seed int64, x, y, cellSize float64, salt string) float64 {
	if cellSize <= 0 {
		cellSize = 1
	}
	gx := x / cellSize
	gy := y / cellSize
	x0 := int(math.Floor(gx))
	y0 := int(math.Floor(gy))
	x1 := x0 + 1
	y1 := y0 + 1
	tx := smoothstep(gx - float64(x0))
	ty := smoothstep(gy - float64(y0))
	n00 := hashUnitFloat(seed, x0, y0, salt)
	n10 := hashUnitFloat(seed, x1, y0, salt)
	n01 := hashUnitFloat(seed, x0, y1, salt)
	n11 := hashUnitFloat(seed, x1, y1, salt)
	nx0 := n00 + (n10-n00)*tx
	nx1 := n01 + (n11-n01)*tx
	return nx0 + (nx1-nx0)*ty
}

func layeredNoise(seed int64, x, y float64, salt string) float64 {
	scales := []float64{52, 26, 13, 6}
	amplitude := 1.0
	total := 0.0
	weight := 0.0
	for i, scale := range scales {
		s := fmt.Sprintf("%s:%d", salt, i)
		total += valueNoise2D(seed, x, y, scale, s) * amplitude
		weight += amplitude
		amplitude *= 0.5
	}
	if weight <= 0 {
		return 0.5
	}
	return total / weight
}

func initialTopoBiome(elev int8, moisture, temp uint8) uint8 {
	e := int(elev)
	m := int(moisture)
	t := int(temp)
	switch {
	case e >= 64:
		return TopoBiomeMountain
	case t <= 62:
		if m >= 140 {
			return TopoBiomeBoreal
		}
		return TopoBiomeTundra
	case m >= 205 && t >= 160:
		return TopoBiomeJungle
	case m >= 185:
		return TopoBiomeWetland
	case m >= 168 && t >= 140:
		return TopoBiomeSwamp
	case m <= 70 && t >= 150:
		return TopoBiomeDesert
	case m <= 95:
		return TopoBiomeGrassland
	default:
		return TopoBiomeForest
	}
}

func roughnessForBiome(biome uint8, noise float64, elevation int8) uint8 {
	base := int(math.Round(1 + noise*5))
	switch biome {
	case TopoBiomeMountain:
		base += 3
	case TopoBiomeSwamp, TopoBiomeWetland, TopoBiomeJungle:
		base += 2
	case TopoBiomeDesert:
		base += 1
	}
	if elevation > 75 {
		base++
	}
	if base < 1 {
		base = 1
	}
	if base > 9 {
		base = 9
	}
	return uint8(base)
}

func GenerateWorldTopology(seed int64, biome string, width, height int) WorldTopology {
	if width < 8 {
		width = 8
	}
	if height < 8 {
		height = 8
	}
	cells := make([]TopoCell, width*height)
	biomeNorm := normalizeBiome(biome)
	tempBias := 0.0
	moistureBias := 0.0
	if biomeIsArctic(biomeNorm) {
		tempBias -= 0.22
	}
	if biomeIsDesertOrDry(biomeNorm) {
		tempBias += 0.09
		moistureBias -= 0.16
	}
	if biomeIsTropicalWet(biomeNorm) {
		tempBias += 0.16
		moistureBias += 0.2
	}
	if strings.Contains(biomeNorm, "coast") || strings.Contains(biomeNorm, "island") || strings.Contains(biomeNorm, "delta") {
		moistureBias += 0.12
	}

	for y := 0; y < height; y++ {
		lat := float64(y) / float64(max(1, height-1))
		latTemp := 1.0 - math.Abs(lat-0.5)*0.7
		for x := 0; x < width; x++ {
			idx := y*width + x
			nElev := layeredNoise(seed, float64(x), float64(y), "elev")
			nRidge := layeredNoise(seed+11, float64(x), float64(y), "ridge")
			nMoist := layeredNoise(seed+31, float64(x), float64(y), "moist")
			nTemp := layeredNoise(seed+47, float64(x), float64(y), "temp")
			elevVal := ((nElev - 0.5) * 120.0) + ((math.Abs(nRidge-0.5) - 0.25) * 34.0) - 8
			elevation := int8(clamp(int(math.Round(elevVal)), -90, 90))
			tempVal := clampFloat(latTemp+(nTemp-0.5)*0.35+tempBias-float64(elevation)/220.0, 0, 1)
			moistVal := clampFloat(nMoist+moistureBias-float64(elevation)/280.0, 0, 1)
			temp := uint8(math.Round(tempVal * 255))
			moist := uint8(math.Round(moistVal * 255))

			flags := uint8(0)
			if elevation <= -22 {
				flags |= TopoFlagWater
			}
			if (strings.Contains(biomeNorm, "coast") || strings.Contains(biomeNorm, "island")) && (x < 2 || y < 2 || x > width-3 || y > height-3) {
				flags |= TopoFlagWater
			}

			b := initialTopoBiome(elevation, moist, temp)
			rough := roughnessForBiome(b, layeredNoise(seed+91, float64(x), float64(y), "rough"), elevation)
			cells[idx] = TopoCell{
				Elevation:   elevation,
				Moisture:    moist,
				Temperature: temp,
				Biome:       b,
				Flags:       flags,
				Roughness:   rough,
			}
		}
	}

	flowTo := make([]int, len(cells))
	accum := make([]int, len(cells))
	for i := range flowTo {
		flowTo[i] = -1
		accum[i] = 1
	}
	neigh := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if cells[idx].Flags&TopoFlagWater != 0 {
				continue
			}
			bestIdx := -1
			bestElev := int(cells[idx].Elevation)
			for _, off := range neigh {
				nx := x + off[0]
				ny := y + off[1]
				if nx < 0 || ny < 0 || nx >= width || ny >= height {
					continue
				}
				nIdx := ny*width + nx
				if int(cells[nIdx].Elevation) < bestElev {
					bestElev = int(cells[nIdx].Elevation)
					bestIdx = nIdx
				}
			}
			flowTo[idx] = bestIdx
		}
	}
	order := make([]int, len(cells))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		return cells[order[i]].Elevation > cells[order[j]].Elevation
	})
	for _, idx := range order {
		next := flowTo[idx]
		if next >= 0 {
			accum[next] += accum[idx]
		}
	}
	riverThreshold := max(16, (width*height)/180)
	for idx := range cells {
		if cells[idx].Flags&TopoFlagWater != 0 {
			continue
		}
		if accum[idx] >= riverThreshold {
			cells[idx].Flags |= TopoFlagRiver | TopoFlagWater
		}
		if cells[idx].Flags&TopoFlagWater == 0 && int(cells[idx].Elevation) <= -10 && cells[idx].Moisture >= 180 {
			cells[idx].Flags |= TopoFlagLake | TopoFlagWater
		}
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			cell := &cells[idx]
			if cell.Flags&TopoFlagWater != 0 {
				continue
			}
			for _, off := range neigh[:4] {
				nx := x + off[0]
				ny := y + off[1]
				if nx < 0 || ny < 0 || nx >= width || ny >= height {
					continue
				}
				nIdx := ny*width + nx
				if cells[nIdx].Flags&TopoFlagWater != 0 {
					cell.Flags |= TopoFlagCoast
					break
				}
			}
			if cell.Flags&TopoFlagCoast != 0 {
				if cell.Biome == TopoBiomeGrassland || cell.Biome == TopoBiomeDesert {
					cell.Biome = TopoBiomeForest
				}
			}
		}
	}

	return WorldTopology{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}
