package motion

import (
	"math"
	"testing"
)

func TestFormulaTrajectoryHasIndependentClocksAndCanPause(t *testing.T) {
	trajectory, err := NewFormulaTrajectory(FormulaTrajectoryConfig{
		X:     ExprAdd(ExprTime(), ExprSecondaryTime()),
		Y:     ExprMul(ExprIndex(), ExprWidth()),
		Start: [2]float64{3, .5}, Step: [2]float64{2, .75},
		Period: [2]float64{0, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := trajectory.At(2, 4, 0, 0); got != (Point{X: 3.5, Y: 8}) {
		t.Fatalf("initial pose: %+v", got)
	}
	if got := trajectory.At(2, 4, 0, 0); got != (Point{X: 3.5, Y: 8}) {
		t.Fatalf("sampling unexpectedly advanced clocks: %+v", got)
	}
	trajectory.Step()
	if got := trajectory.At(2, 4, 0, 0); got != (Point{X: 5.25, Y: 8}) {
		t.Fatalf("wrapped secondary phase: %+v", got)
	}
	if err := trajectory.SetStep([2]float64{-1, .5}); err != nil {
		t.Fatal(err)
	}
	trajectory.Step()
	if got := trajectory.Clocks(); got != ([2]float64{4, .75}) {
		t.Fatalf("changed clock speeds: %+v", got)
	}
	trajectory.Reset()
	if got := trajectory.Clocks(); got != ([2]float64{3, .5}) {
		t.Fatalf("reset phases: %+v", got)
	}
}

func TestFormulaTrajectoryRejectsInvalidClockValues(t *testing.T) {
	for _, config := range []FormulaTrajectoryConfig{
		{X: ExprTime(), Y: ExprSecondaryTime(), Step: [2]float64{math.NaN(), 0}},
		{X: ExprTime(), Y: ExprSecondaryTime(), Period: [2]float64{0, -1}},
	} {
		if _, err := NewFormulaTrajectory(config); err == nil {
			t.Fatalf("accepted invalid clock configuration: %+v", config)
		}
	}
}
