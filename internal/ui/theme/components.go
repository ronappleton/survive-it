package theme

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	PaddingXS = float32(8)
	PaddingS  = float32(12)
	PaddingM  = float32(18)
	PaddingL  = float32(24)

	CornerRadius   = float32(0.08)
	CornerSegments = int32(8)

	BorderWidth      = float32(1.2)
	BorderWidthFocus = float32(2.0)
	RowHeight        = float32(40)
	ButtonHeight     = float32(56)
	AccentStripWidth = float32(4)
)

type PanelVariant int

const (
	PanelStandard PanelVariant = iota
	PanelLifted
	PanelMuted
)

type ButtonState int

const (
	ButtonNormal ButtonState = iota
	ButtonSelected
	ButtonFocused
	ButtonDisabled
)

type ListItemState int

const (
	ListItemNormal ListItemState = iota
	ListItemSelected
	ListItemFocused
	ListItemDisabled
)

// ---------------------------------------------------------------------------
// DrawFrame – full-screen wood border + slight vignette feel.
// Call first in each draw func; all content sits inside the returned inset rect.
// ---------------------------------------------------------------------------

func DrawFrame(screenW, screenH int32) rl.Rectangle {
	fullRect := rl.NewRectangle(0, 0, float32(screenW), float32(screenH))
	if Skin.Frame.Tex.ID != 0 {
		DrawNineSlice(Skin.Frame, fullRect, rl.White)
	} else {
		// Fallback: paint a narrow walnut border using solid rectangles.
		walnut := rl.NewColor(0x3E, 0x28, 0x12, 0xFF)
		bw := float32(frameSlice)
		rl.DrawRectangleRec(rl.NewRectangle(0, 0, float32(screenW), bw), walnut)
		rl.DrawRectangleRec(rl.NewRectangle(0, float32(screenH)-bw, float32(screenW), bw), walnut)
		rl.DrawRectangleRec(rl.NewRectangle(0, 0, bw, float32(screenH)), walnut)
		rl.DrawRectangleRec(rl.NewRectangle(float32(screenW)-bw, 0, bw, float32(screenH)), walnut)
	}
	return FrameInset(screenW, screenH)
}

// ---------------------------------------------------------------------------
// DrawPanel – textured panel background.
// ---------------------------------------------------------------------------

func DrawPanel(rect rl.Rectangle, variant PanelVariant) {
	fill := Panel
	stroke := Border
	strokeWidth := BorderWidth

	switch variant {
	case PanelLifted:
		fill = PanelRaised
		stroke = mix(Border, AccentForest, 0.35)
		strokeWidth = 1.4
	case PanelMuted:
		fill = DisabledPanel
		stroke = rl.Fade(Border, 0.75)
	}

	if Skin.Panel.Tex.ID != 0 {
		DrawNineSlice(Skin.Panel, rect, fill)
	} else {
		rl.DrawRectangleRounded(rect, CornerRadius, CornerSegments, fill)
	}
	rl.DrawRectangleRoundedLinesEx(rect, CornerRadius, CornerSegments, strokeWidth, stroke)

	// Inner highlight line for depth
	inner := rl.NewRectangle(rect.X+1, rect.Y+1, rect.Width-2, rect.Height-2)
	if inner.Width > 4 && inner.Height > 4 {
		rl.DrawRectangleRoundedLinesEx(inner, CornerRadius, CornerSegments, 1.0, rl.Fade(Divider, 0.65))
	}
}

// ---------------------------------------------------------------------------
// DrawButton – textured button face with survival state styles.
// ---------------------------------------------------------------------------

func DrawButton(rect rl.Rectangle, state ButtonState, text string) {
	fill := Panel
	fillTex := Skin.Button
	stroke := Border
	label := TextPrimary
	strokeWidth := BorderWidth

	switch state {
	case ButtonSelected, ButtonFocused:
		fill = PanelRaised
		stroke = AccentEmber
		strokeWidth = BorderWidthFocus
	case ButtonDisabled:
		fill = DisabledPanel
		stroke = rl.Fade(Border, 0.75)
		label = DisabledText
		fillTex = NineSlice{Left: buttonSlice, Right: buttonSlice, Top: buttonSlice, Bottom: buttonSlice} // empty/fallback
	}

	if fillTex.Tex.ID != 0 {
		DrawNineSlice(fillTex, rect, fill)
	} else {
		rl.DrawRectangleRounded(rect, CornerRadius, CornerSegments, fill)
	}
	rl.DrawRectangleRoundedLinesEx(rect, CornerRadius, CornerSegments, strokeWidth, stroke)

	// Ember glow strips for selected state
	if state == ButtonSelected || state == ButtonFocused {
		strip := rl.NewRectangle(rect.X+1, rect.Y+2, AccentStripWidth, rect.Height-4)
		if strip.Height > 0 {
			rl.DrawRectangleRec(strip, AccentEmber)
		}
	}

	if text == "" {
		return
	}
	size := Type.Body
	labelW := measureText(text, size)
	textX := int32(rect.X + (rect.Width-float32(labelW))/2)
	textY := int32(rect.Y + (rect.Height-float32(size))/2 - 1)
	drawText(text, textX, textY, size, label)
}

// ---------------------------------------------------------------------------
// DrawListItem – row item with optional accent strip.
// ---------------------------------------------------------------------------

func DrawListItem(rect rl.Rectangle, state ListItemState, leftText, rightText string) {
	fill := rl.Fade(PanelRaised, 0.45)
	stroke := rl.Fade(Border, 0.9)
	left := TextPrimary
	right := TextSecondary
	strokeWidth := BorderWidth
	strip := rl.Color{}
	drawStrip := false

	switch state {
	case ListItemSelected, ListItemFocused:
		fill = PanelRaised
		stroke = AccentEmber
		strokeWidth = BorderWidthFocus
		strip = AccentEmber
		drawStrip = true
		right = AccentEmber
	case ListItemDisabled:
		fill = DisabledPanel
		stroke = rl.Fade(Border, 0.75)
		left = DisabledText
		right = DisabledText
	}

	if Skin.Panel.Tex.ID != 0 {
		DrawNineSlice(Skin.Panel, rect, fill)
	} else {
		rl.DrawRectangleRounded(rect, CornerRadius, CornerSegments, fill)
	}
	rl.DrawRectangleRoundedLinesEx(rect, CornerRadius, CornerSegments, strokeWidth, stroke)

	if drawStrip {
		stripRect := rl.NewRectangle(rect.X+1, rect.Y+2, AccentStripWidth, rect.Height-4)
		if stripRect.Height > 0 {
			rl.DrawRectangleRec(stripRect, strip)
		}
	}

	if leftText != "" {
		drawText(leftText, int32(rect.X+PaddingM), int32(rect.Y+10), Type.Body, left)
	}
	if rightText != "" {
		rightW := measureText(rightText, Type.Body)
		rightX := int32(rect.X + rect.Width - PaddingM - float32(rightW))
		drawText(rightText, rightX, int32(rect.Y+10), Type.Body, right)
	}
}

// ---------------------------------------------------------------------------
// DrawInput – text field with inset tray look.
// ---------------------------------------------------------------------------

func DrawInput(rect rl.Rectangle, text, placeholder string, focused bool) {
	fill := rl.NewColor(0x10, 0x16, 0x1A, 0xFF)
	stroke := Border
	strokeWidth := BorderWidth
	if focused {
		stroke = AccentEmber
		strokeWidth = BorderWidthFocus
	}

	if Skin.Input.Tex.ID != 0 {
		DrawNineSlice(Skin.Input, rect, fill)
	} else {
		rl.DrawRectangleRounded(rect, CornerRadius, CornerSegments, fill)
	}
	rl.DrawRectangleRoundedLinesEx(rect, CornerRadius, CornerSegments, strokeWidth, stroke)

	display := text
	if display == "" && !focused {
		display = placeholder
	}
	clr := TextPrimary
	if display == placeholder || display == "" {
		clr = TextMuted
	}
	if display != "" {
		cursor := ""
		if focused {
			cursor = "_"
		}
		drawText(display+cursor, int32(rect.X+PaddingM), int32(rect.Y+float32(Type.Body)*0.38), Type.Body, clr)
	} else if focused {
		drawText("_", int32(rect.X+PaddingM), int32(rect.Y+float32(Type.Body)*0.38), Type.Body, TextPrimary)
	}
}

// ---------------------------------------------------------------------------
// DrawHeader – section title with ember underline.
// ---------------------------------------------------------------------------

func DrawHeader(text string, x, y int32) {
	if text == "" {
		return
	}
	drawText(text, x, y, Type.Header, TextPrimary)
	w := measureText(text, Type.Header)
	lineW := int32(float32(w) * 0.6)
	if lineW < 44 {
		lineW = 44
	}
	drawLine(float32(x), float32(y+Type.Header+6), float32(x+lineW), float32(y+Type.Header+6), 2.0, AccentEmber)
}

// ---------------------------------------------------------------------------
// DrawDivider – subtle horizontal / vertical divider line.
// ---------------------------------------------------------------------------

func DrawDivider(x1, y1, x2, y2 float32) {
	drawLine(x1, y1, x2, y2, 1.0, rl.Fade(Divider, 0.95))
}

// ---------------------------------------------------------------------------
// DrawHintText – muted small hint at bottom of panels.
// ---------------------------------------------------------------------------

func DrawHintText(text string, x, y int32) {
	if text == "" {
		return
	}
	drawText(text, x, y, Type.Small, TextMuted)
}

// ---------------------------------------------------------------------------
// private helpers
// ---------------------------------------------------------------------------

func drawLine(x1, y1, x2, y2, thickness float32, clr rl.Color) {
	rl.DrawLineEx(rl.NewVector2(x1, y1), rl.NewVector2(x2, y2), thickness, clr)
}

func mix(a, b rl.Color, t float32) rl.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	inv := 1.0 - t
	return rl.NewColor(
		uint8(float32(a.R)*inv+float32(b.R)*t),
		uint8(float32(a.G)*inv+float32(b.G)*t),
		uint8(float32(a.B)*inv+float32(b.B)*t),
		uint8(float32(a.A)*inv+float32(b.A)*t),
	)
}
