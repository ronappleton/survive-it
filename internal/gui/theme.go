package gui

import (
	"fmt"

	uitheme "github.com/appengine-ltd/survive-it/internal/ui/theme"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Theme struct {
	Background      rl.Color
	Panel           rl.Color
	PanelRaised     rl.Color
	Border          rl.Color
	BorderStrong    rl.Color
	Divider         rl.Color
	TextPrimary     rl.Color
	TextSecondary   rl.Color
	TextMuted       rl.Color
	Accent          rl.Color
	AccentSecondary rl.Color
	AccentSoft      rl.Color
	Warning         rl.Color
	Danger          rl.Color
	DisabledPanel   rl.Color
	DisabledText    rl.Color
}

const (
	spaceXS = uitheme.PaddingXS
	spaceS  = uitheme.PaddingS
	spaceM  = uitheme.PaddingM
	spaceL  = uitheme.PaddingL
)

var AppTheme = Theme{
	Background:      uitheme.BG,
	Panel:           uitheme.Panel,
	PanelRaised:     uitheme.PanelRaised,
	Border:          uitheme.Border,
	BorderStrong:    uitheme.AccentForest,
	Divider:         uitheme.Divider,
	TextPrimary:     uitheme.TextPrimary,
	TextSecondary:   uitheme.TextSecondary,
	TextMuted:       uitheme.TextMuted,
	Accent:          uitheme.AccentEmber,
	AccentSecondary: uitheme.AccentForest,
	AccentSoft:      rl.Fade(uitheme.AccentForest, 0.8),
	Warning:         uitheme.WarningAmber,
	Danger:          uitheme.Danger,
	DisabledPanel:   uitheme.DisabledPanel,
	DisabledText:    uitheme.DisabledText,
}

// Compatibility aliases used by existing render code.
var (
	colorBG     = AppTheme.Background
	colorPanel  = AppTheme.Panel
	colorBorder = AppTheme.Border
	colorText   = AppTheme.TextPrimary
	colorDim    = AppTheme.TextSecondary
	colorMuted  = AppTheme.TextMuted
	colorAccent = AppTheme.Accent
	colorWarn   = AppTheme.Warning
	colorDanger = AppTheme.Danger
)

type TelemetryThresholds struct {
	Warning  int
	Danger   int
	Inverted bool
}

type PanelVariant = uitheme.PanelVariant

const (
	panelVariantDefault = uitheme.PanelStandard
	panelVariantRaised  = uitheme.PanelLifted
	panelVariantMuted   = uitheme.PanelMuted
)

type ButtonState = uitheme.ButtonState

const (
	buttonStateNormal   = uitheme.ButtonNormal
	buttonStateSelected = uitheme.ButtonSelected
	buttonStateFocused  = uitheme.ButtonFocused
	buttonStateDisabled = uitheme.ButtonDisabled
)

type ListItemState = uitheme.ListItemState

const (
	listStateNormal   = uitheme.ListItemNormal
	listStateSelected = uitheme.ListItemSelected
	listStateFocused  = uitheme.ListItemFocused
	listStateDisabled = uitheme.ListItemDisabled
)

// ---------------------------------------------------------------------------
// Frame
// ---------------------------------------------------------------------------

// DrawFrame paints the wood border around the entire window and returns the
// inner inset rectangle that all screen content should stay within.
func DrawFrame(screenW, screenH int32) rl.Rectangle {
	return uitheme.DrawFrame(screenW, screenH)
}

// ---------------------------------------------------------------------------
// Panel
// ---------------------------------------------------------------------------

// DrawPanel draws a themed panel. If title is non-empty, a header with an
// ember underline and a divider are drawn inside the panel top.
func DrawPanel(rect rl.Rectangle, title string, focused bool) {
	variant := panelVariantDefault
	if focused {
		variant = panelVariantRaised
	}
	uitheme.DrawPanel(rect, variant)
	if title != "" {
		DrawHeader(title, int32(rect.X+spaceM), int32(rect.Y+spaceS))
		dividerY := rect.Y + spaceS + float32(typeScale.Header) + 8
		DrawDivider(rect.X+spaceM, dividerY, rect.X+rect.Width-spaceM, dividerY)
	}
}

// ---------------------------------------------------------------------------
// Button
// ---------------------------------------------------------------------------

func DrawButton(rect rl.Rectangle, state ButtonState, text string) {
	uitheme.DrawButton(rect, state, text)
}

// ---------------------------------------------------------------------------
// List item
// ---------------------------------------------------------------------------

func DrawListItem(rect rl.Rectangle, state ListItemState, leftText, rightText string) {
	uitheme.DrawListItem(rect, state, leftText, rightText)
}

// ---------------------------------------------------------------------------
// Input field
// ---------------------------------------------------------------------------

// DrawInputField renders a styled text input field.
// text is the current buffer; placeholder is shown when text is empty and unfocused.
func DrawInputField(rect rl.Rectangle, text, placeholder string, focused bool) {
	uitheme.DrawInput(rect, text, placeholder, focused)
}

// ---------------------------------------------------------------------------
// Typography helpers
// ---------------------------------------------------------------------------

func DrawHeader(text string, x, y int32) {
	uitheme.DrawHeader(text, x, y)
}

func DrawDivider(x1, y1, x2, y2 float32) {
	uitheme.DrawDivider(x1, y1, x2, y2)
}

func DrawHintText(text string, x, y int32) {
	uitheme.DrawHintText(text, x, y)
}

func DrawLabelValue(label, value string, x, y int32, valueColor rl.Color) {
	drawText(label, x, y, typeScale.Body, AppTheme.TextSecondary)
	drawText(value, x+240, y, typeScale.Body, valueColor)
}

// ---------------------------------------------------------------------------
// Telemetry bar
// ---------------------------------------------------------------------------

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
	rl.DrawRectangleRec(track, rl.Fade(AppTheme.PanelRaised, 0.9))
	if fill.Width > 0 {
		rl.DrawRectangleRec(fill, telemetryFillColor(v, thresholds))
	}
	rl.DrawRectangleLinesEx(track, 1.0, rl.Fade(AppTheme.Border, 0.95))
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func listRowRect(x, y, width float32) rl.Rectangle {
	return rl.NewRectangle(x, y, width, uitheme.RowHeight)
}

func drawListRowFrame(rect rl.Rectangle, selected bool) {
	state := listStateNormal
	if selected {
		state = listStateSelected
	}
	DrawListItem(rect, state, "", "")
}

func drawDialogPanel(rect rl.Rectangle) {
	uitheme.DrawPanel(rect, panelVariantRaised)
	rl.DrawRectangleRoundedLinesEx(rect, uitheme.CornerRadius, uitheme.CornerSegments, 1.8, rl.Fade(AppTheme.Accent, 0.9))
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
		return AppTheme.AccentSecondary
	}
	if value <= danger {
		return AppTheme.Danger
	}
	if value <= warning {
		return AppTheme.Warning
	}
	return AppTheme.AccentSecondary
}
