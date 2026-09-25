package motion

import (
	"math"
	"testing"
)

func TestCuedFormationStaggersAndLoopsWithoutPositionJumps(t *testing.T) {
	formation, err := NewCuedFormation(CuedFormationConfig{
		Origin: Point{X: 10, Y: 20}, Spacing: Point{X: 30}, Count: 3, Loop: 4,
		Cues: []FormationCue{{Start: 0, Duration: 1, Stagger: 1, LeadIndex: 2, Fade: .1,
			X: []FormationHarmonic{{FirstAmplitude: 100, LastAmplitude: 100, Cycles: .5}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for index, at := range []float64{2.5, 1.5, .5} {
		got := formation.At(at, index)
		if math.Abs(got.X-(10+float64(index)*30+100)) > 1e-9 || got.Y != 20 {
			t.Fatalf("item %d at %v: %+v", index, at, got)
		}
		if before := formation.At(at-1.5, index); before.X != 10+float64(index)*30 {
			t.Fatalf("item %d moved before its cue: %+v", index, before)
		}
		if next := formation.At(at+4, index); next != got {
			t.Fatalf("item %d loop changed pose: %+v versus %+v", index, next, got)
		}
	}
	for _, at := range []float64{0, 1, 2, 3, 4} {
		for index := 0; index < 3; index++ {
			got := formation.At(at, index)
			if got.X != 10+float64(index)*30 || got.Y != 20 {
				t.Fatalf("cue boundary jumped at %v for item %d: %+v", at, index, got)
			}
		}
	}
}

func TestCuedFormationCopiesCuesAndRejectsOverflow(t *testing.T) {
	cues := []FormationCue{{Start: .2, Duration: .8, LeadIndex: 0, X: []FormationHarmonic{{FirstAmplitude: 20, LastAmplitude: 20, Cycles: .5}}}}
	formation, err := NewCuedFormation(CuedFormationConfig{Count: 1, Loop: 2, Cues: cues})
	if err != nil {
		t.Fatal(err)
	}
	initial := formation.At(.6, 0)
	cues[0].X[0].FirstAmplitude = 200
	if got := formation.At(.6, 0); got != initial {
		t.Fatalf("caller changed compiled formation: %+v versus %+v", got, initial)
	}
	_, err = NewCuedFormation(CuedFormationConfig{Count: 2, Loop: 2, Cues: []FormationCue{{Start: 1, Duration: 1, Stagger: .1, LeadIndex: 1}}})
	if err == nil {
		t.Fatal("cue spilling across loop boundary was accepted")
	}
}
