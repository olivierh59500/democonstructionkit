package geometry

import (
	"math"
	"testing"
)

func TestClipNearSegmentPreservesOrderAndIntersections(t *testing.T) {
	a, b := Vec3{Z: .5}, Vec3{X: 4, Y: 2, Z: 2.5}
	intersection := Vec3{X: 1, Y: .5, Z: 1}
	for _, c := range []struct {
		a, b, first, second Vec3
		visible             bool
	}{
		{a, b, intersection, b, true},
		{b, a, b, intersection, true},
		{intersection, b, intersection, b, true},
		{a, Vec3{Z: .8}, a, Vec3{Z: .8}, false},
	} {
		first, second, visible := ClipNearSegment(c.a, c.b, 1)
		if first != c.first || second != c.second || visible != c.visible {
			t.Fatalf("clip (%v,%v): got (%v,%v,%v)", c.a, c.b, first, second, visible)
		}
	}
	for _, near := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, _, visible := ClipNearSegment(a, b, near); visible {
			t.Fatalf("accepted invalid near plane %v", near)
		}
	}
}
