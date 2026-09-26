package composite

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Repetition maps a viewport back into an infinitely repeated source image.
// Phase is added in source-pixel units, retaining a legacy texture origin exactly.
type Repetition struct {
	Color                                            [4]float32 // Zero means white; otherwise straight vertex RGBA multipliers.
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
	color := p.Color
	if color == ([4]float32{}) {
		color = [4]float32{1, 1, 1, 1}
	}
	var vertices [4]ebiten.Vertex
	for i, v := range corners {
		x, y := (v[0]-p.CenterX)/p.Zoom, (v[1]-p.CenterY)/p.Zoom
		vertices[i] = ebiten.Vertex{DstX: float32(v[0]), DstY: float32(v[1]), SrcX: float32(x*c + y*s + p.PhaseX), SrcY: float32(-x*s + y*c + p.PhaseY), ColorR: color[0], ColorG: color[1], ColorB: color[2], ColorA: color[3]}
	}
	indices := [6]uint16{0, 1, 2, 1, 2, 3}
	dst.DrawTriangles(vertices[:], indices[:], texture, &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat, Filter: p.Filter})
}

// RepeatSourceQuad preserves a source-sized textured quad before rotation and
// zoom. This is useful when a production's rasterization depended on rounded
// destination vertices of a deliberately oversized repeating texture.
func RepeatSourceQuad(dst, texture *ebiten.Image, p Repetition, size image.Point) {
	if dst == nil || texture == nil || p.Zoom <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	var vertices [4]ebiten.Vertex
	fillSourceQuadVertices(&vertices, p, size)
	dst.DrawTriangles(vertices[:], sourceQuadIndices[:], texture,
		&ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat, Filter: p.Filter})
}

var sourceQuadIndices = [6]uint16{0, 1, 2, 1, 2, 3}

func fillSourceQuadVertices(vertices *[4]ebiten.Vertex, p Repetition, size image.Point) {
	color := p.Color
	if color == ([4]float32{}) {
		color = [4]float32{1, 1, 1, 1}
	}
	cosRot, sinRot := math.Cos(p.Rotation), math.Sin(p.Rotation)
	for i, corner := range [...]image.Point{{}, {X: size.X}, {Y: size.Y}, {X: size.X, Y: size.Y}} {
		x := float64(corner.X) - float64(size.X)/2
		y := float64(corner.Y) - float64(size.Y)/2
		vertices[i] = ebiten.Vertex{
			DstX:   float32((x*cosRot-y*sinRot)*p.Zoom + p.CenterX),
			DstY:   float32((x*sinRot+y*cosRot)*p.Zoom + p.CenterY),
			SrcX:   float32(float64(corner.X) + p.PhaseX - float64(size.X)/2),
			SrcY:   float32(float64(corner.Y) + p.PhaseY - float64(size.Y)/2),
			ColorR: color[0], ColorG: color[1], ColorB: color[2], ColorA: color[3],
		}
	}
}
