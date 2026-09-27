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

func TestCuedFormationComposesKeyframedOriginSpacingAndHarmonics(t *testing.T) {
	origin := []Key[Point]{{Time: 0, Value: Point{10, 20}}, {Time: 1, Value: Point{50, 30}}, {Time: 2, Value: Point{10, 20}}}
	spacing := []Key[Point]{{Time: 0, Value: Point{30, 0}}, {Time: 1, Value: Point{5, 0}}, {Time: 2, Value: Point{30, 0}}}
	formation, err := NewCuedFormation(CuedFormationConfig{
		Count: 3, Loop: 4, OriginKeys: origin, SpacingKeys: spacing,
		Cues: []FormationCue{{Start: .25, Duration: .5, LeadIndex: 0,
			X: []FormationHarmonic{{FirstAmplitude: 20, LastAmplitude: 20, Cycles: .5}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		seconds float64
		index   int
		want    Point
	}{
		{0, 2, Point{70, 20}},
		{.5, 0, Point{50, 25}},
		{.5, 2, Point{85, 25}},
		{1, 2, Point{60, 30}},
		{2, 2, Point{70, 20}},
		{4, 2, Point{70, 20}},
	} {
		if got := formation.At(sample.seconds, sample.index); got != sample.want {
			t.Fatalf("pose at %.1fs for item %d = %+v, want %+v", sample.seconds, sample.index, got, sample.want)
		}
	}
	origin[1].Value.X, spacing[1].Value.X = 999, 999
	if got := formation.At(1, 2); got != (Point{60, 30}) {
		t.Fatalf("caller changed compiled keyframes: %+v", got)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = formation.At(1.5, 1) }); allocations != 0 {
		t.Fatalf("keyframed formation allocated %g times per pose", allocations)
	}
}

func TestCuedFormationRejectsInvalidKeyframedTracks(t *testing.T) {
	for _, keys := range [][]Key[Point]{
		{{Time: -1, Value: Point{1, 0}}},
		{{Time: 0, Value: Point{math.NaN(), 0}}},
		{{Time: 0, Value: Point{1, 0}}, {Time: 0, Value: Point{1, 0}}},
		{{Time: 0, Value: Point{1, 0}}, {Time: 5, Value: Point{1, 0}}},
		{{Time: 0, Value: Point{1, 0}}, {Time: 2, Value: Point{2, 0}}},
	} {
		if _, err := NewCuedFormation(CuedFormationConfig{Count: 2, Loop: 4, OriginKeys: keys}); err == nil {
			t.Fatalf("accepted invalid formation keys %+v", keys)
		}
	}
}

func TestCuedFormationPoseKeysComposeAndCopy(t *testing.T) {
	keys := []FormationPoseKey{
		{Time: 0, Origin: Point{10, 20}, Spacing: Point{30, 0}},
		{Time: 1, Origin: Point{50, 30}, Spacing: Point{5, 0}, Arc: Point{Y: -10}},
		{Time: 2, Origin: Point{10, 20}, Spacing: Point{30, 0}},
	}
	formation, err := NewCuedFormation(CuedFormationConfig{
		Count: 3, Loop: 4, PoseKeys: keys,
		Cues: []FormationCue{{Start: .25, Duration: .5, LeadIndex: 0,
			X: []FormationHarmonic{{FirstAmplitude: 20, LastAmplitude: 20, Cycles: .5}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := formation.At(.5, 2); got != (Point{85, 25}) {
		t.Fatalf("unexpected combined pose and cue: %+v", got)
	}
	if got := formation.At(.5, 1); got != (Point{67.5, 20}) {
		t.Fatalf("arch did not compose with spacing and cue: %+v", got)
	}
	keys[1].Origin.X, keys[1].Spacing.X = 999, 999
	if got := formation.At(1, 2); got != (Point{60, 30}) {
		t.Fatalf("caller changed compiled pose: %+v", got)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = formation.At(1.5, 1) }); allocations != 0 {
		t.Fatalf("pose-key formation allocated %g times per pose", allocations)
	}
}

func TestCuedFormationPoseKeysRejectInvalidInput(t *testing.T) {
	for _, config := range []CuedFormationConfig{
		{Count: 2, PoseKeys: []FormationPoseKey{{Time: 0, Origin: Point{math.NaN(), 0}}}},
		{Count: 2, PoseKeys: []FormationPoseKey{{Time: 0, Arc: Point{0, math.Inf(1)}}}},
		{Count: 2, Loop: 2, PoseKeys: []FormationPoseKey{{Time: 0, Spacing: Point{1, 0}}, {Time: 2, Spacing: Point{2, 0}}}},
		{Count: 2, PoseKeys: []FormationPoseKey{{Time: 0}}, OriginKeys: []Key[Point]{{Time: 0}}},
	} {
		if _, err := NewCuedFormation(config); err == nil {
			t.Fatalf("accepted invalid pose-key configuration: %+v", config)
		}
	}
}
