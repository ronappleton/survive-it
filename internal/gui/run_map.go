package gui

import (
	"fmt"
	"math"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	runLogSplitRatio = 0.66
	runLayoutPadding = 16
	runLayoutGap     = 10
	miniMapViewCells = 31
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
		return rl.NewColor(71, 106, 88, 255)
	case game.TopoBiomeGrassland:
		return rl.NewColor(116, 136, 87, 255)
	case game.TopoBiomeJungle:
		return rl.NewColor(56, 98, 80, 255)
	case game.TopoBiomeWetland:
		return rl.NewColor(76, 104, 102, 255)
	case game.TopoBiomeSwamp:
		return rl.NewColor(69, 92, 82, 255)
	case game.TopoBiomeDesert:
		return rl.NewColor(154, 136, 92, 255)
	case game.TopoBiomeMountain:
		return rl.NewColor(130, 133, 145, 255)
	case game.TopoBiomeTundra:
		return rl.NewColor(151, 163, 174, 255)
	case game.TopoBiomeBoreal:
		return rl.NewColor(74, 110, 97, 255)
	default:
		return rl.NewColor(96, 105, 110, 255)
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
	geo, ok := computeSquareGridGeometry(area, cols, rows)
	if !ok {
		return
	}
	for y := 0; y < rows; y++ {
		worldY := startY + y
		for x := 0; x < cols; x++ {
			worldX := startX + x
			idx := worldY*topology.Width + worldX
			cell := topology.Cells[idx]
			clr := shadeByElevation(topoBiomeColor(cell.Biome), cell.Elevation)
			if cell.Flags&game.TopoFlagWater != 0 {
				clr = rl.NewColor(76, 116, 156, 255)
			}
			if cell.Flags&game.TopoFlagRiver != 0 {
				clr = rl.NewColor(95, 141, 185, 255)
			}
			if cell.Flags&game.TopoFlagLake != 0 {
				clr = rl.NewColor(88, 133, 176, 255)
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
				clr = rl.NewColor(20, 24, 30, 255)
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
	rl.DrawRectangleLinesEx(geo.DrawRect, 1.0, rl.Fade(colorBorder, 0.8))

	px, py := ui.run.CurrentMapPosition()
	if px >= startX && px < startX+cols && py >= startY && py < startY+rows {
		localX := px - startX
		localY := py - startY
		cx := geo.OriginX + (float32(localX)+0.5)*geo.CellSize
		cy := geo.OriginY + (float32(localY)+0.5)*geo.CellSize
		r := float32(3)
		if geo.CellSize > 6 {
			r = geo.CellSize * 0.35
		}
		if r < 2 {
			r = 2
		}
		rl.DrawCircle(int32(cx), int32(cy), r, rl.NewColor(255, 88, 88, 255))
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
			{Label: "Water/River", Color: rl.NewColor(88, 133, 176, 255)},
			{Label: "Player", Color: rl.NewColor(255, 88, 88, 255)},
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
		drawText(modeLine, legendX, legendY, typeScale.Small-1, colorDim)
		legendY += 18
		drawText(fmt.Sprintf("Grid: %dx%d", topology.Width, topology.Height), legendX, legendY, typeScale.Small-1, colorDim)
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
		drawText(legend, int32(rect.X+spaceM), int32(rect.Y+spaceS+18), typeScale.Small-1, colorDim)
	}
	footer := fmt.Sprintf("Cell (%d,%d) | %.1fkm moved", posX, posY, ui.run.Travel.TotalKm)
	drawText(footer, int32(rect.X+spaceM), int32(rect.Y+rect.Height)-18, typeScale.Small, colorDim)
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
	drawText("Shift+M or Esc to return", int32(panel.X+spaceM), int32(panel.Y+panel.Height)-24, typeScale.Small, colorDim)
}
