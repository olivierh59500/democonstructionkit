package motion

import "testing"

func TestWalkParallaxMatchesHallAndBannerWraps(t *testing.T) {
	walk, err := NewWalkParallax(WalkParallaxConfig{Layers: []WalkLayer{
		{Start: 0, Speed: 5, Lower: &WrapLimit{Boundary: -5506, Restart: 0, Inclusive: true}, Upper: &WrapLimit{Boundary: 0, Restart: -5506, Inclusive: true}},
		{Start: 0, Speed: 3, Lower: &WrapLimit{Boundary: -94, Restart: -78, Inclusive: true}, Upper: &WrapLimit{Boundary: -78, Restart: -94, Inclusive: true}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	back, banner := 0, 0
	for _, direction := range append(append(make([]int, 0, 2500), repeatDirection(1, 1300)...), repeatDirection(-1, 1300)...) {
		back -= direction * 5
		if back <= -5506 {
			back = 0
		} else if back >= 0 {
			back = -5506
		}
		banner -= direction * 3
		if direction > 0 && banner <= -94 {
			banner = -78
		} else if direction < 0 && banner >= -78 {
			banner = -94
		}
		walk.Advance(direction)
		if walk.At(0) != float64(back) || walk.At(1) != float64(banner) {
			t.Fatalf("direction %d positions = %v/%v, want %d/%d", direction, walk.At(0), walk.At(1), back, banner)
		}
	}
	walk.Set(0, -3682)
	if walk.At(0) != -3682 {
		t.Fatal("camera placement did not update phase")
	}
	walk.Reset()
	if walk.At(0) != 0 || walk.At(1) != 0 {
		t.Fatal("reset kept walk parallax phases")
	}
	if allocs := testing.AllocsPerRun(100, func() { walk.Advance(1) }); allocs != 0 {
		t.Fatalf("walk parallax step allocates %v times", allocs)
	}
}

func repeatDirection(direction, count int) []int {
	result := make([]int, count)
	for i := range result {
		result[i] = direction
	}
	return result
}
