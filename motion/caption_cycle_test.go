package motion

import "testing"

func TestCaptionCycleMatchesUnionTNTTextSlides(t *testing.T) {
	cycle, err := NewCaptionCycle(CaptionCycleConfig{
		Count: 5, Top: -18, Bottom: 0, StartY: -18,
		Speed: 2, InitialWait: 200, HoldWait: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	index, wait, y, increment := 0, 200, -18.0, 2.0
	for tick := 0; tick < 4000; tick++ {
		if got := cycle.Pose(); got != (CaptionPose{Index: index, Y: y, Wait: wait, Velocity: increment}) {
			t.Fatalf("before tick %d pose=%+v, want index=%d y=%v wait=%d velocity=%v", tick,
				got, index, y, wait, increment)
		}
		if wait == 100 {
			y += increment
		}
		if y >= 0 {
			y = 0
			wait--
			if wait <= 0 {
				increment, wait = -2, 100
			}
		}
		if y <= -18 {
			y = -18
			wait--
			if wait <= 0 {
				increment, wait = 2, 100
				index = (index + 1) % 5
			}
		}
		cycle.Step()
	}
	if got := testing.AllocsPerRun(100, cycle.Step); got != 0 {
		t.Fatalf("caption cycle step allocated %.2f objects", got)
	}
}
