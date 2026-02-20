package theme

import rl "github.com/gen2brain/raylib-go/raylib"

// TextDrawFunc renders text with caller-provided font handling.
type TextDrawFunc func(text string, x, y, fontSize int32, clr rl.Color)

// TextMeasureFunc reports text width in pixels for the active font.
type TextMeasureFunc func(text string, fontSize int32) int32

var (
	textDrawFn TextDrawFunc = func(text string, x, y, fontSize int32, clr rl.Color) {
		rl.DrawText(text, x, y, fontSize, clr)
	}
	textMeasureFn TextMeasureFunc = func(text string, fontSize int32) int32 {
		return int32(rl.MeasureText(text, fontSize))
	}
)

// SetTextRenderer wires theme helpers to the GUI text system.
func SetTextRenderer(draw TextDrawFunc, measure TextMeasureFunc) {
	if draw != nil {
		textDrawFn = draw
	}
	if measure != nil {
		textMeasureFn = measure
	}
}

func drawText(text string, x, y, fontSize int32, clr rl.Color) {
	textDrawFn(text, x, y, fontSize, clr)
}

func measureText(text string, fontSize int32) int32 {
	return textMeasureFn(text, fontSize)
}
