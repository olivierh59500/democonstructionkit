package motion

import (
	"math"
	"testing"
)

func TestHarmonicFormationUsesIndependentClocksEnvelopeAndBounds(t *testing.T) {
	terms := []IndexedHarmonic{{Amplitude: 4, Rate: 2, IndexPhase: .25, Cos: true}, {Amplitude: 3, IndexRate: math.Pi / 2, Cos: true}}
	bounds := &FormationBounds{Min: Point{X: 0, Y: 0}, Max: Point{X: 20, Y: 40}}
	formation, err := NewHarmonicFormation(HarmonicFormationConfig{
		Origin: Point{X: 10, Y: 20}, Spacing: Point{X: 2, Y: 3}, X: terms,
		Y:      []IndexedHarmonic{{Amplitude: 2, Rate: 1.5, SecondaryClock: true, Envelope: true, Cos: true}},
		Bounds: bounds,
	})
	if err != nil {
		t.Fatal(err)
	}
	terms[0].Amplitude = 100
	bounds.Max.X = 1
	got := formation.At(1, [2]float64{.4, .8}, 5)
	wantX := 12 + 4*math.Cos((.4+.25)*2) + 3*math.Cos(math.Pi/2)
	wantY := 23 + 10*math.Cos(.8*1.5)
	if math.Abs(got.X-wantX) > 1e-12 || math.Abs(got.Y-wantY) > 1e-12 {
		t.Fatalf("position = %+v, want (%v, %v)", got, wantX, wantY)
	}
	if clipped := formation.At(30, [2]float64{0, 0}, 100); clipped.X != 20 || clipped.Y != 40 {
		t.Fatalf("bounds did not apply after formation: %+v", clipped)
	}
	if allocations := testing.AllocsPerRun(100, func() { formation.At(1, [2]float64{.4, .8}, 5) }); allocations != 0 {
		t.Fatalf("formation sample allocates %v times", allocations)
	}
}

func TestHarmonicFormationRejectsInvalidGeometry(t *testing.T) {
	for _, config := range []HarmonicFormationConfig{
		{X: []IndexedHarmonic{{Amplitude: math.NaN()}}},
		{Bounds: &FormationBounds{Min: Point{X: 5}, Max: Point{X: 4}}},
		{Spacing: Point{Y: math.Inf(1)}},
		{X: []IndexedHarmonic{{Divisor: 10, Rate: 1}}},
		{X: []IndexedHarmonic{{UseIndexOffsets: true}}},
	} {
		if _, err := NewHarmonicFormation(config); err == nil {
			t.Fatalf("accepted invalid formation: %+v", config)
		}
	}
}

func TestHarmonicFormationUsesAuthoredIndexPhasesAndExactDivision(t *testing.T) {
	offsets := []float64{.2, 1.4}
	formation, err := NewHarmonicFormation(HarmonicFormationConfig{
		Origin: Point{Y: 75}, IndexOffsets: offsets,
		Y: []IndexedHarmonic{{Amplitude: -1, Divisor: 10, Cos: true, SecondaryClock: true, Envelope: true, UseIndexOffsets: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	offsets[1] = 100
	got := formation.At(1, [2]float64{0, 1.2}, 40)
	want := 75 - 40*math.Cos((1.2+1.4)/10)
	if math.Abs(got.Y-want) > 1e-12 || formation.IndexOffsetCount() != 2 {
		t.Fatalf("indexed phase = %v, want %v", got.Y, want)
	}
}
