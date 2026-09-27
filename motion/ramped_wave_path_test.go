package motion

import (
	"math"
	"testing"
)

func TestRampedWavePathMatchesUnionPointerAtBothRates(t *testing.T) {
	path, err := NewRampedWavePath(RampedWavePathConfig{
		Base: Point{X: 384, Y: 268},
		X:    Wave{Amplitude: 220, Speed: .8}, Y: Wave{Amplitude: 140, Speed: 1.1},
		Rise: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, rate := range []int{50, 60} {
		for tick := 0; tick <= 3600; tick++ {
			seconds := float64(tick) / float64(rate)
			gain := math.Min(1, seconds/2)
			want := Point{X: 384 + gain*220*math.Sin(seconds*.8),
				Y: 268 + gain*140*math.Sin(seconds*1.1)}
			if got := path.At(seconds); math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
				t.Fatalf("%d Hz tick %d: point %+v, want %+v", rate, tick, got, want)
			}
		}
	}
}

func TestRampedWavePathSupportsIndependentSpatialPhases(t *testing.T) {
	path, err := NewRampedWavePath(RampedWavePathConfig{
		Base: Point{X: 10, Y: 20}, Sample: Point{X: 2, Y: 4},
		X:     Wave{Amplitude: 8, Spatial: math.Pi / 4},
		Y:     Wave{Amplitude: 6, Spatial: math.Pi / 8, Cos: true},
		Start: 1, Rise: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := path.At(0); got != (Point{X: 10, Y: 20}) {
		t.Fatalf("pre-ramp pose = %+v", got)
	}
	if got := path.At(2); math.Abs(got.X-14) > 1e-12 || math.Abs(got.Y-20) > 1e-12 {
		t.Fatalf("half-ramp pose = %+v", got)
	}
}

func TestRampedWavePathRejectsInvalidParameters(t *testing.T) {
	for _, config := range []RampedWavePathConfig{
		{Rise: -1}, {Rise: math.Inf(1)},
		{X: Wave{Speed: math.NaN()}},
		{Base: Point{Y: math.Inf(-1)}},
	} {
		if _, err := NewRampedWavePath(config); err == nil {
			t.Fatalf("accepted %+v", config)
		}
	}
}
