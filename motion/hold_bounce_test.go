package motion

import "testing"

func TestHoldBounceMatchesLongMenuLogoCycle(t *testing.T) {
	bounce, err := NewHoldBounce(HoldBounceConfig{Min: 0, Max: 1, Velocity: -.04, HoldTicks: 1000, LeadTicks: 5})
	if err != nil {
		t.Fatal(err)
	}
	scale, step, wait := 1.0, -.04, 1000
	for tick := 1; tick <= 5000; tick++ {
		if wait <= 5 {
			scale += step
			if scale <= 0 {
				scale, step = 0, .04
			}
		}
		if wait <= 0 && scale >= 1 {
			scale, step, wait = 1, -.04, 1000
		}
		wait--
		if got := bounce.Step(); got != scale {
			t.Fatalf("tick %d logo scale = %v, want %v", tick, got, scale)
		}
	}
	bounce.Reset()
	if bounce.At() != 1 {
		t.Fatal("reset did not restore full logo scale")
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = bounce.Step() }); allocs != 0 {
		t.Fatalf("hold bounce step allocates %v times", allocs)
	}
}
