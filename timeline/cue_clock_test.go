package timeline

import (
	"math"
	"testing"
	"time"
)

func TestCueClockPreservesOverlappingCountdownAndRateSwitch(t *testing.T) {
	for _, initialRate := range []int{50, 60} {
		clock, err := NewCueClock(CueClockConfig{Rate: initialRate, Windows: []CueWindow{
			{StartTick: 235, Duration: .25, Tolerance: 1e-9},
			{StartTick: 169, Duration: 1.5},
		}})
		if err != nil {
			t.Fatal(err)
		}
		tick, rate := 0, initialRate
		blackElapsed, fadeElapsed := 0.0, 0.0
		for step := 0; step < 500; step++ {
			if step == 240 {
				if rate == 50 {
					rate = 60
				} else {
					rate = 50
				}
				if err := clock.SetRate(rate); err != nil {
					t.Fatal(err)
				}
			}
			delta := 1 / float64(rate)
			if tick >= 235 {
				blackElapsed += delta
			}
			if tick >= 169 {
				fadeElapsed += delta
			}
			tick++
			clock.Step()
			if clock.Tick() != tick || clock.Elapsed(0) != blackElapsed || clock.Elapsed(1) != fadeElapsed || clock.Done(0) != (tick >= 235 && blackElapsed+1e-9 >= .25) {
				t.Fatalf("%d Hz step %d: cue clock diverged", initialRate, step)
			}
		}
		clock.Reset()
		if clock.Tick() != 0 || clock.Elapsed(0) != 0 || clock.Elapsed(1) != 0 {
			t.Fatalf("%d Hz reset kept state", initialRate)
		}
		if allocations := testing.AllocsPerRun(100, clock.Step); allocations != 0 {
			t.Fatalf("cue clock step allocates %v times", allocations)
		}
	}
}

func TestCueClockDurationMatchesRationalTickComparison(t *testing.T) {
	duration := 3777120 * time.Microsecond
	for _, rate := range []int{50, 60} {
		clock, err := NewCueClock(CueClockConfig{Rate: rate})
		if err != nil {
			t.Fatal(err)
		}
		for tick := 0; tick < 300; tick++ {
			want := time.Duration(tick)*time.Second >= duration*time.Duration(rate)
			if got := clock.Reached(duration); got != want {
				t.Fatalf("%d Hz tick %d reached = %v, want %v", rate, tick, got, want)
			}
			clock.Step()
		}
	}
}

func TestCueClockRejectsInvalidWindows(t *testing.T) {
	for _, config := range []CueClockConfig{
		{Rate: 0}, {Rate: 60, Windows: []CueWindow{{Duration: -1}}},
		{Rate: 60, Windows: []CueWindow{{Tolerance: math.NaN()}}},
	} {
		if _, err := NewCueClock(config); err == nil {
			t.Fatalf("accepted invalid cue clock: %+v", config)
		}
	}
}
