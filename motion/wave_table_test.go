package motion

import (
	"math"
	"testing"
)

func TestWaveTableSectionsAndValidation(t *testing.T) {
	got, err := CompileWaveTable(
		WaveSection{Samples: 2, Offset: 3},
		WaveSection{Samples: 4, Offset: 2, Terms: []WaveTerm{{Amplitude: 4, Step: .3}, {Amplitude: 7, Step: .2, Phase: -1, Cosine: true}}},
		WaveSection{Samples: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 7 || got[0] != 3 || got[1] != 3 || got[6] != 0 {
		t.Fatalf("unexpected table %v", got)
	}
	for i := 0; i < 4; i++ {
		want := 2 + 4*math.Sin(float64(i)*.3) + 7*math.Cos(float64(i)*.2-1)
		if got[i+2] != want {
			t.Fatalf("sample %d: %g != %g", i, got[i+2], want)
		}
	}
	for _, sections := range [][]WaveSection{nil, {{Samples: -1}}, {{Samples: 1<<20 + 1}}, {{Samples: 1, Offset: math.NaN()}}, {{Samples: 1, Terms: []WaveTerm{{Step: math.Inf(1)}}}}, {{Samples: 1, Offset: math.MaxFloat64, Terms: []WaveTerm{{Amplitude: math.MaxFloat64, Cosine: true}}}}} {
		if _, err := CompileWaveTable(sections...); err == nil {
			t.Fatalf("accepted invalid sections %+v", sections)
		}
	}
}
