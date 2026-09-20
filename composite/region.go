package composite

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Region preserves fractional source coordinates, as used by atlases whose
// width is not an integer multiple of the glyph count. X/Y are absolute atlas
// coordinates; the local destination starts at (0,0) and has size Width/Height.
type Region struct{ X, Y, Width, Height float64 }

func (r Region) Valid() bool {
	for _, v := range []float64{r.X, r.Y, r.Width, r.Height} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return r.Width > 0 && r.Height > 0
}

// DrawRegion applies the usual image transform, filtering, tint and blending
// without rounding a source crop to image.Rectangle. It allocates no textures.
func DrawRegion(dst, src *ebiten.Image, r Region, options *ebiten.DrawImageOptions) {
	if dst == nil || src == nil || !r.Valid() {
		return
	}
	var op ebiten.DrawImageOptions
	if options != nil {
		op = *options
	}
	points := [4][2]float64{{0, 0}, {r.Width, 0}, {r.Width, r.Height}, {0, r.Height}}
	uv := [4][2]float64{{r.X, r.Y}, {r.X + r.Width, r.Y}, {r.X + r.Width, r.Y + r.Height}, {r.X, r.Y + r.Height}}
	var vertices [4]ebiten.Vertex
	for i, p := range points {
		x, y := op.GeoM.Apply(p[0], p[1])
		vertices[i] = ebiten.Vertex{DstX: float32(x), DstY: float32(y), SrcX: float32(uv[i][0]), SrcY: float32(uv[i][1]), ColorR: op.ColorScale.R(), ColorG: op.ColorScale.G(), ColorB: op.ColorScale.B(), ColorA: op.ColorScale.A()}
	}
	dst.DrawTriangles(vertices[:], []uint16{0, 1, 2, 0, 2, 3}, src, &ebiten.DrawTrianglesOptions{Filter: op.Filter, Blend: op.Blend, ColorM: op.ColorM, CompositeMode: op.CompositeMode, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha, Address: ebiten.AddressClampToZero})
}
