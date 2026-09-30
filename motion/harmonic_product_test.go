package motion

import (
	"math"
	"testing"
)

func TestIndexedHarmonicExplicitProductRounding(t *testing.T) {
	amplitude, phase := 3.1, .57
	wave := math.Sin(phase)
	product := float64(amplitude * wave)
	term := IndexedHarmonic{Amplitude: amplitude, Phase: phase}
	config := HarmonicFormationConfig{Origin: Point{X: -product}, X: []IndexedHarmonic{term}}
	ordinary, err := NewHarmonicFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	config.X[0].RoundProduct = true
	rounded, err := NewHarmonicFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	first := ordinary.At(0, [2]float64{}, 1).X
	second := rounded.At(0, [2]float64{}, 1).X
	if second != 0 {
		t.Fatalf("separately stored wave result did not cancel its exact product: %v", second)
	}
	// FMA is permitted but not required by Go. Keep this check portable while
	// exposing its rounding difference on architectures that fuse the default.
	fused := math.FMA(amplitude, wave, -product)
	if fused == 0 || first != 0 && first != fused {
		t.Fatalf("default=%v explicit=%v fused=%v", first, second, fused)
	}
	t.Logf("default product=%g; explicit product=%g; FMA residual=%g", first, second, fused)
}

func TestIndexedHarmonicExplicitFusedIndexPhase(t *testing.T) {
	index, indexRate := 3, .17
	product := float64(float64(index) * indexRate)
	term := IndexedHarmonic{Amplitude: 1, Rate: 1, IndexRate: indexRate}
	config := HarmonicFormationConfig{X: []IndexedHarmonic{term}}
	ordinary, err := NewHarmonicFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	config.X[0].FusedIndexPhase = true
	fused, err := NewHarmonicFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	first := ordinary.At(index, [2]float64{-product}, 1).X
	second := fused.At(index, [2]float64{-product}, 1).X
	want := math.Sin(math.FMA(float64(index), indexRate, -product))
	if first != 0 || want == 0 || second != want {
		t.Fatalf("default phase=%g; fused phase=%g; want explicit FMA %g", first, second, want)
	}
}
