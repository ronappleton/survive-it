package theme

import rl "github.com/gen2brain/raylib-go/raylib"

// NineSlice describes a scalable 9-patch texture.
// The corner sizes are specified in source-texture pixels.
// Centre area is stretched; edges are stretched along one axis; corners are drawn verbatim.
type NineSlice struct {
	Tex    rl.Texture2D
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

// DrawNineSlice renders ns into dest, tinted by tint.
// If the texture is not loaded (ID == 0) a flat filled rectangle is drawn instead.
func DrawNineSlice(ns NineSlice, dest rl.Rectangle, tint rl.Color) {
	if ns.Tex.ID == 0 {
		rl.DrawRectangleRec(dest, rl.Fade(tint, 0.35))
		return
	}

	sw := float32(ns.Tex.Width)
	sh := float32(ns.Tex.Height)
	l := float32(ns.Left)
	r := float32(ns.Right)
	t := float32(ns.Top)
	b := float32(ns.Bottom)
	cx := sw - l - r // center width in source
	cy := sh - t - b // center height in source

	dx := dest.X
	dy := dest.Y
	dw := dest.Width
	dh := dest.Height
	dl := l
	dr := r
	dt := t
	db := b

	// Clamp so corners never overflow
	if dl+dr > dw {
		dl = dw / 2
		dr = dw / 2
	}
	if dt+db > dh {
		dt = dh / 2
		db = dh / 2
	}
	dcx := dw - dl - dr // center width in dest
	dcy := dh - dt - db // center height in dest

	type patch struct {
		src  rl.Rectangle
		dest rl.Rectangle
	}

	patches := [9]patch{
		// top-left
		{rl.NewRectangle(0, 0, l, t), rl.NewRectangle(dx, dy, dl, dt)},
		// top-centre
		{rl.NewRectangle(l, 0, cx, t), rl.NewRectangle(dx+dl, dy, dcx, dt)},
		// top-right
		{rl.NewRectangle(sw-r, 0, r, t), rl.NewRectangle(dx+dw-dr, dy, dr, dt)},
		// mid-left
		{rl.NewRectangle(0, t, l, cy), rl.NewRectangle(dx, dy+dt, dl, dcy)},
		// centre
		{rl.NewRectangle(l, t, cx, cy), rl.NewRectangle(dx+dl, dy+dt, dcx, dcy)},
		// mid-right
		{rl.NewRectangle(sw-r, t, r, cy), rl.NewRectangle(dx+dw-dr, dy+dt, dr, dcy)},
		// bottom-left
		{rl.NewRectangle(0, sh-b, l, b), rl.NewRectangle(dx, dy+dh-db, dl, db)},
		// bottom-centre
		{rl.NewRectangle(l, sh-b, cx, b), rl.NewRectangle(dx+dl, dy+dh-db, dcx, db)},
		// bottom-right
		{rl.NewRectangle(sw-r, sh-b, r, b), rl.NewRectangle(dx+dw-dr, dy+dh-db, dr, db)},
	}

	for _, p := range patches {
		if p.dest.Width <= 0 || p.dest.Height <= 0 {
			continue
		}
		rl.DrawTexturePro(ns.Tex, p.src, p.dest, rl.Vector2{}, 0, tint)
	}
}
