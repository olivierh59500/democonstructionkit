package composite

import "github.com/hajimehoshi/ebiten/v2"

// ProfileStrips deforms any image with a repeating, caller-supplied displacement
// table. Phase and Speed are table entries; Step controls sampling within the
// image (zero defaults to one). Thickness controls coarse versus fine bands.
type ProfileStrips struct {
	Offsets                       []float64
	Axis                          Axis
	Phase, Speed, Step, Thickness int
	Filter                        ebiten.Filter
}

func (p *ProfileStrips) Advance() { p.Phase += p.Speed }
func (p *ProfileStrips) DrawAt(dst, src *ebiten.Image, x, y float64) {
	if dst == nil || src == nil || len(p.Offsets) == 0 {
		return
	}
	b := src.Bounds()
	extent := b.Dy()
	if p.Axis == Columns {
		extent = b.Dx()
	}
	band, step := max(1, p.Thickness), p.Step
	if step == 0 {
		step = 1
	}
	for i := 0; i < extent; i += band {
		index := ((p.Phase+i*step)%len(p.Offsets) + len(p.Offsets)) % len(p.Offsets)
		r := Region{X: float64(b.Min.X), Y: float64(b.Min.Y + i), Width: float64(b.Dx()), Height: float64(min(band, extent-i))}
		dx, dy := x+p.Offsets[index], y+float64(i)
		if p.Axis == Columns {
			r = Region{X: float64(b.Min.X + i), Y: float64(b.Min.Y), Width: float64(min(band, extent-i)), Height: float64(b.Dy())}
			dx, dy = x+float64(i), y+p.Offsets[index]
		}
		op := ebiten.DrawImageOptions{Filter: p.Filter}
		op.GeoM.Translate(dx, dy)
		DrawRegion(dst, src, r, &op)
	}
}
