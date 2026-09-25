package motion

import "testing"

func TestCameraFollowMatchesNearCenterAndFarMenuEdges(t *testing.T) {
	follow, err := NewCameraFollow(CameraFollowConfig{
		ViewportW: 768, ViewportH: 400, WorldW: 5504, WorldH: 800,
		AnchorX: 352, AnchorY: 168,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range [][2]int{{0, 0}, {352, 168}, {353, 169}, {2500, 500}, {5504 - 416, 800 - 232}, {5504 - 415, 800 - 231}, {5504, 800}} {
		got := follow.At(point[0], point[1])
		xCamera, xScreen := legacyFollowAxis(point[0], 768, 5504, 352)
		yCamera, yScreen := legacyFollowAxis(point[1], 400, 800, 168)
		if got != (CameraFollowPose{xCamera, yCamera, xScreen, yScreen}) {
			t.Fatalf("world %v = %+v, want camera(%d,%d) screen(%d,%d)", point, got, xCamera, yCamera, xScreen, yScreen)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = follow.At(2500, 500) }); allocs != 0 {
		t.Fatalf("camera follow sample allocates %v times", allocs)
	}
}

func legacyFollowAxis(position, viewport, world, anchor int) (camera, screen int) {
	if position <= anchor {
		return 0, position
	}
	if position > world-(viewport-anchor) {
		return world - viewport, viewport - (world - position)
	}
	return position - anchor, anchor
}
