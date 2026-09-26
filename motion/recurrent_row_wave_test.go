package motion

import (
	"math"
	"testing"
)

func TestRecurrentRowWaveMatchesPhenomenaAtSkippedStripsAndReveal(t *testing.T) {
	c := RecurrentRowWaveConfig{
		Base: 67, Flat: 80, Amplitude: 80,
		RevealTime: 5 * 50, IndexDelay: .0033, SampleTimeStep: 1.0 / 6.0,
		StartAngle: 5 * 10.50, TimeDivisor: 6, AngleStep: 1.0 / 36.0,
	}
	wave, err := NewRecurrentRowWave(c)
	if err != nil {
		t.Fatal(err)
	}
	stepSin, stepCos := math.Sincos(1.0 / 36.0)
	for _, time := range []float64{0, 249, 249.5, 250, 250.5, 400, 800, 4800} {
		if err := wave.Begin(time); err != nil {
			t.Fatal(err)
		}
		t2 := time
		sin, cos := math.Sincos(5*10.50 + time/6)
		for i := 0; i < 240; i++ {
			if i%7 == 0 || i%11 == 4 {
				continue // The renderer does not call Y for an invalid source slice.
			}
			y := 80.0
			if t2 > 5*50-float64(i)*.0033 {
				y = 80 * cos
			}
			t2 += 1.0 / 6.0
			sin, cos = sin*stepCos+cos*stepSin, cos*stepCos-sin*stepSin
			want := 67 + y
			if got := wave.At(i); got != want {
				t.Fatalf("time=%v strip=%d got=%v want=%v", time, i, got, want)
			}
		}
	}
	if got := testing.AllocsPerRun(100, func() {
		_ = wave.Begin(500)
		for i := 0; i < 240; i++ {
			wave.At(i)
		}
	}); got != 0 {
		t.Fatalf("row wave draw allocated %.2f objects", got)
	}
}

func TestRecurrentRowWaveOwnedFrameClockMatchesExternalSchedule(t *testing.T) {
	c := RecurrentRowWaveConfig{
		Base: 67, Flat: 80, Amplitude: 80,
		RevealTime: 250, IndexDelay: .0033, SampleTimeStep: 1.0 / 6.0,
		StartAngle: 52.5, TimeDivisor: 6, AngleStep: 1.0 / 36.0,
		FrameStep: .30,
	}
	owned, err := NewRecurrentRowWave(c)
	if err != nil {
		t.Fatal(err)
	}
	external, err := NewRecurrentRowWave(c)
	if err != nil {
		t.Fatal(err)
	}
	frameTime := 0.0
	step := .30
	for frame := 0; frame < 5000; frame++ {
		if frame == 2500 {
			step = .45
			if err := owned.SetFrameStep(step); err != nil {
				t.Fatal(err)
			}
		}
		if frame > 0 {
			frameTime += step
			if err := owned.AdvanceFrame(); err != nil {
				t.Fatal(err)
			}
		}
		if owned.FrameTime() != frameTime {
			t.Fatalf("frame %d time = %g, want %g", frame, owned.FrameTime(), frameTime)
		}
		if err := owned.BeginFrame(); err != nil {
			t.Fatal(err)
		}
		if err := external.Begin(frameTime); err != nil {
			t.Fatal(err)
		}
		for row := 0; row < 120; row++ {
			if row%7 == 0 {
				continue
			}
			if got, want := owned.At(row), external.At(row); got != want {
				t.Fatalf("frame %d row %d = %g, want %g", frame, row, got, want)
			}
		}
	}
	if err := owned.SetFrameStep(math.NaN()); err == nil {
		t.Fatal("accepted nonfinite frame step")
	}
	if err := owned.SetFrameTime(7.5); err != nil || owned.FrameTime() != 7.5 {
		t.Fatalf("could not seek scene time: %g, %v", owned.FrameTime(), err)
	}
	owned.ResetFrame()
	if owned.FrameTime() != c.FrameStart {
		t.Fatalf("reset time = %g, want %g", owned.FrameTime(), c.FrameStart)
	}
	if got := testing.AllocsPerRun(100, func() {
		_ = owned.AdvanceFrame()
		_ = owned.BeginFrame()
		for row := 0; row < 120; row++ {
			owned.At(row)
		}
	}); got != 0 {
		t.Fatalf("owned row-wave frame allocated %v times", got)
	}
}
