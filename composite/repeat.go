package composite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"math"
)

// Repetition maps a viewport back into an infinitely repeated source image.
// Phase is added in source-pixel units, retaining a legacy texture origin exactly.
type Repetition struct {
	CenterX, CenterY, Zoom, Rotation, PhaseX, PhaseY float64
	Filter                                           ebiten.Filter
}

func Repeat(dst, texture *ebiten.Image, p Repetition) {
	if texture == nil || p.Zoom <= 0 {
		return
	}
	bounds := dst.Bounds()
	corners := [4][2]float64{{float64(bounds.Min.X), float64(bounds.Min.Y)}, {float64(bounds.Max.X), float64(bounds.Min.Y)}, {float64(bounds.Min.X), float64(bounds.Max.Y)}, {float64(bounds.Max.X), float64(bounds.Max.Y)}}
	c, s := math.Cos(p.Rotation), math.Sin(p.Rotation)
	var vertices [4]ebiten.Vertex
	for i, v := range corners {
		x, y := (v[0]-p.CenterX)/p.Zoom, (v[1]-p.CenterY)/p.Zoom
		vertices[i] = ebiten.Vertex{DstX: float32(v[0]), DstY: float32(v[1]), SrcX: float32(x*c + y*s + p.PhaseX), SrcY: float32(-x*s + y*c + p.PhaseY), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1}
	}
	indices := [6]uint16{0, 1, 2, 1, 2, 3}
	dst.DrawTriangles(vertices[:], indices[:], texture, &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat, Filter: p.Filter})
}
