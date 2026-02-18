package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
	"github.com/fogleman/gg"
)

func renderPlayerBodyANSI(p game.PlayerConfig, widthChars, heightRows int) string {
	if widthChars < 12 || heightRows < 10 {
		return playerASCIIArt(p)
	}

	widthChars = clampInt(widthChars, 12, 44)
	heightRows = clampInt(heightRows, 10, 64)

	w := widthChars
	h := heightRows * 2
	dc := gg.NewContext(w, h)

	// Transparent background so pane border/background stays visible.
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

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

	figureHeight := lerp(0.62, 0.90, heightNorm) * float64(h)
	feetY := float64(h) * 0.95
	topY := feetY - figureHeight
	cx := float64(w) * 0.5

	shoulderBase := lerp(0.16, 0.23, weightNorm) * float64(w)
	hipBase := shoulderBase * 0.8
	switch p.BodyType {
	case game.BodyTypeMale:
		shoulderBase *= 1.08
		hipBase *= 0.95
	case game.BodyTypeFemale:
		shoulderBase *= 0.95
		hipBase *= 1.08
	}
	if shoulderBase < 6 {
		shoulderBase = 6
	}

	headR := figureHeight * 0.09
	headCY := topY + headR
	neckY := headCY + headR*0.95
	shoulderY := neckY + figureHeight*0.03
	torsoHeight := figureHeight * 0.33
	hipY := shoulderY + torsoHeight
	legLength := feetY - hipY

	armLen := figureHeight * 0.30
	handY := shoulderY + armLen
	armSpread := shoulderBase*0.72 + lerp(1.0, 4.0, weightNorm)

	legSpread := hipBase * 0.34
	leftFootX := cx - legSpread - 1.2
	rightFootX := cx + legSpread + 1.2

	primary := color.RGBA{R: 0, G: 230, B: 110, A: 235}
	shade := color.RGBA{R: 0, G: 145, B: 60, A: 230}
	accent := color.RGBA{R: 120, G: 255, B: 170, A: 210}

	// Soft glow and shadow for depth.
	dc.SetRGBA(0, 1, 0, 0.08)
	dc.DrawEllipse(cx, shoulderY+torsoHeight*0.8, shoulderBase*1.9, figureHeight*0.56)
	dc.Fill()
	dc.SetRGBA(0, 0.7, 0.25, 0.20)
	dc.DrawEllipse(cx, feetY+2, shoulderBase*0.9, 2.4)
	dc.Fill()

	// Head with simple gradient.
	headGrad := gg.NewRadialGradient(cx-headR*0.25, headCY-headR*0.35, headR*0.2, cx, headCY, headR*1.2)
	headGrad.AddColorStop(0.0, accent)
	headGrad.AddColorStop(1.0, primary)
	dc.SetFillStyle(headGrad)
	dc.DrawCircle(cx, headCY, headR)
	dc.Fill()
	dc.SetRGBA(0, 0.25, 0.1, 0.75)
	dc.SetLineWidth(1)
	dc.DrawCircle(cx, headCY, headR)
	dc.Stroke()

	// Torso.
	torsoGrad := gg.NewLinearGradient(cx, shoulderY, cx, hipY)
	torsoGrad.AddColorStop(0.0, primary)
	torsoGrad.AddColorStop(1.0, shade)
	dc.SetFillStyle(torsoGrad)
	dc.DrawRoundedRectangle(cx-shoulderBase*0.52, shoulderY, shoulderBase*1.04, torsoHeight, shoulderBase*0.22)
	dc.Fill()

	// Arms and legs.
	limbThickness := lerp(1.8, 3.1, weightNorm)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	dc.SetLineWidth(limbThickness)
	dc.SetColor(primary)

	leftShoulderX := cx - shoulderBase*0.48
	rightShoulderX := cx + shoulderBase*0.48
	dc.DrawLine(leftShoulderX, shoulderY+2, leftShoulderX-armSpread, handY)
	dc.Stroke()
	dc.DrawLine(rightShoulderX, shoulderY+2, rightShoulderX+armSpread, handY)
	dc.Stroke()

	leftHipX := cx - hipBase*0.35
	rightHipX := cx + hipBase*0.35
	kneeDrop := legLength * 0.46
	dc.DrawLine(leftHipX, hipY, leftHipX-0.8, hipY+kneeDrop)
	dc.Stroke()
	dc.DrawLine(leftHipX-0.8, hipY+kneeDrop, leftFootX, feetY)
	dc.Stroke()

	dc.DrawLine(rightHipX, hipY, rightHipX+0.8, hipY+kneeDrop)
	dc.Stroke()
	dc.DrawLine(rightHipX+0.8, hipY+kneeDrop, rightFootX, feetY)
	dc.Stroke()

	// Chest and waist contours for a less "flat" body.
	dc.SetRGBA(0.85, 1, 0.85, 0.22)
	dc.SetLineWidth(1)
	dc.DrawLine(cx-shoulderBase*0.35, shoulderY+torsoHeight*0.27, cx+shoulderBase*0.35, shoulderY+torsoHeight*0.27)
	dc.Stroke()
	dc.DrawLine(cx-hipBase*0.32, shoulderY+torsoHeight*0.76, cx+hipBase*0.32, shoulderY+torsoHeight*0.76)
	dc.Stroke()

	return rgbaImageToANSIHalfBlocks(dc.Image())
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
			tr, tg, tb, ta := rgba8(img.At(bounds.Min.X+x, bounds.Min.Y+y))
			br, bg, bb, ba := uint8(0), uint8(0), uint8(0), uint8(0)
			if y+1 < height {
				br, bg, bb, ba = rgba8(img.At(bounds.Min.X+x, bounds.Min.Y+y+1))
			}

			if ta < 8 && ba < 8 {
				out.WriteByte(' ')
				continue
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
