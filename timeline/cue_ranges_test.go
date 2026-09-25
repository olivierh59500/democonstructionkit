package timeline

import (
	"math"
	"testing"
)

func TestCueRangesPreserveStrictStageBoundaries(t *testing.T) {
	program, err := NewCueRanges([]CueRange{
		{Start: 0, End: 100},
		{Start: 120, End: 360, OpenStart: true},
		{Start: 380, End: 480, OpenStart: true},
		{Start: 500, End: 740, OpenStart: true},
		{Start: 760, End: 1000, OpenStart: true},
		{Start: 1000, End: 1100, OpenStart: true},
		{Start: 1120, End: math.Inf(1), OpenStart: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		time  float64
		index int
	}{
		{0, 0}, {99.5, 0}, {100, -1}, {120, -1}, {120.5, 1}, {359.5, 1},
		{360, -1}, {380.5, 2}, {480, -1}, {500.5, 3}, {739.5, 3},
		{740, -1}, {760.5, 4}, {999.5, 4}, {1000, -1}, {1000.5, 5},
		{1100, -1}, {1120, -1}, {1120.5, 6}, {1220.5, 6},
	} {
		index, _, active := program.At(check.time)
		if index != check.index || active != (check.index >= 0) {
			t.Fatalf("time %g: stage %d active %v, want %d", check.time, index, active, check.index)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { program.At(760.5) }); allocations != 0 {
		t.Fatalf("cue range sampling allocates %v times", allocations)
	}
}

func TestSteppedEnvelopePreservesPaletteLevels(t *testing.T) {
	envelope, err := NewSteppedEnvelope(SteppedEnvelopeConfig{MaxIndex: 7, EntryLength: 14, ExitLead: 16, Step: 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		time  float64
		shade int
	}{
		{0, 7}, {2, 6}, {12, 1}, {14, 0}, {84, 0}, {86, 1}, {96, 6}, {98, 7}, {100, 7},
	} {
		if got := envelope.At(check.time, 0, 100); got != check.shade {
			t.Fatalf("time %g: shade %d, want %d", check.time, got, check.shade)
		}
	}
	if envelope.At(1200, 1120, math.Inf(1)) != 0 {
		t.Fatal("open-ended stage unexpectedly faded out")
	}
}

func TestCueRangesRejectOverlapAndInvalidEnvelope(t *testing.T) {
	if _, err := NewCueRanges([]CueRange{{Start: 0, End: 10, IncludeEnd: true}, {Start: 10, End: 20}}); err == nil {
		t.Fatal("accepted overlapping inclusive ranges")
	}
	if _, err := NewSteppedEnvelope(SteppedEnvelopeConfig{MaxIndex: 7, Step: 0}); err == nil {
		t.Fatal("accepted zero envelope step")
	}
}
