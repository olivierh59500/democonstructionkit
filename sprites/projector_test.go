package sprites

import (
	"reflect"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

type indexedTestPoints struct{ values []Point }

func (points *indexedTestPoints) Len() int { return len(points.values) }
func (points *indexedTestPoints) XYZ(index int) geometry.Vec3 {
	point := points.values[index]
	return geometry.Vec3{X: point.X, Y: point.Y, Z: point.Z}
}
func (points *indexedTestPoints) ImageIndex(index int) int { return points.values[index].Image }

func TestProjectionAxisAndDepthConventions(t *testing.T) {
	var renderer Projector
	points := []Point{{X: 20, Y: -10, Z: 100}, {Z: 50}}
	renderer.Draw(nil, points, nil, Projection{Focal: 100, CenterX: 100, CenterY: 50, YUp: true, AscendingDepth: true})
	if len(renderer.points) != 2 || renderer.points[0].depth != 50 || renderer.points[1].x != 110 || renderer.points[1].y != 55 {
		t.Fatal(renderer.points)
	}
	renderer.Draw(nil, points, nil, Projection{Focal: 100, CenterX: 100, CenterY: 50, YUp: false, AscendingDepth: false})
	if renderer.points[0].depth != 100 || renderer.points[0].y != 45 {
		t.Fatal(renderer.points)
	}
}

func TestProjectorDrawIndexedMatchesPointSlice(t *testing.T) {
	points := &indexedTestPoints{values: []Point{{X: 20, Y: -10, Z: 100, Image: 3},
		{X: -30, Y: 40, Z: -50, Image: 1},
		{X: 5, Y: 8, Z: 200, Image: 2},
		{X: -12, Y: -7, Z: 0, Image: 0}}}
	config := Projection{
		Matrix:    [9]float64{.75, .25, -.4, -.25, 1.1, .3, .15, -.6, .9},
		Translate: Point{X: 7, Y: -4, Z: 350}, Focal: 600,
		CenterX: 320, CenterY: 193, YUp: true, AscendingDepth: true,
	}
	var fromSlice, fromReader Projector
	for _, cull := range []bool{false, true} {
		config.CullPositiveModelZ = cull
		fromSlice.Draw(nil, points.values, nil, config)
		fromReader.DrawIndexed(nil, points, nil, config)
		if !reflect.DeepEqual(fromReader.points, fromSlice.points) {
			t.Fatalf("indexed projection differs with culling %t: %v vs %v", cull, fromReader.points, fromSlice.points)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		fromReader.DrawIndexed(nil, points, nil, config)
	}); allocations != 0 {
		t.Fatalf("indexed projection allocates %g times per draw", allocations)
	}
}
