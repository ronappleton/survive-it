//go:build ignore

// gen_ui_placeholders.go – run with:
//
//	go run scripts/gen_ui_placeholders.go
//
// Creates assets/ui/*.png placeholder textures for the survive-it UI skin.
// Each file is a small PNG with a coloured border that makes the 9-slice
// corners clearly visible.  Replace with real art at any time – the 9-slice
// slice constants are in internal/ui/theme/textures.go.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if err := os.MkdirAll(filepath.Join("assets", "ui"), 0o755); err != nil {
		log.Fatal(err)
	}

	// frame_wood.png  – 64×64, slice=12
	// Walnut-brown frame border, very dark centre (the inner playfield BG).
	genTexture("assets/ui/frame_wood.png", 64, 64, 12,
		color.RGBA{0x3E, 0x28, 0x12, 0xFF}, // border: dark walnut
		color.RGBA{0x14, 0x1A, 0x1F, 0xFF}, // centre: app background
	)

	// panel_9slice.png – 48×48, slice=8
	// Dark slate interior with a subtle steel border.
	genTexture("assets/ui/panel_9slice.png", 48, 48, 8,
		color.RGBA{0x2E, 0x3A, 0x40, 0xFF}, // border: steel
		color.RGBA{0x1C, 0x23, 0x29, 0xFF}, // centre: panel dark
	)

	// button_9slice.png – 32×32, slice=8
	// Warmer wood-plate feel; ember accent on the border already.
	genTexture("assets/ui/button_9slice.png", 32, 32, 8,
		color.RGBA{0x5C, 0x38, 0x18, 0xFF}, // border: warm wood
		color.RGBA{0x21, 0x2A, 0x31, 0xFF}, // centre: panel raised
	)

	// input_9slice.png – 24×24, slice=6
	// Slightly inset infield tray.
	genTexture("assets/ui/input_9slice.png", 24, 24, 6,
		color.RGBA{0x2E, 0x3A, 0x40, 0xFF}, // border: steel
		color.RGBA{0x10, 0x16, 0x1A, 0xFF}, // centre: near-black tray
	)

	log.Println("Placeholder textures written to assets/ui/")
}

// genTexture writes a PNG of size w×h.
// The outer 'slice' pixels on all four sides are coloured 'border'.
// The remaining centre is coloured 'centre'.
func genTexture(path string, w, h, slice int, border, centre color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < slice || y < slice || x >= w-slice || y >= h-slice {
				img.SetRGBA(x, y, border)
			} else {
				img.SetRGBA(x, y, centre)
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatalf("encode %s: %v", path, err)
	}
	log.Printf("  wrote %s (%dx%d slice=%d)", path, w, h, slice)
}
