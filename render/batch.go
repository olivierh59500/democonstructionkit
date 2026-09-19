// Package render contains bounded reusable Ebitengine drawing helpers.
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Batch streams triangles, flushing before its fixed capacity or uint16 limit.
// A batch is confined to the Ebitengine drawing goroutine.
type Batch struct {
	vertices    []ebiten.Vertex
	indices     []uint16
	dst, source *ebiten.Image
	Options     ebiten.DrawTrianglesOptions
}

func NewBatch(triangles int) *Batch {
	triangles = max(2, min(triangles, 20000))
	b := &Batch{vertices: make([]ebiten.Vertex, 0, triangles*3), indices: make([]uint16, 0, triangles*3)}
	b.Options.ColorScaleMode = ebiten.ColorScaleModePremultipliedAlpha
	return b
}
func (b *Batch) Begin(dst, source *ebiten.Image) { b.Flush(); b.dst, b.source = dst, source }
func (b *Batch) Triangle(a, c, d ebiten.Vertex) {
	if len(b.vertices)+3 > cap(b.vertices) {
		b.Flush()
	}
	i := uint16(len(b.vertices))
	b.vertices = append(b.vertices, a, c, d)
	b.indices = append(b.indices, i, i+1, i+2)
}
func (b *Batch) Quad(v [4]ebiten.Vertex) { b.Triangle(v[0], v[1], v[2]); b.Triangle(v[0], v[2], v[3]) }
func (b *Batch) Flush() {
	if len(b.indices) == 0 {
		return
	}
	b.dst.DrawTriangles(b.vertices, b.indices, b.source, &b.Options)
	b.vertices = b.vertices[:0]
	b.indices = b.indices[:0]
}

// Vertex converts a straight-alpha color to Ebitengine's premultiplied format.
func Vertex(x, y, u, v float64, c color.Color) ebiten.Vertex {
	r, g, b, a := c.RGBA()
	return ebiten.Vertex{DstX: float32(x), DstY: float32(y), SrcX: float32(u), SrcY: float32(v), ColorR: float32(r) / 65535, ColorG: float32(g) / 65535, ColorB: float32(b) / 65535, ColorA: float32(a) / 65535}
}
func (b *Batch) Rect(x, y, w, h float64, source image.Rectangle, c color.Color) {
	u, v := float64(source.Min.X), float64(source.Min.Y)
	u2, v2 := float64(source.Max.X), float64(source.Max.Y)
	b.Quad([4]ebiten.Vertex{Vertex(x, y, u, v, c), Vertex(x+w, y, u2, v, c), Vertex(x+w, y+h, u2, v2, c), Vertex(x, y+h, u, v2, c)})
}
