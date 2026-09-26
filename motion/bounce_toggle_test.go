package motion

import "testing"

func TestBounceToggleMatchesDeltaLogoScaleAndTile(t *testing.T) {
	clock, err := NewBounceToggle(BounceToggleConfig{Start: 1, Velocity: -.02,
		Min: 0, Max: 1, Inclusive: true, ToggleLower: true, Materials: 2})
	if err != nil {
		t.Fatal(err)
	}
	scale, step, tile := 1.0, .02, 0
	toggles := 0
	for tick := 0; tick < 2000; tick++ {
		if pose := clock.Pose(); pose.Value != scale || pose.Velocity != -step || pose.Material != tile {
			t.Fatalf("tick %d pose=%+v, want scale=%v velocity=%v tile=%d", tick,
				pose, scale, -step, tile)
		}
		scale -= step
		if scale <= 0 {
			step = -.02
			tile = (tile + 1) % 2
			toggles++
		}
		if scale >= 1 {
			step = .02
		}
		clock.Step()
	}
	if toggles < 4 {
		t.Fatalf("only %d lower-boundary material switches", toggles)
	}
	if got := testing.AllocsPerRun(100, clock.Step); got != 0 {
		t.Fatalf("bounce toggle step allocated %.2f objects", got)
	}
}
