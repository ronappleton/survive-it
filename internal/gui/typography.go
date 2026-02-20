package gui

import (
	"math"
	"os"
	"path/filepath"

	uitheme "github.com/appengine-ltd/survive-it/internal/ui/theme"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type typographyScale struct {
	Title  int32
	Header int32
	Body   int32
	Small  int32
	Log    int32
}

type typographyState struct {
	base       rl.Font
	log        rl.Font
	ownsBase   bool
	ownsLog    bool
	lineFactor float32
}

var (
	typeScale = typographyScale{
		Title:  30,
		Header: 21,
		Body:   19,
		Small:  16,
		Log:    18,
	}
	uiType = typographyState{lineFactor: 1.34}
)

func initTypography() {
	uiType.base = rl.GetFontDefault()
	uiType.log = uiType.base

	fontCandidates := []string{
		filepath.Join("assets", "fonts", "Inter-Regular.ttf"),
		filepath.Join("assets", "fonts", "IBMPlexSans-Regular.ttf"),
		filepath.Join("assets", "fonts", "NotoSans-Regular.ttf"),
	}
	if f, ok := loadFontFromCandidates(fontCandidates, 36); ok {
		uiType.base = f
		uiType.log = f
		uiType.ownsBase = true
		uiType.ownsLog = true
	}

	rl.SetTextureFilter(uiType.base.Texture, rl.FilterBilinear)
	uitheme.SetTextRenderer(drawText, measureText)
}

func shutdownTypography() {
	if uiType.ownsLog && uiType.log.Texture.ID != 0 {
		rl.UnloadFont(uiType.log)
	}
	if uiType.ownsBase && uiType.base.Texture.ID != 0 && uiType.base.Texture.ID != uiType.log.Texture.ID {
		rl.UnloadFont(uiType.base)
	}
	uiType = typographyState{lineFactor: 1.34}
}

func loadFontFromCandidates(candidates []string, fontSize int32) (rl.Font, bool) {
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		font := rl.LoadFontEx(path, fontSize, nil, 0)
		if font.Texture.ID == 0 {
			continue
		}
		return font, true
	}
	return rl.Font{}, false
}

func drawText(text string, x, y, fontSize int32, clr rl.Color) {
	if uiType.base.Texture.ID == 0 {
		rl.DrawText(text, x, y, fontSize, clr)
		return
	}
	rl.DrawTextEx(uiType.base, text, rl.Vector2{X: float32(x), Y: float32(y)}, float32(fontSize), 1, clr)
}

func drawLogText(text string, x, y, fontSize int32, clr rl.Color) {
	if uiType.log.Texture.ID == 0 {
		rl.DrawText(text, x, y, fontSize, clr)
		return
	}
	rl.DrawTextEx(uiType.log, text, rl.Vector2{X: float32(x), Y: float32(y)}, float32(fontSize), 1, clr)
}

func measureText(text string, fontSize int32) int32 {
	if uiType.base.Texture.ID == 0 {
		return int32(rl.MeasureText(text, fontSize))
	}
	return int32(math.Round(float64(rl.MeasureTextEx(uiType.base, text, float32(fontSize), 1).X)))
}

func textLineHeight(size int32) int32 {
	if size < 1 {
		size = 1
	}
	return int32(math.Round(float64(size) * float64(uiType.lineFactor)))
}
