package timeline

import (
	"math"
	"testing"
)

func TestStageSequencePreservesSameTickHandoffsAndTwoFades(t *testing.T) {
	sequence, err := NewStageSequence(StageSequenceConfig{
		Stages: EventStagesConfig{
			Stages: []string{"first", "second", "show", "final"}, Initial: "first",
			Transitions: []StageTransition{{"first", "next", "second"}, {"second", "next", "show"}, {"show", "next", "final"}},
		},
		TimedStage: "show", Rate: 1,
		Windows: []CueRange{{Start: 0, End: 41}, {Start: 41, End: 81}, {Start: 81, End: math.Inf(1)}},
		Fades:   []HoldRampConfig{{Frames: 40}, {Frames: 40}, {}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sequence.Trigger("next") || !sequence.Active("second") || !sequence.Trigger("next") || !sequence.Active("show") {
		t.Fatal("same-tick stage cascade failed")
	}
	for tick := 0; tick <= 82; tick++ {
		index, alpha := sequence.Window()
		wantIndex, wantAlpha := 0, 0.0
		switch {
		case tick <= 40:
			wantAlpha = float64(tick) / 40
		case tick <= 80:
			wantIndex, wantAlpha = 1, float64(tick-41)/40
		default:
			wantIndex, wantAlpha = 2, 1
		}
		if index != wantIndex || alpha != wantAlpha {
			t.Fatalf("tick %d = window %d alpha %v, want %d/%v", tick, index, alpha, wantIndex, wantAlpha)
		}
		sequence.StepWindow()
	}
	sequence.Trigger("next")
	if !sequence.Active("final") {
		t.Fatal("final stage did not become active immediately")
	}
	sequence.Reset()
	if !sequence.Active("first") || sequence.Tick() != 0 {
		t.Fatal("reset kept stage or tick")
	}
	sequence.Trigger("next")
	sequence.Trigger("next")
	if allocations := testing.AllocsPerRun(100, sequence.StepWindow); allocations != 0 {
		t.Fatalf("stage window step allocates %v times", allocations)
	}
}
