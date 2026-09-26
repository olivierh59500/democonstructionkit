package motion

import (
	"math"
	"testing"
)

func TestPairedPhaseProgramMatchesResetRasterOrder(t *testing.T) {
	program, err := NewPairedPhaseProgram(PairedPhaseConfig{
		Count: 16, Spacing: .25, Step: .05, Period: 2 * math.Pi,
		CenterY: 270, AmplitudeY: 200,
		Passes: []PhasePass{
			{From: math.Pi, To: 2 * math.Pi, Material: 0, Alpha: .5},
			{From: 0, To: math.Pi / 2, IncludeFrom: true, IncludeTo: true, Material: 1, Alpha: 1},
			{From: math.Pi / 2, To: math.Pi, IncludeTo: true, Reverse: true, Material: 1, Alpha: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	phases := make([]float64, 16)
	for i := range phases {
		phases[i] = float64(i) * .25
	}
	want := make([]PhasePairSample, 0, 48)
	appendPair := func(i, material int, alpha float64) {
		want = append(want, PhasePairSample{Index: i, Material: material,
			Y: 270 + 200*math.Cos(phases[i]), Alpha: alpha})
	}
	for tick := 1; tick <= 1000; tick++ {
		want = want[:0]
		for i, a := range phases {
			if a > math.Pi && a < 2*math.Pi {
				appendPair(i, 0, .5)
			}
		}
		for i, a := range phases {
			if a >= 0 && a <= math.Pi/2 {
				appendPair(i, 1, 1)
			}
		}
		for i := len(phases) - 1; i >= 0; i-- {
			a := phases[i]
			if a > math.Pi/2 && a <= math.Pi {
				appendPair(i, 1, 1)
			}
			phases[i] += .05
			if phases[i] >= 2*math.Pi {
				phases[i] -= 2 * math.Pi
			}
		}
		program.Step()
		if program.Tick() != tick || len(program.Samples()) != len(want) {
			t.Fatalf("tick %d: step=%d samples=%d, want %d", tick, program.Tick(), len(program.Samples()), len(want))
		}
		for i := range want {
			if program.Samples()[i] != want[i] {
				t.Fatalf("tick %d sample %d = %+v, want %+v", tick, i, program.Samples()[i], want[i])
			}
		}
		for i, phase := range phases {
			if program.Phases()[i] != phase {
				t.Fatalf("tick %d phase %d = %v, want %v", tick, i, program.Phases()[i], phase)
			}
		}
	}
	if got := testing.AllocsPerRun(100, program.Step); got != 0 {
		t.Fatalf("paired phase step allocated %.2f objects", got)
	}
}
