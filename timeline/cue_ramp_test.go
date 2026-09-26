package timeline

import (
	"math"
	"testing"
)

func TestCueRampTracksLoaderFadeAtBothRatesAndAfterRateChange(t *testing.T) {
	for _, rate := range []int{50, 60} {
		clock, err := NewCueClock(CueClockConfig{Rate: rate, Windows: []CueWindow{{StartTick: 169, Duration: 1.5}}})
		if err != nil {
			t.Fatal(err)
		}
		ramp, err := NewCueRamp(clock, CueRampConfig{Window: 0, From: 1, To: 0})
		if err != nil {
			t.Fatal(err)
		}
		for tick := 0; tick < 400; tick++ {
			want := max(0, 1-clock.Elapsed(0)/1.5)
			if got := ramp.Value(); math.Abs(got-want) > 1e-15 {
				t.Fatalf("%d Hz tick %d volume = %g, want %g", rate, tick, got, want)
			}
			if tick == 200 {
				if err := clock.SetRate(60); err != nil {
					t.Fatal(err)
				}
			}
			clock.Step()
		}
		clock.Reset()
		if ramp.Value() != 1 {
			t.Fatalf("%d Hz reset did not restore full volume", rate)
		}
		if allocations := testing.AllocsPerRun(100, func() { _ = ramp.Value() }); allocations != 0 {
			t.Fatalf("%d Hz cue ramp allocated %v times per sample", rate, allocations)
		}
	}
}

func TestCueRampRejectsMissingWindowOrDuration(t *testing.T) {
	clock, err := NewCueClock(CueClockConfig{Rate: 60, Windows: []CueWindow{{StartTick: 0}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, config := range []CueRampConfig{
		{Window: 1, From: 1, To: 0},
		{Window: 0, From: 1, To: 0},
		{Window: 0, From: math.NaN(), To: 0, Duration: 1},
	} {
		if _, err := NewCueRamp(clock, config); err == nil {
			t.Fatalf("accepted invalid ramp %+v", config)
		}
	}
}
