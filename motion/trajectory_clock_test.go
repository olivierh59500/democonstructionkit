package motion

import (
	"math"
	"testing"
)

func TestTrajectoryClockMatchesFourOriginalOrbitTimings(t *testing.T) {
	tests := []struct {
		name        string
		center      Point
		radius      Point
		start, step float64
	}{
		{"Cuddly Ehhh background", Point{384, 270}, Point{135, 200}, 0, .006},
		{"Cuddly Big Sprite face", Point{320, 200}, Point{160, 400 / 3.7}, 9, .008},
		{"DMA Is Back logo", Point{320, 200}, Point{160, 400 / 2.7}, 0, .014},
		{"Union Level 16 ball", Point{384, 268}, Point{192, 536 / 2.7}, 0, .008},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := DefaultNestedOrbit(test.center, test.radius)
			clock, err := NewTrajectoryClock(TrajectoryClockConfig{Sample: path.At, Start: test.start, Step: test.step})
			if err != nil {
				t.Fatal(err)
			}
			phase := test.start
			for tick := 0; tick < 5000; tick++ {
				// Each source advances the phase before drawing the next frame.
				phase += test.step
				if err := clock.Step(); err != nil {
					t.Fatal(err)
				}
				want, got := path.At(phase), clock.At()
				if clock.Phase() != phase || math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
					t.Fatalf("tick %d: phase %g, point %+v; want phase %g, point %+v", tick, clock.Phase(), got, phase, want)
				}
			}
			if allocations := testing.AllocsPerRun(100, func() {
				if err := clock.Step(); err != nil {
					t.Fatal(err)
				}
				_ = clock.At()
			}); allocations != 0 {
				t.Fatalf("trajectory update allocated %v times", allocations)
			}
		})
	}
}

func TestTrajectoryClockEditsKeepPhaseAndRejectInvalidPoints(t *testing.T) {
	clock, err := NewTrajectoryClock(TrajectoryClockConfig{
		Sample: func(phase float64) Point { return Point{X: phase, Y: phase * phase} },
		Start:  2, Step: .25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := clock.Step(); err != nil {
		t.Fatal(err)
	}
	if err := clock.SetStep(.5); err != nil {
		t.Fatal(err)
	}
	if err := clock.SetSample(func(phase float64) Point { return Point{X: -phase, Y: phase + 10} }); err != nil {
		t.Fatal(err)
	}
	if clock.Phase() != 2.25 || clock.At() != (Point{X: -2.25, Y: 12.25}) {
		t.Fatalf("path replacement lost phase: %g %+v", clock.Phase(), clock.At())
	}
	if err := clock.SetSample(func(float64) Point { return Point{X: math.NaN()} }); err == nil {
		t.Fatal("accepted a nonfinite replacement path")
	}
	if err := clock.Step(); err != nil || clock.Phase() != 2.75 || clock.At() != (Point{X: -2.75, Y: 12.75}) {
		t.Fatalf("invalid path changed a valid clock: phase=%g point=%+v err=%v", clock.Phase(), clock.At(), err)
	}
	if err := clock.SetPhase(math.Inf(1)); err == nil {
		t.Fatal("accepted an infinite phase")
	}
	if err := clock.Reset(); err != nil || clock.Phase() != 2 || clock.At() != (Point{X: -2, Y: 12}) {
		t.Fatalf("reset failed: phase=%g point=%+v err=%v", clock.Phase(), clock.At(), err)
	}
}
