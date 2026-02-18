package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
)

func renderPlayerBodyANSI(p game.PlayerConfig, widthChars, heightRows int) string {
	if widthChars < 12 || heightRows < 10 {
		return playerASCIIArt(p)
	}

	widthChars = clampInt(widthChars, 12, 48)
	heightRows = clampInt(heightRows, 10, 72)
	targetW := widthChars
	targetH := heightRows * 2

	// Supersample and downscale for smoother edges in terminal half-block output.
	scale := 4
	hiW := targetW * scale
	hiH := targetH * scale
	dc := gg.NewContext(hiW, hiH)
	drawPlayerFigure(dc, p)

	scaled := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), dc.Image(), dc.Image().Bounds(), xdraw.Over, nil)

	return rgbaImageToANSIHalfBlocks(scaled)
}

func drawPlayerFigure(dc *gg.Context, p game.PlayerConfig) {
	w := float64(dc.Width())
	h := float64(dc.Height())

	heightInches := float64((p.HeightFt * 12) + p.HeightIn)
	if heightInches <= 0 {
		heightInches = 70
	}
	heightNorm := clampFloat((heightInches-56.0)/28.0, 0.0, 1.0)

	weight := float64(p.WeightKg)
	if weight <= 0 {
		weight = 75
	}
	weightNorm := clampFloat((weight-45.0)/95.0, 0.0, 1.0)

	// Background: dark gradient panel + subtle glow.
	bgGrad := gg.NewLinearGradient(0, 0, 0, h)
	bgGrad.AddColorStop(0.0, color.RGBA{R: 3, G: 18, B: 9, A: 255})
	bgGrad.AddColorStop(1.0, color.RGBA{R: 1, G: 10, B: 5, A: 255})
	dc.SetFillStyle(bgGrad)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()

	figureHeight := lerp(0.60, 0.90, heightNorm) * h
	feetY := h * 0.93
	topY := feetY - figureHeight
	cx := w * 0.5

	shoulderBase := lerp(0.14, 0.22, weightNorm) * w
	hipBase := shoulderBase * 0.82
	waistBase := shoulderBase * 0.66
	switch p.BodyType {
	case game.BodyTypeMale:
		shoulderBase *= 1.12
		hipBase *= 0.94
	case game.BodyTypeFemale:
		shoulderBase *= 0.95
		hipBase *= 1.12
		waistBase *= 0.92
	case game.BodyTypeNeutral:
		waistBase *= 0.97
	}
	if shoulderBase < 28 {
		shoulderBase = 28
	}
	if hipBase < 24 {
		hipBase = 24
	}
	waistBase = clampFloat(waistBase, 18, shoulderBase*0.95)

	headR := figureHeight * 0.09
	headCY := topY + headR
	neckY := headCY + headR*1.05
	shoulderY := neckY + figureHeight*0.03
	torsoHeight := figureHeight * 0.34
	hipY := shoulderY + torsoHeight
	legLength := feetY - hipY

	armLen := figureHeight * 0.31
	elbowY := shoulderY + armLen*0.55
	handY := shoulderY + armLen
	armSpread := shoulderBase*0.78 + lerp(7.0, 18.0, weightNorm)

	legSpread := hipBase * 0.38
	leftFootX := cx - legSpread
	rightFootX := cx + legSpread

	primary := color.RGBA{R: 35, G: 225, B: 125, A: 255}
	shade := color.RGBA{R: 18, G: 130, B: 70, A: 255}
	highlight := color.RGBA{R: 125, G: 255, B: 180, A: 255}
	outline := color.RGBA{R: 5, G: 42, B: 20, A: 210}

	dc.SetRGBA(0.05, 0.85, 0.35, 0.14)
	dc.DrawEllipse(cx, shoulderY+torsoHeight*0.9, shoulderBase*1.95, figureHeight*0.58)
	dc.Fill()
	dc.SetRGBA(0.03, 0.52, 0.20, 0.30)
	dc.DrawEllipse(cx, feetY+8, shoulderBase*0.95, 9)
	dc.Fill()

	headGrad := gg.NewRadialGradient(cx-headR*0.26, headCY-headR*0.34, headR*0.15, cx, headCY, headR*1.2)
	headGrad.AddColorStop(0.0, highlight)
	headGrad.AddColorStop(1.0, primary)
	dc.SetFillStyle(headGrad)
	dc.DrawCircle(cx, headCY, headR)
	dc.Fill()
	dc.SetColor(outline)
	dc.SetLineWidth(2.2)
	dc.DrawCircle(cx, headCY, headR)
	dc.Stroke()

	torsoGrad := gg.NewLinearGradient(cx, shoulderY, cx, hipY)
	torsoGrad.AddColorStop(0.0, primary)
	torsoGrad.AddColorStop(1.0, shade)
	dc.SetFillStyle(torsoGrad)
	dc.MoveTo(cx-shoulderBase*0.5, shoulderY)
	dc.LineTo(cx+shoulderBase*0.5, shoulderY)
	dc.LineTo(cx+waistBase*0.58, shoulderY+torsoHeight*0.62)
	dc.LineTo(cx+hipBase*0.5, hipY)
	dc.LineTo(cx-hipBase*0.5, hipY)
	dc.LineTo(cx-waistBase*0.58, shoulderY+torsoHeight*0.62)
	dc.ClosePath()
	dc.Fill()
	dc.SetColor(outline)
	dc.SetLineWidth(2.5)
	dc.Stroke()

	limbThickness := lerp(8.0, 13.0, weightNorm)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	dc.SetLineWidth(limbThickness)
	dc.SetColor(primary)

	leftShoulderX := cx - shoulderBase*0.47
	rightShoulderX := cx + shoulderBase*0.47
	dc.DrawLine(leftShoulderX, shoulderY+8, leftShoulderX-armSpread*0.48, elbowY)
	dc.Stroke()
	dc.DrawLine(leftShoulderX-armSpread*0.48, elbowY, leftShoulderX-armSpread, handY)
	dc.Stroke()
	dc.DrawLine(rightShoulderX, shoulderY+8, rightShoulderX+armSpread*0.48, elbowY)
	dc.Stroke()
	dc.DrawLine(rightShoulderX+armSpread*0.48, elbowY, rightShoulderX+armSpread, handY)
	dc.Stroke()

	leftHipX := cx - hipBase*0.27
	rightHipX := cx + hipBase*0.27
	kneeDrop := legLength * 0.5
	dc.DrawLine(leftHipX, hipY, leftHipX-3, hipY+kneeDrop)
	dc.Stroke()
	dc.DrawLine(leftHipX-3, hipY+kneeDrop, leftFootX, feetY)
	dc.Stroke()

	dc.DrawLine(rightHipX, hipY, rightHipX+3, hipY+kneeDrop)
	dc.Stroke()
	dc.DrawLine(rightHipX+3, hipY+kneeDrop, rightFootX, feetY)
	dc.Stroke()

	dc.SetRGBA(0.92, 1, 0.92, 0.22)
	dc.SetLineWidth(1.8)
	dc.DrawLine(cx-shoulderBase*0.35, shoulderY+torsoHeight*0.27, cx+shoulderBase*0.35, shoulderY+torsoHeight*0.27)
	dc.Stroke()
	dc.DrawLine(cx-waistBase*0.40, shoulderY+torsoHeight*0.72, cx+waistBase*0.40, shoulderY+torsoHeight*0.72)
	dc.Stroke()
}

func rgbaImageToANSIHalfBlocks(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return ""
	}

	var out strings.Builder
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x++ {
			tr, tg, tb, _ := rgba8(img.At(bounds.Min.X+x, bounds.Min.Y+y))
			br, bg, bb, _ := uint8(0), uint8(0), uint8(0), uint8(0)
			if y+1 < height {
				br, bg, bb, _ = rgba8(img.At(bounds.Min.X+x, bounds.Min.Y+y+1))
			} else {
				br, bg, bb = tr, tg, tb
			}
			out.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", tr, tg, tb, br, bg, bb))
		}
		out.WriteString("\x1b[0m\n")
	}
	return out.String()
}

func rgba8(c color.Color) (r, g, b, a uint8) {
	r16, g16, b16, a16 := c.RGBA()
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8), uint8(a16 >> 8)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clampFloat(v, minV, maxV float64) float64 {
	return math.Min(maxV, math.Max(minV, v))
}
