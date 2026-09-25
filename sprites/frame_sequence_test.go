package sprites

import "testing"

func TestFrameSequenceKeepsAuthoredOrderAndLoopBoundary(t *testing.T) {
	indices := []int{0, 3, 5, 6}
	sequence, err := NewFrameSequence(FrameSequence{Duration: .35, Indices: indices, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	indices[0] = 99
	for _, check := range []struct {
		time  float64
		frame int
	}{
		{-1, 0}, {0, 0}, {.1, 3}, {.2, 5}, {.3, 6}, {.35, 0}, {.45, 3},
	} {
		if got := sequence.Current(check.time); got != check.frame {
			t.Fatalf("time %g frame = %d, want %d", check.time, got, check.frame)
		}
	}
	sequence.Loop = false
	if got := sequence.Current(1); got != 6 {
		t.Fatalf("finite sequence final frame = %d, want 6", got)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = sequence.Current(.2) }); allocations != 0 {
		t.Fatalf("frame sampling allocates %v times", allocations)
	}
}
