package motion

import (
	"math"
	"testing"
)

func TestWaveSelectsSineOrCosineWithoutChangingItsClock(t *testing.T) {
	sine := Wave{Amplitude: 44, Spatial: .3, Speed: .06, Phase: .3}
	cosine := sine
	cosine.Cos = true
	for _, sample := range []struct{ index, tick float64 }{{0, 0}, {3, 1}, {6, 120}} {
		phase := sample.index*.3 + sample.tick*.06 + .3
		if got := sine.At(sample.index, sample.tick); math.Abs(got-44*math.Sin(phase)) > 1e-12 {
			t.Fatalf("sine at %+v = %v", sample, got)
		}
		if got := cosine.At(sample.index, sample.tick); math.Abs(got-44*math.Cos(phase)) > 1e-12 {
			t.Fatalf("cosine at %+v = %v", sample, got)
		}
	}
}
