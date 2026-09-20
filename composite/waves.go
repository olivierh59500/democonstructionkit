package composite

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// StripWave separates phase, amplitude, per-pixel frequency and per-tick speed.
type StripWave struct{ Phase, Amplitude, Spatial, Speed float64 }

// WaveStrips translates rows or columns of any image. Source crops, destination
// offsets and phase are independent; the same instance may deform a text surface,
// sprite or logo. Advance is explicit, so repeated DrawAt calls are deterministic.
type WaveStrips struct {
	Axis      Axis
	Thickness int
	Waves     []StripWave
	Filter    ebiten.Filter
}

func (w *WaveStrips) Advance() {
	for i := range w.Waves {
		w.Waves[i].Phase += w.Waves[i].Speed
	}
}
func (w *WaveStrips) DrawAt(dst, src *ebiten.Image, x, y float64) {
	if dst == nil || src == nil {
		return
	}
	band := w.Thickness
	if band < 1 {
		band = 1
	}
	b := src.Bounds()
	extent := b.Dy()
	if w.Axis == Columns {
		extent = b.Dx()
	}
	for i := 0; i < extent; i += band {
		offset := 0.0
		for _, wave := range w.Waves {
			offset += math.Sin(wave.Phase+float64(i)*wave.Spatial) * wave.Amplitude
		}
		r := Region{X: float64(b.Min.X), Y: float64(b.Min.Y + i), Width: float64(b.Dx()), Height: float64(min(band, extent-i))}
		dx, dy := x+offset, y+float64(i)
		if w.Axis == Columns {
			r = Region{X: float64(b.Min.X + i), Y: float64(b.Min.Y), Width: float64(min(band, extent-i)), Height: float64(b.Dy())}
			dx, dy = x+float64(i), y+offset
		}
		op := ebiten.DrawImageOptions{Filter: w.Filter}
		op.GeoM.Translate(dx, dy)
		DrawRegion(dst, src, r, &op)
	}
}
