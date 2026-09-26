package motion

import "testing"

func TestSteppedRevealKeepsReplicantsRowsAndFinalHold(t *testing.T) {
	reveal, err := NewSteppedReveal(SteppedRevealConfig{
		Blocks: 10, BlocksPerStep: 1, StepTicks: 3, DoneTick: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		tick    uint64
		visible int
		done    bool
	}{{0, 0, false}, {1, 0, false}, {2, 0, false}, {3, 1, false},
		{29, 9, false}, {30, 10, false}, {99, 10, false}, {100, 10, true}} {
		reveal.SetTick(sample.tick)
		if got := reveal.VisibleCount(); got != sample.visible || reveal.Done() != sample.done {
			t.Fatalf("tick %d: visible %d done %t, want %d/%t", sample.tick, got, reveal.Done(), sample.visible, sample.done)
		}
	}
	reveal.SetTick(0)
	for tick := 1; tick <= 100; tick++ {
		if err := reveal.Step(); err != nil {
			t.Fatal(err)
		}
		want := min(tick/3, 10)
		if got := reveal.VisibleCount(); got != want {
			t.Fatalf("tick %d visible %d, want %d", tick, got, want)
		}
	}
	if got := testing.AllocsPerRun(100, func() { _ = reveal.Step(); reveal.VisibleCount() }); got != 0 {
		t.Fatalf("reveal clock allocated %.2f objects", got)
	}
}
