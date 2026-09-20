package sprites

import "testing"

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
