package motion

import (
	"math"
	"testing"
)

func TestHarmonicRowProfilePreservesGroupedAndOrderedStages(t *testing.T) {
	large, opposite := IndexedHarmonic{Amplitude: 1e16, Cos: true}, IndexedHarmonic{Amplitude: -1e16, Cos: true}
	grouped, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 1,
		Stages: []HarmonicRowStage{{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{large, opposite}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 1,
		Stages: []HarmonicRowStage{
			{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{large}}},
			{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{opposite}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	first, firstOK := grouped.Apply(0, 0, Point{X: 1})
	second, secondOK := ordered.Apply(0, 0, Point{X: 1})
	if !firstOK || !secondOK || first.X != 1 || second.X != 0 {
		t.Fatalf("grouped %v, ordered %v; grouping must preserve floating-point association", first, second)
	}
}

func TestHarmonicRowProfileCopiesConfigAndSharesFixedStage(t *testing.T) {
	index := 2
	terms := []IndexedHarmonic{{Amplitude: 3, Cos: true}}
	p, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 5, ClockScale: [2]float64{1},
		Stages: []HarmonicRowStage{
			{Motion: HarmonicFormationConfig{Spacing: Point{X: 2}}, Index: &index},
			{Motion: HarmonicFormationConfig{X: terms, Spacing: Point{Y: 4}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	index, terms[0].Amplitude = 9, 99
	if len(p.stages[0].values) != 1 || len(p.stages[1].values) != 5 || p.Len() != 5 {
		t.Fatal("fixed stages should retain one cached position")
	}
	for row := 0; row < 5; row++ {
		got, ok := p.Apply(row, 1, Point{X: 10, Y: 20})
		if !ok || got != (Point{X: 17, Y: 20 + float64(row*4)}) {
			t.Fatalf("row %d: %v", row, got)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { p.Apply(3, 1, Point{X: 10}) }); allocations != 0 {
		t.Fatalf("cached sampling allocated %v times", allocations)
	}
	seconds := 0.0
	if allocations := testing.AllocsPerRun(100, func() { seconds += .02; p.Apply(3, seconds, Point{X: 10}) }); allocations != 0 {
		t.Fatalf("new-time sampling allocated %v times", allocations)
	}
}

func TestHarmonicRowProfileNativeGlobalSpacingAndLocalWaves(t *testing.T) {
	global := 0
	wave := func(amplitude, rate, spacing float64) IndexedHarmonic {
		return IndexedHarmonic{Amplitude: amplitude, Rate: rate, IndexRate: spacing, PhasePeriod: 2 * math.Pi,
			RoundProduct: true, FusedIndexPhase: true}
	}
	p, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 8, ClockScale: [2]float64{85},
		Stages: []HarmonicRowStage{
			{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{wave(70, .0067, 0), wave(70, .0239, 0)},
				Y: []IndexedHarmonic{wave(30, .0097, 0), wave(30, .0339, 0)}}, Index: &global},
			{Motion: HarmonicFormationConfig{Spacing: Point{X: -30}}},
			{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{wave(-30, .103, .222)}, Y: []IndexedHarmonic{wave(10, .1, .17)}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// The independent source equations retain their original grouping, signed
	// phase reduction, and positive-amplitude subtraction on the row's Y axis.
	original := func(seconds, amplitude, rate, spacing float64, row int) float64 {
		clockPhase := float64(seconds * 85 * rate)
		phase := math.FMA(float64(row), spacing, clockPhase)
		return float64(amplitude * math.Sin(math.Mod(phase, 2*math.Pi)))
	}
	for tick := -50; tick <= 9000; tick++ {
		seconds := float64(tick) / 50
		globalY := original(seconds, 70, .0067, 0, 0) + original(seconds, 70, .0239, 0, 0)
		globalZ := original(seconds, 30, .0097, 0, 0) + original(seconds, 30, .0339, 0, 0)
		for row := 0; row < 8; row++ {
			want := Point{X: 120 + globalY - float64(row*30) - original(seconds, 30, .103, .222, row),
				Y: 140 + globalZ + original(seconds, 10, .1, .17, row)}
			got, ok := p.Apply(row, seconds, Point{X: 120, Y: 140})
			if !ok || got != want {
				t.Fatalf("tick %d row %d: %v; want %v", tick, row, got, want)
			}
		}
	}
}

func TestHarmonicRowProfileValidationAndRejectedSamples(t *testing.T) {
	fixed := 1_000_001
	for _, config := range []HarmonicRowProfileConfig{
		{}, {Count: 65537, Stages: []HarmonicRowStage{{}}},
		{Count: 1, Stages: make([]HarmonicRowStage, 65)},
		{Count: 65536, Stages: make([]HarmonicRowStage, 64)},
		{Count: 1, ClockScale: [2]float64{math.NaN()}, Stages: []HarmonicRowStage{{}}},
		{Count: 1, Stages: []HarmonicRowStage{{Index: &fixed}}},
		{Count: 1, Stages: []HarmonicRowStage{{Motion: HarmonicFormationConfig{Origin: Point{X: math.Inf(1)}}}}},
	} {
		if _, err := NewHarmonicRowProfile(config); err == nil {
			t.Fatalf("invalid config accepted: %+v", config)
		}
	}
	p, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 2, ClockScale: [2]float64{2},
		Stages: []HarmonicRowStage{{Motion: HarmonicFormationConfig{Spacing: Point{X: math.MaxFloat64}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		row     int
		seconds float64
		base    Point
	}{
		{-1, 0, Point{}}, {2, 0, Point{}}, {0, math.NaN(), Point{}},
		{0, math.MaxFloat64, Point{}}, {0, 0, Point{X: math.NaN()}},
		{1, 0, Point{X: math.MaxFloat64}},
	} {
		if _, ok := p.Apply(input.row, input.seconds, input.base); ok {
			t.Fatalf("invalid input accepted: %+v", input)
		}
	}
	badWave, err := NewHarmonicRowProfile(HarmonicRowProfileConfig{Count: 1, ClockScale: [2]float64{1},
		Stages: []HarmonicRowStage{{Motion: HarmonicFormationConfig{X: []IndexedHarmonic{{Rate: math.MaxFloat64, Amplitude: 1}}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := badWave.Sample(2); err == nil || badWave.valid {
		t.Fatal("an invalid phase should invalidate the cache")
	}
	if got, ok := badWave.Apply(0, 0, Point{X: 7}); !ok || got.X != 7 {
		t.Fatal("rejected samples should not prevent later valid sampling")
	}
	var absent *HarmonicRowProfile
	if absent.Len() != 0 || absent.Sample(0) == nil {
		t.Fatal("nil profile was not handled")
	}
}
