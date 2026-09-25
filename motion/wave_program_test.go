package motion

import (
	"math"
	"testing"
)

func TestWaveProgramAppendsOverwritesAndRetainsSourcePhase(t *testing.T) {
	got, err := CompileWaveProgram(
		WaveWrite{At: 0, Section: WaveSection{Samples: 2, Offset: 8}},
		WaveWrite{At: 4, Section: WaveSection{Samples: 2, Terms: []WaveTerm{{Amplitude: 4, Step: .5}}}},
		WaveWrite{At: 1, Section: WaveSection{Samples: 2, SampleStart: 3, Terms: []WaveTerm{{Amplitude: 7, Step: .2}}}},
		WaveWrite{At: WaveAppend, Section: WaveSection{Samples: 1, Offset: -2}},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{8, 0, 0, 0, 0, 4 * math.Sin(.5), -2}
	for i := 1; i <= 2; i++ {
		want[i] = 7 * math.Sin(float64(i+2)*.2)
	}
	if len(got) != len(want) {
		t.Fatalf("length %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sample %d = %v, want %v", i, got[i], want[i])
		}
	}
	for _, writes := range [][]WaveWrite{
		nil,
		{{At: -2, Section: WaveSection{Samples: 1}}},
		{{At: 1 << 20, Section: WaveSection{Samples: 1}}},
		{{At: 0, Section: WaveSection{Samples: 0}}},
	} {
		if _, err := CompileWaveProgram(writes...); err == nil {
			t.Fatalf("accepted invalid writes %+v", writes)
		}
	}
}
