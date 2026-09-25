package timeline

import "testing"

func TestHoldRampPreservesSplashProgressAndExitTick(t *testing.T) {
	ramp, err := NewHoldRamp(HoldRampConfig{Frames: 90, InteriorOffset: .02})
	if err != nil {
		t.Fatal(err)
	}
	iteration, progress := 0, 0.0
	for tick := 1; tick <= 91; tick++ {
		legacyDone := false
		if iteration < 90 {
			iteration++
			progress = float64(iteration) / 90
		} else {
			legacyDone = true
			progress = 0
		}
		if progress > 0 && progress < 1 {
			progress = min(progress+.02, 1)
		}
		gotDone := ramp.Step()
		if gotDone != legacyDone {
			t.Fatalf("tick %d done = %v, want %v", tick, gotDone, legacyDone)
		}
		if gotDone {
			ramp.Reset()
		}
		if ramp.Progress() != progress {
			t.Fatalf("tick %d progress = %v, want %v", tick, ramp.Progress(), progress)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { ramp.Step() }); allocations != 0 {
		t.Fatalf("hold ramp step allocates %v times", allocations)
	}
}

func TestHoldRampRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewHoldRamp(HoldRampConfig{}); err == nil {
		t.Fatal("accepted zero frames")
	}
}
