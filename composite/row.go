package composite

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Row draws an axis-aligned image band with analytical vertical pixel coverage.
// Unlike multisample antialiasing, even a band thinner than a quarter pixel stays
// visible. This is useful for perspective text, distant scanlines and thin logos.
// Source and destination dimensions are independent of font metrics.
type Row struct {
	Source              Region
	X, Y, Width, Height float64
	Filter              ebiten.Filter
	Blend               ebiten.Blend
	ColorScale          ebiten.ColorScale
}

func rowCoverage(top, height float64, pixel int) float64 {
	return math.Max(0, math.Min(top+height, float64(pixel+1))-math.Max(top, float64(pixel)))
}

func (r Row) Draw(dst, src *ebiten.Image) {
	if dst == nil || src == nil || !r.Source.Valid() || !(Region{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}).Valid() {
		return
	}
	first, last := max(dst.Bounds().Min.Y, int(math.Floor(r.Y))), min(dst.Bounds().Max.Y, int(math.Ceil(r.Y+r.Height)))
	for y := first; y < last; y++ {
		coverage := rowCoverage(r.Y, r.Height, y)
		lo, hi := math.Max(r.Y, float64(y)), math.Min(r.Y+r.Height, float64(y+1))
		v := r.Source.Y + ((lo+hi)/2-r.Y)/r.Height*r.Source.Height
		c := r.ColorScale
		c.ScaleAlpha(float32(coverage))
		vertex := func(x, dy, u float64) ebiten.Vertex {
			return ebiten.Vertex{DstX: float32(x), DstY: float32(dy), SrcX: float32(u), SrcY: float32(v), ColorR: c.R(), ColorG: c.G(), ColorB: c.B(), ColorA: c.A()}
		}
		vertices := [4]ebiten.Vertex{vertex(r.X, float64(y), r.Source.X), vertex(r.X+r.Width, float64(y), r.Source.X+r.Source.Width), vertex(r.X+r.Width, float64(y+1), r.Source.X+r.Source.Width), vertex(r.X, float64(y+1), r.Source.X)}
		dst.DrawTriangles(vertices[:], []uint16{0, 1, 2, 0, 2, 3}, src, &ebiten.DrawTrianglesOptions{Filter: r.Filter, Blend: r.Blend, Address: ebiten.AddressClampToZero, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
	}
}
