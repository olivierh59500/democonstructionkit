package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

// WarpOutlineConfig styles the boundaries of the warp's existing grid. Width
// defaults to one, Color to white. White optionally borrows a one-pixel material.
// Map, grid dimensions, depth order and Frame.Time are shared with filled drawing.
type WarpOutlineConfig struct {
	Width float64
	Color color.Color
	White *ebiten.Image
	Blend ebiten.Blend
}

type warpOutline struct {
	config WarpOutlineConfig
	white  *ebiten.Image
	owned  bool
}

func (w *Warp) SetOutline(c WarpOutlineConfig) error {
	if w == nil || w.closed || c.Width < 0 || math.IsNaN(c.Width) || math.IsInf(c.Width, 0) {
		return fmt.Errorf("effects: invalid warp outline target or width")
	}
	if c.White != nil && (c.White.Bounds().Min.X != 0 || c.White.Bounds().Min.Y != 0 || c.White.Bounds().Dx() != 1 || c.White.Bounds().Dy() != 1) {
		return fmt.Errorf("effects: warp outline needs a one-pixel white material")
	}
	if c.Width == 0 {
		c.Width = 1
	}
	if c.Color == nil {
		c.Color = color.White
	}
	o := &warpOutline{config: c, white: c.White}
	if o.white == nil {
		o.white = ebiten.NewImage(1, 1)
		o.white.Fill(color.White)
		o.owned = true
	}
	if w.outline != nil {
		w.outline.close()
	}
	w.outline = o
	return nil
}

// DrawOutline maps ordered grid boundaries without drawing the source image
// or changing its clock. Tint is a filled-image material; outline Color remains
// independently configurable. Shared edges retain per-cell draw order.
func (w *Warp) DrawOutline(dst *ebiten.Image) {
	if w == nil || w.closed || w.outline == nil || dst == nil {
		return
	}
	o := w.outline
	w.batch.Options.Blend = o.config.Blend
	w.batch.Begin(dst, o.white)
	for _, cell := range w.cells {
		quad := [4]geometry.Vec2{{X: cell.x, Y: cell.y}, {X: cell.x + cell.w, Y: cell.y},
			{X: cell.x + cell.w, Y: cell.y + cell.h}, {X: cell.x, Y: cell.y + cell.h}}
		if w.Map != nil {
			for i, point := range quad {
				quad[i] = w.Map(point.X, point.Y, w.Frame.Time)
			}
		}
		w.batch.StrokePath(quad[:], render.PathStroke{Width: o.config.Width}, o.config.Color)
	}
	w.batch.Flush()
}

func (o *warpOutline) close() {
	if o.owned && o.white != nil {
		o.white.Deallocate()
	}
	o.white = nil
}
