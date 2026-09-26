package motion

import (
	"math"
	"testing"
)

func TestRecurrentTranslationMatchesCocoLogoClocks(t *testing.T) {
	config := RecurrentTranslationConfig{
		Harmonics: HarmonicTranslation{
			X: []HarmonicTerm{{Amplitude: 100, Rate: 1.35, Phase: 1.25}, {Amplitude: 100, Rate: 1.86, Phase: .54}},
			Y: []HarmonicTerm{{Amplitude: 60, Rate: 1.72, Phase: .23, Cos: true}, {Amplitude: 60, Rate: 1.63, Phase: .98, Cos: true}},
		},
		PhaseStep: .02, ReanchorEvery: 1024,
		XStepDeltas: []float64{.02 * 1.35, .02 * 1.86},
		YStepDeltas: []float64{.02 * 1.72, .02 * 1.63},
	}
	program, err := NewRecurrentTranslation(config)
	if err != nil {
		t.Fatal(err)
	}
	initial := program.At()
	starts := [4]float64{1.25, .54, .23, .98}
	deltas := [4]float64{.02 * 1.35, .02 * 1.86, .02 * 1.72, .02 * 1.63}
	var sin, cos, sinStep, cosStep [4]float64
	for i := range starts {
		sin[i], cos[i] = math.Sincos(starts[i])
		sinStep[i], cosStep[i] = math.Sincos(deltas[i])
	}
	for tick := 1; tick <= 10_000; tick++ {
		for i := range sin {
			if tick&1023 == 0 {
				phase := starts[i] + float64(tick)*deltas[i]
				sin[i], cos[i] = math.Sincos(math.Mod(phase, 2*math.Pi))
			} else {
				sin[i], cos[i] = sin[i]*cosStep[i]+cos[i]*sinStep[i], cos[i]*cosStep[i]-sin[i]*sinStep[i]
			}
		}
		want := Point{X: 100*sin[0] + 100*sin[1], Y: 60*cos[2] + 60*cos[3]}
		if got := program.Step(); got != want || program.Tick() != uint64(tick) {
			t.Fatalf("tick %d translation %+v, want %+v", tick, got, want)
		}
	}
	program.Reset()
	if program.Tick() != 0 || program.At() != initial {
		t.Fatalf("reset translation %+v at tick %d", program.At(), program.Tick())
	}
	if allocations := testing.AllocsPerRun(100, func() { program.Step() }); allocations != 0 {
		t.Fatalf("recurrent translation allocates %v times per step", allocations)
	}
}

func TestRecurrentTranslationRejectsInvalidTerms(t *testing.T) {
	for _, config := range []RecurrentTranslationConfig{
		{PhaseStep: math.NaN()},
		{Harmonics: HarmonicTranslation{X: []HarmonicTerm{{Amplitude: math.Inf(1)}}}},
		{PhaseStep: math.MaxFloat64, Harmonics: HarmonicTranslation{Y: []HarmonicTerm{{Rate: 2}}}},
	} {
		if _, err := NewRecurrentTranslation(config); err == nil {
			t.Fatalf("accepted invalid recurrence %+v", config)
		}
	}
}
