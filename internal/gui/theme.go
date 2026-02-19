package gui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Theme struct {
	Background    rl.Color
	Panel         rl.Color
	Border        rl.Color
	BorderStrong  rl.Color
	TextPrimary   rl.Color
	TextSecondary rl.Color
	Accent        rl.Color
	AccentSoft    rl.Color
	Warning       rl.Color
	Danger        rl.Color
}

const (
	spaceXS = float32(6)
	spaceS  = float32(10)
	spaceM  = float32(14)
	spaceL  = float32(20)

	panelRadius      = float32(0.08)
	panelSegments    = int32(8)
	panelBorderThin  = float32(1.0)
	panelBorderThick = float32(1.8)
)

var AppTheme = Theme{
	Background:    rl.NewColor(15, 19, 26, 255),
	Panel:         rl.NewColor(25, 31, 40, 244),
	Border:        rl.NewColor(74, 92, 108, 220),
	BorderStrong:  rl.NewColor(110, 137, 162, 240),
	TextPrimary:   rl.NewColor(224, 232, 238, 255),
	TextSecondary: rl.NewColor(156, 171, 184, 255),
	Accent:        rl.NewColor(118, 167, 204, 255),
	AccentSoft:    rl.NewColor(88, 123, 152, 222),
	Warning:       rl.NewColor(218, 167, 92, 255),
	Danger:        rl.NewColor(198, 92, 96, 255),
}

// Compatibility aliases used by existing render code.
var (
	colorBG     = AppTheme.Background
	colorPanel  = AppTheme.Panel
	colorBorder = AppTheme.Border
	colorText   = AppTheme.TextPrimary
	colorDim    = AppTheme.TextSecondary
	colorAccent = AppTheme.Accent
	colorWarn   = AppTheme.Warning
)

type TelemetryThresholds struct {
	Warning  int
	Danger   int
	Inverted bool
}

func DrawPanel(rect rl.Rectangle, title string, focused bool) {
	rl.DrawRectangleRounded(rect, panelRadius, panelSegments, AppTheme.Panel)
	border := AppTheme.Border
	thickness := panelBorderThin
	if focused {
		border = AppTheme.BorderStrong
		thickness = panelBorderThick
	}
	rl.DrawRectangleRoundedLinesEx(rect, panelRadius, panelSegments, thickness, border)
	if title != "" {
		drawText(title, int32(rect.X+spaceM), int32(rect.Y+spaceS), typeScale.Header, AppTheme.TextPrimary)
		DrawDivider(rect.X+spaceM, rect.Y+spaceS+float32(typeScale.Header)+3, rect.X+rect.Width-spaceM, rect.Y+spaceS+float32(typeScale.Header)+3)
	}
}

func DrawDivider(x1, y1, x2, y2 float32) {
	drawUILine(x1, y1, x2, y2, 1.0, rl.Fade(AppTheme.Border, 0.7))
}

func DrawLabelValue(label, value string, x, y int32, valueColor rl.Color) {
	drawText(label, x, y, typeScale.Body, AppTheme.TextSecondary)
	drawText(value, x+240, y, typeScale.Body, valueColor)
}

func DrawTelemetryBar(label string, value int, rect rl.Rectangle, thresholds TelemetryThresholds) {
	v := clampInt(value, 0, 100)
	barHeight := float32(8)
	if rect.Height >= 6 && rect.Height <= 10 {
		barHeight = rect.Height
	}
	if rect.Height > 10 {
		barHeight = 8
	}
	labelY := int32(rect.Y)
	barY := rect.Y + float32(typeScale.Small) + 2
	track := rl.NewRectangle(rect.X, barY, rect.Width, barHeight)
	fill := rl.NewRectangle(track.X+1, track.Y+1, (track.Width-2)*float32(v)/100.0, track.Height-2)

	drawText(fmt.Sprintf("%s %d%%", label, v), int32(rect.X), labelY, typeScale.Small, AppTheme.TextSecondary)
	rl.DrawRectangleRec(track, rl.Fade(AppTheme.TextSecondary, 0.18))
	if fill.Width > 0 {
		rl.DrawRectangleRec(fill, telemetryFillColor(v, thresholds))
	}
	rl.DrawRectangleLinesEx(track, 1.0, rl.Fade(AppTheme.Border, 0.9))
}

func telemetryFillColor(value int, thresholds TelemetryThresholds) rl.Color {
	warning := clampInt(thresholds.Warning, 0, 100)
	danger := clampInt(thresholds.Danger, 0, 100)
	if warning == 0 {
		warning = 35
	}
	if danger == 0 {
		danger = 20
	}
	if thresholds.Inverted {
		if value >= danger {
			return AppTheme.Danger
		}
		if value >= warning {
			return AppTheme.Warning
		}
		return AppTheme.AccentSoft
	}
	if value <= danger {
		return AppTheme.Danger
	}
	if value <= warning {
		return AppTheme.Warning
	}
	return AppTheme.AccentSoft
}
