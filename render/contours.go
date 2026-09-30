package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// VertexAt supplies mapped positions, colors and UVs without copying source
// point/index banks. The mapping must be stable within one contour call. The
// callback is consumed synchronously and never retained.
type VertexAt func(index int) ebiten.Vertex

// Fan submits one contour in its supplied vertex order. With an even-odd fill
// rule, several contours in the same batch can describe concavity and holes.
// Count excludes a repeated closing point when the source stores one.
func (b *Batch) Fan(count int, at VertexAt) {
	if count < 3 || at == nil {
		return
	}
	first := at(0)
	previous := at(1)
	for i := 2; i < count; i++ {
		current := at(i)
		b.Triangle(first, previous, current)
		previous = current
	}
}

// ContourStroke controls a closed or open path over mapped float32 vertices.
// Offsets are rounded to float32 before addition, retaining parity-mask edge
// arithmetic. A zero width selects one pixel. Horizontal exclusion supports
// source programs that draw only the crossing edges of an outline.
type ContourStroke struct {
	Width          float64
	Open           bool
	SkipHorizontal bool
}

func (b *Batch) StrokeContour(count int, at VertexAt, c ContourStroke) {
	if count < 2 || at == nil {
		return
	}
	if c.Width == 0 {
		c.Width = 1
	}
	if c.Width <= 0 || math.IsNaN(c.Width) || math.IsInf(c.Width, 0) {
		return
	}
	end := count
	if c.Open {
		end--
	}
	for i := 0; i < end; i++ {
		a, d := at(i), at((i+1)%count)
		dx, dy := float64(d.DstX-a.DstX), float64(d.DstY-a.DstY)
		length := math.Hypot(dx, dy)
		if length == 0 || c.SkipHorizontal && dy == 0 {
			continue
		}
		nx, ny := float32(-dy/length*c.Width/2), float32(dx/length*c.Width/2)
		q := [4]ebiten.Vertex{a, d, d, a}
		q[0].DstX += nx
		q[0].DstY += ny
		q[1].DstX += nx
		q[1].DstY += ny
		q[2].DstX -= nx
		q[2].DstY -= ny
		q[3].DstX -= nx
		q[3].DstY -= ny
		b.Quad(q)
	}
}

// ParityContour submits ray spans for independently modified edges. SwapY
// exchanges each edge's endpoints on Y without reordering the source contour.
// RayX is the left boundary of the spans, normally just outside the viewport.
// Use FillRuleEvenOdd when the spans supply a parity mask.
type ParityContour struct {
	RayX  float32
	SwapY bool
}

func (b *Batch) ParityContour(count int, at VertexAt, c ParityContour) {
	if count < 3 || at == nil {
		return
	}
	for i := 0; i < count; i++ {
		a, d := at(i), at((i+1)%count)
		if c.SwapY {
			a.DstY, d.DstY = d.DstY, a.DstY
		}
		if a.DstY == d.DstY {
			continue
		}
		leftA, leftD := a, d
		leftA.DstX, leftD.DstX = c.RayX, c.RayX
		b.Quad([4]ebiten.Vertex{a, d, leftD, leftA})
	}
}
