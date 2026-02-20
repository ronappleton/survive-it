package gui

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Discovery summary:
// - Minimap/full-map rendering both route through drawTopologyRegion and share cell color logic.
// - Water visuals previously ignored runtime temperature, so freezing conditions looked like open water.
// - A debug env flag now surfaces biome/temp/frozen coherence data without changing default UI flow.

const (
	runLogSplitRatio = 0.66
	runLayoutPadding = 16
	runLayoutGap     = 10
	miniMapViewCells = 31
	topoRenderDetail = 16
)

type runLayout struct {
	Outer       rl.Rectangle
	TopRect     rl.Rectangle
	LogRect     rl.Rectangle
	MiniMapRect rl.Rectangle
	InputRect   rl.Rectangle
}

func runScreenLayout(width, height int32) runLayout {
	outer := rl.NewRectangle(runLayoutPadding, runLayoutPadding, float32(width-runLayoutPadding*2), float32(height-runLayoutPadding*2))
	topH := float32(212)
	inputH := float32(92)
	if outer.Height < 520 {
		topH = 186
		inputH = 82
	}
	gap := float32(runLayoutGap)
	middleTop := outer.Y + topH + gap
	inputTop := outer.Y + outer.Height - inputH
	if inputTop-middleTop-gap < 160 {
		inputTop = middleTop + gap + 160
	}
	middleH := inputTop - middleTop - gap
	splitX := outer.X + outer.Width*runLogSplitRatio
	logW := splitX - outer.X - gap/2
	miniX := splitX + gap/2
	return runLayout{
		Outer:       outer,
		TopRect:     rl.NewRectangle(outer.X, outer.Y, outer.Width, topH),
		LogRect:     rl.NewRectangle(outer.X, middleTop, logW, middleH),
		MiniMapRect: rl.NewRectangle(miniX, middleTop, outer.X+outer.Width-miniX, middleH),
		InputRect:   rl.NewRectangle(outer.X, inputTop, outer.Width, outer.Y+outer.Height-inputTop),
	}
}

func drawRunMessageLog(rect rl.Rectangle, messages []string) {
	maxWidth := int32(rect.Width - spaceM*2)
	lineHeight := textLineHeight(typeScale.Log)
	maxLines := int((rect.Height - 52) / float32(lineHeight))
	if maxLines < 4 {
		maxLines = 4
	}
	flattened := make([]string, 0, maxLines)
	for i := len(messages) - 1; i >= 0 && len(flattened) < maxLines; i-- {
		lines := wrapText(messages[i], typeScale.Log, maxWidth)
		for j := len(lines) - 1; j >= 0 && len(flattened) < maxLines; j-- {
			flattened = append(flattened, lines[j])
		}
	}
	y := int32(rect.Y+rect.Height) - 26
	for i := 0; i < len(flattened); i++ {
		line := flattened[i]
		drawLogText(line, int32(rect.X+spaceM), y-lineHeight, typeScale.Log, colorText)
		y -= lineHeight
		if y < int32(rect.Y)+26 {
			break
		}
	}
}

func topoBiomeColor(b uint8) rl.Color {
	switch b {
	case game.TopoBiomeForest:
		return rl.NewColor(73, 103, 80, 255)
	case game.TopoBiomeGrassland:
		return rl.NewColor(112, 126, 78, 255)
	case game.TopoBiomeJungle:
		return rl.NewColor(58, 94, 72, 255)
	case game.TopoBiomeWetland:
		return rl.NewColor(72, 95, 86, 255)
	case game.TopoBiomeSwamp:
		return rl.NewColor(63, 84, 75, 255)
	case game.TopoBiomeDesert:
		return rl.NewColor(148, 126, 85, 255)
	case game.TopoBiomeMountain:
		return rl.NewColor(117, 121, 128, 255)
	case game.TopoBiomeTundra:
		return rl.NewColor(141, 146, 151, 255)
	case game.TopoBiomeBoreal:
		return rl.NewColor(78, 104, 90, 255)
	default:
		return rl.NewColor(90, 96, 100, 255)
	}
}

func shadeByElevation(clr rl.Color, elevation int8) rl.Color {
	f := 1.0 + float64(elevation)/230.0
	if f < 0.55 {
		f = 0.55
	}
	if f > 1.3 {
		f = 1.3
	}
	return rl.NewColor(
		uint8(clampInt(int(float64(clr.R)*f), 0, 255)),
		uint8(clampInt(int(float64(clr.G)*f), 0, 255)),
		uint8(clampInt(int(float64(clr.B)*f), 0, 255)),
		clr.A,
	)
}

func colorScale(clr rl.Color, factor float64) rl.Color {
	if factor < 0.45 {
		factor = 0.45
	}
	if factor > 1.55 {
		factor = 1.55
	}
	return rl.NewColor(
		uint8(clampInt(int(float64(clr.R)*factor), 0, 255)),
		uint8(clampInt(int(float64(clr.G)*factor), 0, 255)),
		uint8(clampInt(int(float64(clr.B)*factor), 0, 255)),
		clr.A,
	)
}

func topoCellClamp(topology game.WorldTopology, x, y int) game.TopoCell {
	if topology.Width <= 0 || topology.Height <= 0 || len(topology.Cells) == 0 {
		return game.TopoCell{}
	}
	x = clampInt(x, 0, topology.Width-1)
	y = clampInt(y, 0, topology.Height-1)
	return topology.Cells[y*topology.Width+x]
}

func topoElevationSample(topology game.WorldTopology, xf, yf float64) float64 {
	if topology.Width <= 0 || topology.Height <= 0 || len(topology.Cells) == 0 {
		return 0
	}
	x0 := int(math.Floor(xf))
	y0 := int(math.Floor(yf))
	tx := xf - float64(x0)
	ty := yf - float64(y0)
	x1 := x0 + 1
	y1 := y0 + 1

	e00 := float64(topoCellClamp(topology, x0, y0).Elevation)
	e10 := float64(topoCellClamp(topology, x1, y0).Elevation)
	e01 := float64(topoCellClamp(topology, x0, y1).Elevation)
	e11 := float64(topoCellClamp(topology, x1, y1).Elevation)

	e0 := e00 + (e10-e00)*tx
	e1 := e01 + (e11-e01)*tx
	return e0 + (e1-e0)*ty
}

func topoContourBand(elevation int8) int {
	// ~10m buckets (given the game's compact elevation scale) for visible contouring.
	return (int(elevation) + 128) / 10
}

func terrainReliefShade(topology game.WorldTopology, xf, yf float64, base rl.Color) rl.Color {
	e := topoElevationSample(topology, xf, yf)
	gx := topoElevationSample(topology, xf+0.5, yf) - topoElevationSample(topology, xf-0.5, yf)
	gy := topoElevationSample(topology, xf, yf+0.5) - topoElevationSample(topology, xf, yf-0.5)

	// Elevation brightening + directional hillshade to make relief obvious.
	elevFactor := 1.0 + e/170.0
	hillshade := 1.0 + ((-gx * 0.028) + (-gy * 0.018))
	return colorScale(base, elevFactor*hillshade)
}

type squareGridGeometry struct {
	OriginX  float32
	OriginY  float32
	CellSize float32
	Cols     int
	Rows     int
	DrawRect rl.Rectangle
}

func computeSquareGridGeometry(area rl.Rectangle, cols, rows int) (squareGridGeometry, bool) {
	if cols <= 0 || rows <= 0 || area.Width <= 1 || area.Height <= 1 {
		return squareGridGeometry{}, false
	}
	cellSize := float32(math.Min(float64(area.Width/float32(cols)), float64(area.Height/float32(rows))))
	if cellSize < 1 {
		cellSize = 1
	}
	drawWidth := cellSize * float32(cols)
	drawHeight := cellSize * float32(rows)
	originX := area.X + (area.Width-drawWidth)/2
	originY := area.Y + (area.Height-drawHeight)/2
	return squareGridGeometry{
		OriginX:  originX,
		OriginY:  originY,
		CellSize: cellSize,
		Cols:     cols,
		Rows:     rows,
		DrawRect: rl.NewRectangle(originX, originY, drawWidth, drawHeight),
	}, true
}

func computeMiniMapWindow(topologyW, topologyH, playerX, playerY, desired int) (startX, startY, cols, rows int) {
	if topologyW <= 0 || topologyH <= 0 {
		return 0, 0, 0, 0
	}
	if desired < 5 {
		desired = 5
	}
	cols = min(topologyW, desired)
	rows = min(topologyH, desired)
	startX = clampInt(playerX-cols/2, 0, max(0, topologyW-cols))
	startY = clampInt(playerY-rows/2, 0, max(0, topologyH-rows))
	return startX, startY, cols, rows
}

func (ui *gameUI) drawTopologyRegion(area rl.Rectangle, startX, startY, cols, rows int) {
	if ui.run == nil || cols <= 0 || rows <= 0 {
		return
	}
	topology := ui.run.Topology
	waterFrozen := ui.run.IsWaterCurrentlyFrozen()
	detail := topoRenderDetail
	maxDetailX := int(math.Floor(float64(area.Width) / float64(cols)))
	maxDetailY := int(math.Floor(float64(area.Height) / float64(rows)))
	if maxDetailX < detail {
		detail = maxDetailX
	}
	if maxDetailY < detail {
		detail = maxDetailY
	}
	if detail < 1 {
		detail = 1
	}

	renderCols := cols * detail
	renderRows := rows * detail
	geo, ok := computeSquareGridGeometry(area, renderCols, renderRows)
	if !ok {
		return
	}
	for y := 0; y < renderRows; y++ {
		worldYf := float64(startY) + (float64(y)+0.5)/float64(detail)
		worldY := clampInt(int(worldYf), 0, topology.Height-1)
		for x := 0; x < renderCols; x++ {
			worldXf := float64(startX) + (float64(x)+0.5)/float64(detail)
			worldX := clampInt(int(worldXf), 0, topology.Width-1)
			idx := worldY*topology.Width + worldX
			cell := topology.Cells[idx]
			clr := topoBiomeColor(cell.Biome)
			if cell.Flags&game.TopoFlagWater != 0 {
				if waterFrozen {
					clr = rl.NewColor(139, 146, 152, 255)
				} else {
					clr = rl.NewColor(78, 101, 117, 255)
				}
			}
			if cell.Flags&game.TopoFlagRiver != 0 {
				if waterFrozen {
					clr = rl.NewColor(147, 154, 160, 255)
				} else {
					clr = rl.NewColor(87, 111, 128, 255)
				}
			}
			if cell.Flags&game.TopoFlagLake != 0 {
				if waterFrozen {
					clr = rl.NewColor(143, 150, 157, 255)
				} else {
					clr = rl.NewColor(84, 107, 124, 255)
				}
			}
			if cell.Flags&game.TopoFlagCoast != 0 {
				clr = rl.NewColor(
					uint8(min(255, int(clr.R)+18)),
					uint8(min(255, int(clr.G)+18)),
					uint8(min(255, int(clr.B)+10)),
					255,
				)
			}
			if ui.run.Config.Mode == game.ModeAlone && !ui.run.IsRevealed(worldX, worldY) {
				clr = colorBG
			} else if cell.Flags&game.TopoFlagWater == 0 {
				clr = terrainReliefShade(topology, worldXf, worldYf, clr)
			} else {
				clr = shadeByElevation(clr, cell.Elevation)
			}
			x0 := int32(geo.OriginX + float32(x)*geo.CellSize)
			y0 := int32(geo.OriginY + float32(y)*geo.CellSize)
			x1 := int32(geo.OriginX + float32(x+1)*geo.CellSize)
			y1 := int32(geo.OriginY + float32(y+1)*geo.CellSize)
			w := max(1, int(x1-x0))
			h := max(1, int(y1-y0))
			rl.DrawRectangle(x0, y0, int32(w), int32(h), clr)
		}
	}

	baseStep := geo.CellSize * float32(detail)
	for y := 0; y < rows; y++ {
		worldY := startY + y
		for x := 0; x < cols; x++ {
			worldX := startX + x
			cell := topoCellClamp(topology, worldX, worldY)
			if ui.run.Config.Mode == game.ModeAlone && !ui.run.IsRevealed(worldX, worldY) {
				continue
			}
			thisBand := topoContourBand(cell.Elevation)
			lineX := geo.OriginX + float32((x+1)*detail)*geo.CellSize
			lineY := geo.OriginY + float32(y*detail)*geo.CellSize
			if x+1 < cols {
				rightX := worldX + 1
				if ui.run.Config.Mode != game.ModeAlone || ui.run.IsRevealed(rightX, worldY) {
					rightBand := topoContourBand(topoCellClamp(topology, rightX, worldY).Elevation)
					if rightBand != thisBand {
						rl.DrawLineEx(
							rl.NewVector2(lineX, lineY),
							rl.NewVector2(lineX, lineY+baseStep),
							1.1,
							rl.Fade(AppTheme.BorderStrong, 0.5),
						)
					}
				}
			}
			if y+1 < rows {
				downY := worldY + 1
				if ui.run.Config.Mode != game.ModeAlone || ui.run.IsRevealed(worldX, downY) {
					downBand := topoContourBand(topoCellClamp(topology, worldX, downY).Elevation)
					if downBand != thisBand {
						hy := geo.OriginY + float32((y+1)*detail)*geo.CellSize
						hx := geo.OriginX + float32(x*detail)*geo.CellSize
						rl.DrawLineEx(
							rl.NewVector2(hx, hy),
							rl.NewVector2(hx+baseStep, hy),
							1.1,
							rl.Fade(AppTheme.BorderStrong, 0.5),
						)
					}
				}
			}
		}
	}

	rl.DrawRectangleLinesEx(geo.DrawRect, 1.0, rl.Fade(colorBorder, 0.8))

	px, py := ui.run.CurrentMapPosition()
	if px >= startX && px < startX+cols && py >= startY && py < startY+rows {
		localX := px - startX
		localY := py - startY
		cellStep := geo.CellSize * float32(detail)
		cx := geo.OriginX + (float32(localX)+0.5)*cellStep
		cy := geo.OriginY + (float32(localY)+0.5)*cellStep
		r := float32(3)
		if cellStep > 6 {
			r = cellStep * 0.35
		}
		if r < 2 {
			r = 2
		}
		rl.DrawCircle(int32(cx), int32(cy), r, colorDanger)
	}
}

func (ui *gameUI) drawTopologyMap(rect rl.Rectangle, withLegend bool) {
	if ui.run == nil {
		return
	}
	ui.run.EnsureTopology()
	topology := ui.run.Topology
	if topology.Width <= 0 || topology.Height <= 0 || len(topology.Cells) == 0 {
		drawWrappedText("No topology data.", rect, 44, 18, colorWarn)
		return
	}
	area := rl.NewRectangle(rect.X+10, rect.Y+36, rect.Width-20, rect.Height-48)
	if withLegend {
		area = rl.NewRectangle(rect.X+10, rect.Y+36, rect.Width-182, rect.Height-48)
	}
	if area.Width < 20 || area.Height < 20 {
		return
	}
	ui.drawTopologyRegion(area, 0, 0, topology.Width, topology.Height)

	if withLegend {
		legendX := int32(rect.X + rect.Width - 162)
		legendY := int32(rect.Y + 40)
		drawText("Legend", legendX, legendY, typeScale.Body, colorAccent)
		legendY += 24
		legendRows := []struct {
			Label string
			Color rl.Color
		}{
			{Label: "Forest", Color: topoBiomeColor(game.TopoBiomeForest)},
			{Label: "Grassland", Color: topoBiomeColor(game.TopoBiomeGrassland)},
			{Label: "Mountain", Color: topoBiomeColor(game.TopoBiomeMountain)},
			{Label: "Desert", Color: topoBiomeColor(game.TopoBiomeDesert)},
			{Label: "Wetland/Jungle", Color: topoBiomeColor(game.TopoBiomeWetland)},
			{Label: "Water/River", Color: rl.NewColor(84, 107, 124, 255)},
			{Label: "Ice (frozen)", Color: rl.NewColor(143, 150, 157, 255)},
			{Label: "Player", Color: colorDanger},
		}
		for _, row := range legendRows {
			rl.DrawRectangle(legendX, legendY+2, 14, 14, row.Color)
			rl.DrawRectangleLines(legendX, legendY+2, 14, 14, rl.Fade(colorBorder, 0.8))
			drawText(row.Label, legendX+20, legendY, typeScale.Small, colorText)
			legendY += 20
		}
		modeLine := "Fog: off"
		if ui.run.Config.Mode == game.ModeAlone {
			modeLine = "Fog: on (permanent reveal)"
		}
		legendY += 8
		drawText(modeLine, legendX, legendY, typeScale.Small-1, colorMuted)
		legendY += 18
		drawText(fmt.Sprintf("Grid: %dx%d", topology.Width, topology.Height), legendX, legendY, typeScale.Small-1, colorMuted)
		legendY += 18
		drawText(fmt.Sprintf("Render: %dx detail + contours", topoRenderDetail), legendX, legendY, typeScale.Small-1, colorMuted)
	}
}

func (ui *gameUI) drawMiniMap(rect rl.Rectangle, withLegend bool) {
	if ui.run == nil {
		return
	}
	ui.run.EnsureTopology()
	topology := ui.run.Topology
	if topology.Width <= 0 || topology.Height <= 0 || len(topology.Cells) == 0 {
		return
	}
	// Reserve a bit more room for the footer so the cell text remains readable.
	area := rl.NewRectangle(rect.X+10, rect.Y+36, rect.Width-20, rect.Height-70)
	if area.Width < 20 || area.Height < 20 {
		return
	}
	posX, posY := ui.run.CurrentMapPosition()
	startX, startY, cols, rows := computeMiniMapWindow(topology.Width, topology.Height, posX, posY, miniMapViewCells)
	ui.drawTopologyRegion(area, startX, startY, cols, rows)
	if withLegend {
		legend := fmt.Sprintf("View %dx%d around player", cols, rows)
		drawText(legend, int32(rect.X+spaceM), int32(rect.Y+spaceS+18), typeScale.Small-1, colorMuted)
	}
	footer := fmt.Sprintf("Cell (%d,%d) | %.1fkm moved", posX, posY, ui.run.Travel.TotalKm)
	drawText(footer, int32(rect.X+spaceM), int32(rect.Y+rect.Height)-18, typeScale.Small, colorMuted)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SURVIVE_IT_DEBUG_WORLD")), "1") {
		debugLine := ui.run.CoherenceDebugLine()
		drawText(debugLine, int32(rect.X+spaceM), int32(rect.Y+rect.Height)-34, typeScale.Small-1, colorAccent)
	}
}

func (ui *gameUI) updateRunMap() {
	if ui.run == nil {
		ui.screen = screenRun
		return
	}
	if ShiftPressedKey(rl.KeyM) || rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
}

func (ui *gameUI) drawRunMap() {
	if ui.run == nil {
		return
	}
	ui.run.EnsureTopology()
	panel := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
	drawPanel(panel, "Topology Map")
	ui.drawTopologyMap(panel, true)
	DrawHintText("Shift+M or Esc to return", int32(panel.X+spaceM), int32(panel.Y+panel.Height)-24)
}
