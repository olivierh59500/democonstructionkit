package motion

import (
	"math"
	"testing"
)

func TestKeyframedFormationInterpolatesIndependentPathsAndLoops(t *testing.T) {
	frames := []FormationFrame{
		{Time: 0, Points: []Point{{X: 10, Y: 20}, {X: 100, Y: 40}}, Ease: Smooth},
		{Time: 1, Points: []Point{{X: 30, Y: 60}, {X: 60, Y: 0}}},
		{Time: 2, Points: []Point{{X: 10, Y: 20}, {X: 100, Y: 40}}},
	}
	formation, err := NewKeyframedFormation(KeyframedFormationConfig{Count: 2, Loop: 2, Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	if got := formation.At(.25, 0); got != (Point{X: 13.125, Y: 26.25}) {
		t.Fatalf("first eased sprite = %+v", got)
	}
	if got := formation.At(.25, 1); got != (Point{X: 93.75, Y: 33.75}) {
		t.Fatalf("second sprite lost its own path: %+v", got)
	}
	if got := formation.At(1.5, 1); got != (Point{X: 80, Y: 20}) {
		t.Fatalf("linear return path = %+v", got)
	}
	frames[1].Points[0].X = 999
	if got := formation.At(1, 0); got.X != 30 {
		t.Fatalf("caller changed compiled formation: %+v", got)
	}
	if got := formation.At(2, 1); got != formation.At(0, 1) {
		t.Fatalf("loop position jumped: %+v", got)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = formation.At(.25, 1) }); allocations != 0 {
		t.Fatalf("sampling allocated %g times per sprite", allocations)
	}
}

func TestKeyframedFormationRejectsInvalidBanks(t *testing.T) {
	valid := []FormationFrame{
		{Time: 0, Points: []Point{{1, 2}}},
		{Time: 1, Points: []Point{{1, 2}}},
	}
	for _, config := range []KeyframedFormationConfig{
		{Count: 1, Loop: 1},
		{Count: 1, Loop: math.NaN(), Frames: valid},
		{Count: 1, Loop: 1, Frames: []FormationFrame{{Time: .1, Points: []Point{{1, 2}}}, valid[1]}},
		{Count: 1, Loop: 1, Frames: []FormationFrame{valid[0], {Time: 0, Points: []Point{{1, 2}}}}},
		{Count: 1, Loop: 1, Frames: []FormationFrame{valid[0], {Time: 1, Points: []Point{{2, 2}}}}},
		{Count: 2, Loop: 1, Frames: valid},
		{Count: 1, Loop: 1, Frames: []FormationFrame{valid[0], {Time: 1, Points: []Point{{math.Inf(1), 2}}}}},
	} {
		if _, err := NewKeyframedFormation(config); err == nil {
			t.Fatalf("accepted invalid formation: %+v", config)
		}
	}
}
