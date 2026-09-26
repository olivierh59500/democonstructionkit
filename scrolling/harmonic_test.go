package scrolling

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicSineRestartsSpatialPhaseWithoutResettingTime(t *testing.T) {
	waves := []motion.Wave{
		{Amplitude: 15, Spatial: .06 / 8, Speed: 18},
		{Amplitude: 15, Spatial: -.04 / 8, Speed: 12},
	}
	mode, err := HarmonicSineWith(HarmonicSineConfig{
		Direction: WaveVertical, SpatialPeriod: 168, Waves: waves,
	})
	if err != nil {
		t.Fatal(err)
	}
	waves[0].Amplitude = 999 // The compiled mode owns its recipe.
	var op ebiten.DrawImageOptions
	if !mode.Map(Sample{Glyph: Glyph{Offset: 168 + 24}, Time: 7.5}, &op) {
		t.Fatal("harmonic mode hid a glyph")
	}
	x, y := op.GeoM.Apply(0, 0)
	want := 15*math.Sin(3*.06+7.5*18) + 15*math.Sin(-3*.04+7.5*12)
	if x != 0 || math.Abs(y-want) > 1e-12 {
		t.Fatalf("wrapped spatial phase: (%.12f, %.12f), want (0, %.12f)", x, y, want)
	}
	horizontal, err := HarmonicSine(WaveHorizontal, motion.Wave{Amplitude: 4, Phase: math.Pi / 2})
	if err != nil {
		t.Fatal(err)
	}
	op = ebiten.DrawImageOptions{}
	horizontal.Map(Sample{}, &op)
	x, y = op.GeoM.Apply(0, 0)
	if math.Abs(x-4) > 1e-12 || y != 0 {
		t.Fatalf("horizontal wave: (%.12f, %.12f)", x, y)
	}
	if _, err := HarmonicSineWith(HarmonicSineConfig{SpatialPeriod: -1, Waves: waves}); err == nil {
		t.Fatal("accepted negative spatial period")
	}
}
