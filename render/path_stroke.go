package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// PathStroke controls a uniform-color path built from double-precision points.
// Open omits the closing edge; Width defaults to one pixel. Source coordinates
// cover the unit square on each edge, suitable for a one-pixel white material.
type PathStroke struct {
	Width float64
	Open  bool
}

// StrokePath retains double-precision subtraction, normalization and offsets
// until vertices are submitted. It differs deliberately from StrokeContour,
// which offsets float32 image-mask vertices and retains their attributes.
// Geometry uses the caller's current batch/source without allocating storage.
func (b *Batch) StrokePath(points []geometry.Vec2, style PathStroke, paint color.Color) {
	if len(points) < 2 || paint == nil || math.IsNaN(style.Width) || math.IsInf(style.Width, 0) || style.Width < 0 {
		return
	}
	width := style.Width
	if width == 0 {
		width = 1
	}
	edges := len(points)
	if style.Open {
		edges--
	}
	for i := 0; i < edges; i++ {
		a, c := points[i], points[(i+1)%len(points)]
		dx, dy := c.X-a.X, c.Y-a.Y
		length := math.Hypot(dx, dy)
		if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
			continue
		}
		x, y := -dy/length*width/2, dx/length*width/2
		b.Quad([4]ebiten.Vertex{
			Vertex(a.X+x, a.Y+y, 0, 0, paint), Vertex(c.X+x, c.Y+y, 1, 0, paint),
			Vertex(c.X-x, c.Y-y, 1, 1, paint), Vertex(a.X-x, a.Y-y, 0, 1, paint),
		})
	}
}
