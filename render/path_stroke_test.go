package render

import (
	"image/color"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestStrokePathPreservesPerpendicularWidthAndOpenEnds(t *testing.T) {
	b := NewBatch(16)
	points := []geometry.Vec2{{}, {X: 3, Y: 4}}
	paint := color.Color(color.White)
	b.StrokePath(points, PathStroke{Width: 2, Open: true}, paint)
	if len(b.vertices) != 6 || len(b.indices) != 6 {
		t.Fatalf("open segment submitted %d vertices", len(b.vertices))
	}
	// A 3/4/5 segment has the unit perpendicular (-0.8, 0.6).
	if b.vertices[0].DstX != float32(-.8) || b.vertices[0].DstY != float32(.6) ||
		b.vertices[1].DstX != float32(2.2) || b.vertices[1].DstY != float32(4.6) ||
		b.vertices[2].DstX != float32(3.8) || b.vertices[2].DstY != float32(3.4) {
		t.Fatalf("wrong perpendicular stroke: %v", b.vertices)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		b.vertices, b.indices = b.vertices[:0], b.indices[:0]
		b.StrokePath(points, PathStroke{Width: 2, Open: true}, paint)
	}); allocations != 0 {
		t.Fatalf("stroke allocated %v times", allocations)
	}
}
