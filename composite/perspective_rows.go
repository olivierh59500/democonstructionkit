package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type projectedRowGroup struct {
	vertices []ebiten.Vertex
	filter   ebiten.Filter
	blend    ebiten.Blend
}

// RowProjection compiles an arbitrary scanline mapping and analytical pixel
// coverage into cached geometry. Draw only submits that geometry, without
// recalculating projection/coverage or allocating per frame. Any live image can
// be used as its source; source and destination remain borrowed and distinct.
type RowProjection struct {
	groups  []projectedRowGroup
	indices []uint16
}

// NewRowProjection preserves row order, filters, blends and subpixel coverage.
// At most 65,536 covered pixel rows are accepted, with bounded draw batches.
func NewRowProjection(rows []Row) (*RowProjection, error) {
	if len(rows) == 0 || len(rows) > 65536 {
		return nil, fmt.Errorf("composite: invalid row projection size")
	}
	const batchRows = 256
	p := &RowProjection{indices: make([]uint16, batchRows*6)}
	for i := 0; i < batchRows; i++ {
		base := uint16(i * 4)
		copy(p.indices[i*6:], []uint16{base, base + 1, base + 2, base, base + 2, base + 3})
	}
	count := 0
	for _, r := range rows {
		if !r.Source.Valid() || !bandFinite(r.X) || !bandFinite(r.Y) || !bandFinite(r.Width) || !bandFinite(r.Height) || r.Width <= 0 || r.Height <= 0 || math.Abs(r.Y) > 1e9 || r.Height > 65536 {
			return nil, fmt.Errorf("composite: invalid projected row")
		}
		first, last := int(math.Floor(r.Y)), int(math.Ceil(r.Y+r.Height))
		if last-first > 65536-count {
			return nil, fmt.Errorf("composite: row projection exceeds geometry budget")
		}
		count += last - first
		for y := first; y < last; y++ {
			if len(p.groups) == 0 || len(p.groups[len(p.groups)-1].vertices) == batchRows*4 || p.groups[len(p.groups)-1].filter != r.Filter || p.groups[len(p.groups)-1].blend != r.Blend {
				p.groups = append(p.groups, projectedRowGroup{filter: r.Filter, blend: r.Blend})
			}
			g := &p.groups[len(p.groups)-1]
			coverage := rowCoverage(r.Y, r.Height, y)
			lo, hi := math.Max(r.Y, float64(y)), math.Min(r.Y+r.Height, float64(y+1))
			v := r.Source.Y + ((lo+hi)/2-r.Y)/r.Height*r.Source.Height
			color := r.ColorScale
			color.ScaleAlpha(float32(coverage))
			vertex := func(x, dy, u float64) ebiten.Vertex {
				return ebiten.Vertex{DstX: float32(x), DstY: float32(dy), SrcX: float32(u), SrcY: float32(v), ColorR: color.R(), ColorG: color.G(), ColorB: color.B(), ColorA: color.A()}
			}
			g.vertices = append(g.vertices, vertex(r.X, float64(y), r.Source.X), vertex(r.X+r.Width, float64(y), r.Source.X+r.Source.Width), vertex(r.X+r.Width, float64(y+1), r.Source.X+r.Source.Width), vertex(r.X, float64(y+1), r.Source.X))
		}
	}
	return p, nil
}
func (p *RowProjection) Draw(dst, src *ebiten.Image) {
	if p == nil || dst == nil || src == nil {
		return
	}
	for _, group := range p.groups {
		dst.DrawTriangles(group.vertices, p.indices[:len(group.vertices)/4*6], src, &ebiten.DrawTrianglesOptions{Filter: group.filter, Blend: group.blend, Address: ebiten.AddressClampToZero, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
	}
}
