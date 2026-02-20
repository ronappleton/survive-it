package theme

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Skin holds the loaded nine-slice and flat texture assets.
// All fields are safe to read after InitSkin(); zero-value fields produce
// a safe flat-colour fallback in DrawNineSlice.
var Skin skinAssets

type skinAssets struct {
	Frame  NineSlice // outer wood frame (full-screen border)
	Panel  NineSlice // interior panel background
	Button NineSlice // button face
	Input  NineSlice // text-input field

	loaded bool
}

// Sizing constants for the embedded placeholder textures.
// Real art can use any values here – just update these to match the new slice guides.
const (
	frameSlice  = int32(12)
	panelSlice  = int32(8)
	buttonSlice = int32(8)
	inputSlice  = int32(6)
)

// InitSkin loads ui-skin textures.  Call once after rl.InitWindow().
// If a texture file is missing the slot stays as zero-value (safe fallback).
func InitSkin() {
	if Skin.loaded {
		return
	}
	Skin.loaded = true
	Skin.Frame = loadNineSlice("assets/ui/frame_wood.png", frameSlice, frameSlice, frameSlice, frameSlice)
	Skin.Panel = loadNineSlice("assets/ui/panel_9slice.png", panelSlice, panelSlice, panelSlice, panelSlice)
	Skin.Button = loadNineSlice("assets/ui/button_9slice.png", buttonSlice, buttonSlice, buttonSlice, buttonSlice)
	Skin.Input = loadNineSlice("assets/ui/input_9slice.png", inputSlice, inputSlice, inputSlice, inputSlice)
}

// UnloadSkin releases GPU texture memory.  Call before rl.CloseWindow().
func UnloadSkin() {
	unloadTex(&Skin.Frame.Tex)
	unloadTex(&Skin.Panel.Tex)
	unloadTex(&Skin.Button.Tex)
	unloadTex(&Skin.Input.Tex)
	Skin.loaded = false
}

// FrameInset returns the inner safe area rectangle after the wood frame border.
func FrameInset(screenW, screenH int32) rl.Rectangle {
	m := float32(frameSlice)
	return rl.NewRectangle(m, m, float32(screenW)-m*2, float32(screenH)-m*2)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func loadNineSlice(path string, left, right, top, bottom int32) NineSlice {
	if _, err := os.Stat(path); err != nil {
		return NineSlice{Left: left, Right: right, Top: top, Bottom: bottom}
	}
	tex := rl.LoadTexture(path)
	if tex.ID == 0 {
		return NineSlice{Left: left, Right: right, Top: top, Bottom: bottom}
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	return NineSlice{Tex: tex, Left: left, Right: right, Top: top, Bottom: bottom}
}

func unloadTex(t *rl.Texture2D) {
	if t != nil && t.ID != 0 {
		rl.UnloadTexture(*t)
		*t = rl.Texture2D{}
	}
}
