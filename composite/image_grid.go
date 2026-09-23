package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ImageGrid places a finite number of copies of a borrowed image. Step is
// independent of image dimensions, so copies may overlap or have gaps. Drawing
// order is left-to-right, top-to-bottom, preserving alpha overlap exactly.
type ImageGrid struct {
	Columns, Rows int
	StepX, StepY  float64
	Options       ebiten.DrawImageOptions
}

func (g ImageGrid) Validate() error {
	if g.Columns < 1 || g.Rows < 1 || g.Columns > 256 || g.Rows > 256 || g.Columns*g.Rows > 16384 ||
		math.IsNaN(g.StepX) || math.IsInf(g.StepX, 0) || math.IsNaN(g.StepY) || math.IsInf(g.StepY, 0) || g.StepX <= 0 || g.StepY <= 0 {
		return fmt.Errorf("composite: invalid image grid")
	}
	return nil
}

// DrawAt places the upper-left copy at x,y. Configure once and reuse each frame.
func (g ImageGrid) DrawAt(dst, source *ebiten.Image, x, y float64) {
	if dst == nil || source == nil || g.Validate() != nil {
		return
	}
	op := g.Options
	for row := 0; row < g.Rows; row++ {
		for column := 0; column < g.Columns; column++ {
			op.GeoM.Reset()
			op.GeoM.Translate(x+float64(column)*g.StepX, y+float64(row)*g.StepY)
			dst.DrawImage(source, &op)
		}
	}
}

// DrawCentered centers the complete step grid around x,y. The last image can
// extend beyond that footprint when the source overlaps its following copy.
func (g ImageGrid) DrawCentered(dst, source *ebiten.Image, x, y float64) {
	g.DrawAt(dst, source, x-float64(g.Columns)*g.StepX/2, y-float64(g.Rows)*g.StepY/2)
}
