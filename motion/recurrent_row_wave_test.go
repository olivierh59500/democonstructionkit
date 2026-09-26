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
