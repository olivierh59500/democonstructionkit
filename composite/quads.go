package composite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

// QuadBatch preserves source/destination quad topology for original raster and
// strip renderers. It owns bounded reusable storage and flushes before overflow.
type QuadBatch struct {
	vertices            []ebiten.Vertex
	indices             []uint16
	destination, source *ebiten.Image
	Options             ebiten.DrawTrianglesOptions
}

func NewQuadBatch(capacity int) *QuadBatch {
	capacity = max(1, min(capacity, 16383))
	return &QuadBatch{vertices: make([]ebiten.Vertex, 0, capacity*4), indices: make([]uint16, 0, capacity*6)}
}
func (q *QuadBatch) Begin(destination, source *ebiten.Image) {
	q.Flush()
	q.destination, q.source = destination, source
}

// Rect scales a source rectangle to a destination rectangle without normalizing
// or rounding supplied coordinates. Color uses Ebitengine's vertex conventions.
func (q *QuadBatch) Rect(source image.Rectangle, x, y, w, h float32) {
	if len(q.vertices)+4 > cap(q.vertices) {
		q.Flush()
	}
	base := uint16(len(q.vertices))
	sx, sy, ex, ey := float32(source.Min.X), float32(source.Min.Y), float32(source.Max.X), float32(source.Max.Y)
	for _, v := range [4][4]float32{{x, y, sx, sy}, {x + w, y, ex, sy}, {x + w, y + h, ex, ey}, {x, y + h, sx, ey}} {
		q.vertices = append(q.vertices, ebiten.Vertex{DstX: v[0], DstY: v[1], SrcX: v[2], SrcY: v[3], ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1})
	}
	q.indices = append(q.indices, base, base+1, base+2, base, base+2, base+3)
}
func (q *QuadBatch) Flush() {
	if len(q.indices) == 0 {
		return
	}
	q.destination.DrawTriangles(q.vertices, q.indices, q.source, &q.Options)
	q.vertices = q.vertices[:0]
	q.indices = q.indices[:0]
}
