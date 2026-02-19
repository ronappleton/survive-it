package gui

import (
	"fmt"
	"math"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	runLogSplitRatio = 0.66
	runLayoutPadding = 20
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
	topH := float32(194)
	inputH := float32(92)
	if outer.Height < 520 {
		topH = 168
		inputH = 82
	}
	middleTop := outer.Y + topH
	inputTop := outer.Y + outer.Height - inputH
	if inputTop-middleTop < 160 {
		inputTop = middleTop + 160
	}
	middleH := inputTop - middleTop
	splitX := outer.X + outer.Width*runLogSplitRatio
	return runLayout{
		Outer:       outer,
		TopRect:     rl.NewRectangle(outer.X, outer.Y, outer.Width, topH),
		LogRect:     rl.NewRectangle(outer.X, middleTop, splitX-outer.X, middleH),
		MiniMapRect: rl.NewRectangle(splitX, middleTop, outer.X+outer.Width-splitX, middleH),
		InputRect:   rl.NewRectangle(outer.X, inputTop, outer.Width, outer.Y+outer.Height-inputTop),
	}
}

func drawRunMessageLog(rect rl.Rectangle, messages []string) {
	maxWidth := int32(rect.Width) - 28
	lineHeight := int32(22)
	maxLines := int((rect.Height - 52) / float32(lineHeight))
	if maxLines < 4 {
		maxLines = 4
	}
	flattened := make([]string, 0, maxLines)
	for i := len(messages) - 1; i >= 0 && len(flattened) < maxLines; i-- {
		lines := wrapText(messages[i], 17, maxWidth)
		for j := len(lines) - 1; j >= 0 && len(flattened) < maxLines; j-- {
			flattened = append(flattened, lines[j])
		}
	}
	y := int32(rect.Y+rect.Height) - 26
	for i := 0; i < len(flattened); i++ {
		line := flattened[i]
		rl.DrawText(line, int32(rect.X)+12, y-lineHeight, 17, colorText)
		y -= lineHeight
		if y < int32(rect.Y)+26 {
			break
		}
	}
}

func topoBiomeColor(b uint8) rl.Color {
	switch b {
	case game.TopoBiomeForest:
		return rl.NewColor(40, 120, 64, 255)
	case game.TopoBiomeGrassland:
		return rl.NewColor(104, 144, 72, 255)
	case game.TopoBiomeJungle:
		return rl.NewColor(24, 110, 58, 255)
	case game.TopoBiomeWetland:
		return rl.NewColor(70, 118, 102, 255)
	case game.TopoBiomeSwamp:
		return rl.NewColor(60, 92, 70, 255)
	case game.TopoBiomeDesert:
		return rl.NewColor(178, 154, 88, 255)
	case game.TopoBiomeMountain:
		return rl.NewColor(128, 128, 138, 255)
	case game.TopoBiomeTundra:
		return rl.NewColor(158, 172, 178, 255)
	case game.TopoBiomeBoreal:
		return rl.NewColor(66, 112, 88, 255)
	default:
		return rl.NewColor(80, 96, 96, 255)
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

	cellW := area.Width / float32(topology.Width)
	cellH := area.Height / float32(topology.Height)
	for y := 0; y < topology.Height; y++ {
		for x := 0; x < topology.Width; x++ {
			idx := y*topology.Width + x
			cell := topology.Cells[idx]
			clr := shadeByElevation(topoBiomeColor(cell.Biome), cell.Elevation)
			if cell.Flags&game.TopoFlagWater != 0 {
				clr = rl.NewColor(48, 104, 184, 255)
			}
			if cell.Flags&game.TopoFlagRiver != 0 {
				clr = rl.NewColor(66, 140, 210, 255)
			}
			if cell.Flags&game.TopoFlagLake != 0 {
				clr = rl.NewColor(58, 124, 198, 255)
			}
			if cell.Flags&game.TopoFlagCoast != 0 {
				clr = rl.NewColor(
					uint8(min(255, int(clr.R)+18)),
					uint8(min(255, int(clr.G)+18)),
					uint8(min(255, int(clr.B)+10)),
					255,
				)
			}
			if ui.run.Config.Mode == game.ModeAlone && !ui.run.IsRevealed(x, y) {
				clr = rl.NewColor(14, 18, 24, 255)
			}
			x0 := int32(area.X + float32(x)*cellW)
			y0 := int32(area.Y + float32(y)*cellH)
			x1 := int32(area.X + float32(x+1)*cellW)
			y1 := int32(area.Y + float32(y+1)*cellH)
			w := max(1, int(x1-x0))
			h := max(1, int(y1-y0))
			rl.DrawRectangle(x0, y0, int32(w), int32(h), clr)
		}
	}
	rl.DrawRectangleLinesEx(area, 1.2, rl.Fade(colorBorder, 0.8))

	px, py := ui.run.CurrentMapPosition()
	if px >= 0 && py >= 0 && px < topology.Width && py < topology.Height {
		cx := area.X + (float32(px)+0.5)*cellW
		cy := area.Y + (float32(py)+0.5)*cellH
		r := float32(3)
		if cellW > 6 && cellH > 6 {
			r = float32(math.Min(float64(cellW), float64(cellH))) * 0.35
		}
		if r < 2 {
			r = 2
		}
		rl.DrawCircle(int32(cx), int32(cy), r, rl.NewColor(255, 88, 88, 255))
	}

	if withLegend {
		legendX := int32(rect.X + rect.Width - 162)
		legendY := int32(rect.Y + 40)
		rl.DrawText("Legend", legendX, legendY, 18, colorAccent)
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
			{Label: "Water/River", Color: rl.NewColor(58, 124, 198, 255)},
			{Label: "Player", Color: rl.NewColor(255, 88, 88, 255)},
		}
		for _, row := range legendRows {
			rl.DrawRectangle(legendX, legendY+2, 14, 14, row.Color)
			rl.DrawRectangleLines(legendX, legendY+2, 14, 14, rl.Fade(colorBorder, 0.8))
			rl.DrawText(row.Label, legendX+20, legendY, 15, colorText)
			legendY += 20
		}
		modeLine := "Fog: off"
		if ui.run.Config.Mode == game.ModeAlone {
			modeLine = "Fog: on (permanent reveal)"
		}
		legendY += 8
		rl.DrawText(modeLine, legendX, legendY, 14, colorDim)
		legendY += 18
		rl.DrawText(fmt.Sprintf("Grid: %dx%d", topology.Width, topology.Height), legendX, legendY, 14, colorDim)
	}
}

func (ui *gameUI) drawMiniMap(rect rl.Rectangle, withLegend bool) {
	ui.drawTopologyMap(rect, withLegend)
	if ui.run == nil {
		return
	}
	posX, posY := ui.run.CurrentMapPosition()
	footer := fmt.Sprintf("Cell (%d,%d) | %.1fkm moved", posX, posY, ui.run.Travel.TotalKm)
	rl.DrawText(footer, int32(rect.X)+12, int32(rect.Y+rect.Height)-20, 15, colorDim)
}

func (ui *gameUI) updateRunMap() {
	if ui.run == nil {
		ui.screen = screenRun
		return
	}
	if rl.IsKeyPressed(rl.KeyM) || rl.IsKeyPressed(rl.KeyEscape) {
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
	rl.DrawText("M or Esc to return", int32(panel.X)+14, int32(panel.Y+panel.Height)-24, 17, colorDim)
}
