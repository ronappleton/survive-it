package gui

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestComputeMiniMapWindowClampsAtBounds(t *testing.T) {
	startX, startY, cols, rows := computeMiniMapWindow(100, 80, 50, 40, 21)
	if cols != 21 || rows != 21 {
		t.Fatalf("expected 21x21 window, got %dx%d", cols, rows)
	}
	if startX != 40 || startY != 30 {
		t.Fatalf("expected centered window start 40,30 got %d,%d", startX, startY)
	}

	startX, startY, _, _ = computeMiniMapWindow(100, 80, 1, 2, 21)
	if startX != 0 || startY != 0 {
		t.Fatalf("expected clamped start at origin, got %d,%d", startX, startY)
	}

	startX, startY, _, _ = computeMiniMapWindow(100, 80, 99, 79, 21)
	if startX != 79 || startY != 59 {
		t.Fatalf("expected clamped end window start 79,59 got %d,%d", startX, startY)
	}
}

func TestComputeSquareGridGeometryUsesMinDimension(t *testing.T) {
	area := rl.NewRectangle(0, 0, 300, 180)
	geo, ok := computeSquareGridGeometry(area, 20, 20)
	if !ok {
		t.Fatalf("expected valid geometry")
	}
	if math.Abs(float64(geo.CellSize-9.0)) > 0.0001 {
		t.Fatalf("expected cell size 9, got %.4f", geo.CellSize)
	}
	if math.Abs(float64(geo.OriginX-60.0)) > 0.0001 {
		t.Fatalf("expected centered originX 60, got %.4f", geo.OriginX)
	}
	if math.Abs(float64(geo.OriginY-0.0)) > 0.0001 {
		t.Fatalf("expected centered originY 0, got %.4f", geo.OriginY)
	}
}
