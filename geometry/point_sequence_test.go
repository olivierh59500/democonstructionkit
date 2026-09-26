package geometry

import (
	"fmt"
	"testing"
)

type sequencePoints struct{ values []Vec3 }

func (p *sequencePoints) Len() int                 { return len(p.values) }
func (p *sequencePoints) XYZ(index int) Vec3       { return p.values[index] }
func (p *sequencePoints) SetXYZ(index int, v Vec3) { p.values[index] = v }

type sequenceShapes map[string][]Vec3

func (b sequenceShapes) Clone(name string) (PointWriter, error) {
	points, ok := b[name]
	if !ok {
		return nil, fmt.Errorf("unknown shape %q", name)
	}
	return &sequencePoints{values: append([]Vec3(nil), points...)}, nil
}

func (b sequenceShapes) Target(name string) (PointReader, error) {
	points, ok := b[name]
	if !ok {
		return nil, fmt.Errorf("unknown target %q", name)
	}
	return &sequencePoints{values: points}, nil
}

func TestPointSequenceInheritsAndClearsActionFieldsAtExactBoundaries(t *testing.T) {
	position := Vec3{X: 1, Y: 2, Z: 3}
	zero := Vec3{}
	rotationStep := Vec3{X: 1}
	translationStep := Vec3{X: 1, Y: 2}
	textZero, textOne := 0, 1
	sequence, err := NewPointSequence(PointSequenceConfig{
		Shapes: sequenceShapes{"a": {{X: 4}}}, DurationScale: 1, Loop: true,
		Actions: []PointAction{
			{Frames: 3, Shape: "a", SetShape: true, Position: &position,
				RotationStep: &rotationStep, TranslationStep: &translationStep,
				TextIndex: &textZero, SetAnimations: true,
				Animations: []PointAnimationSpec{{Kind: PointAnimBounce, Bounce: BounceCurveConfig{Samples: []float64{10, 20}}}}},
			{Frames: 2, Position: &zero, TextIndex: &textOne},
			{Frames: 2, RotationStep: &zero, TranslationStep: &zero, SetAnimations: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sequence.ActionIndex() != 0 || sequence.Remaining() != 3 || sequence.AnimationCount() != 1 ||
		sequence.State().Position != position {
		t.Fatalf("first stage = %+v, action %d, remaining %d", sequence.State(), sequence.ActionIndex(), sequence.Remaining())
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State(); got.Position != (Vec3{X: 2, Y: 10, Z: 3}) || got.Rotation.X != 1 {
		t.Fatalf("first moving tick = %+v", got)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State(); got.Position.Y != 20 || got.Rotation.X != 2 {
		t.Fatalf("second tick = %+v", got)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State(); sequence.ActionIndex() != 1 || sequence.Remaining() != 2 || got.Position != zero ||
		got.Rotation.X != 2 || got.TextIndex != 1 || sequence.AnimationCount() != 1 {
		t.Fatalf("inheritance boundary = %+v, action %d", got, sequence.ActionIndex())
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State(); got.Position.X != 1 || got.Position.Y != 10 || got.Rotation.X != 3 {
		t.Fatalf("inherited motion = %+v", got)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if sequence.ActionIndex() != 2 || sequence.AnimationCount() != 0 || sequence.State().Position.Y != 10 {
		t.Fatal("explicit empty animation list did not clear at the boundary")
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State(); got.Position.X != 1 || got.Position.Y != 10 || got.Rotation.X != 3 {
		t.Fatalf("explicit zero motion was not retained: %+v", got)
	}
}

func TestPointSequenceMorphContinuesFromCurrentPoseWithoutBoundaryStep(t *testing.T) {
	sequence, err := NewPointSequence(PointSequenceConfig{
		Shapes:        sequenceShapes{"a": {{X: 1, Y: 2}, {X: 5}}, "b": {{X: 11, Y: 7}, {X: 20}}},
		DurationScale: 1, Loop: true,
		Actions: []PointAction{
			{Frames: 2, Shape: "a", SetShape: true, SetAnimations: true},
			{Frames: 5, SetAnimations: true, Animations: []PointAnimationSpec{{Kind: PointAnimMorph, Target: "b"}}},
			{Frames: 2, SetAnimations: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	first := sequence.State().Points.XYZ(0)
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if sequence.ActionIndex() != 1 || sequence.State().Points.XYZ(0) != first {
		t.Fatal("entering the morph moved the first point on its boundary tick")
	}
	for i := 0; i < 4; i++ {
		if err := sequence.Step(); err != nil {
			t.Fatal(err)
		}
	}
	before := sequence.State().Points.XYZ(0)
	if before.X != 1+4*float64(11-1)/5 || before.Y != 2+4*float64(7-2)/5 {
		t.Fatalf("morph pose before handoff = %+v", before)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if sequence.ActionIndex() != 2 || sequence.State().Points.XYZ(0) != before || sequence.AnimationCount() != 0 {
		t.Fatal("clearing the morph teleported its last displayed pose")
	}
}

func TestPointSequenceOrdersAppendAndDropPositionalOverrides(t *testing.T) {
	sequence, err := NewPointSequence(PointSequenceConfig{
		Shapes: sequenceShapes{"a": {{}}}, DurationScale: 1, Loop: true,
		Actions: []PointAction{
			{Frames: 2, Shape: "a", SetShape: true, SetAnimations: true,
				Animations: []PointAnimationSpec{{Kind: PointAnimBounce, Bounce: BounceCurveConfig{Samples: []float64{7}}}}},
			{Frames: 2, AppendAnimations: []PointAnimationSpec{{Kind: PointAnimYOrbit,
				YOrbit: YOrbitConfig{Center: Vec3{Z: 850}, Radius: 100, RatioStep: 1, Min: 0, Max: 1}}}},
			{Frames: 2, DropLast: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := sequence.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if sequence.AnimationCount() != 2 {
		t.Fatal("append cue did not retain the bounce")
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State().Position; got != (Vec3{X: 100, Z: 850}) {
		t.Fatalf("late full-position override = %+v", got)
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if sequence.AnimationCount() != 1 {
		t.Fatal("drop cue did not remove the last animation")
	}
	if err := sequence.Step(); err != nil {
		t.Fatal(err)
	}
	if got := sequence.State().Position.Y; got != 7 {
		t.Fatalf("remaining bounce Y = %v", got)
	}
}

func TestPointSequenceOrdinaryStepsReuseBuffers(t *testing.T) {
	sequence, err := NewPointSequence(PointSequenceConfig{
		Shapes: sequenceShapes{"a": {{}}}, DurationScale: 1,
		Actions: []PointAction{{Frames: 1000, Shape: "a", SetShape: true,
			SetAnimations: true, Animations: []PointAnimationSpec{{Kind: PointAnimBounce,
				Bounce: BounceCurveConfig{Samples: []float64{1, 2}}}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		if err := sequence.Step(); err != nil {
			t.Fatal(err)
		}
	}); allocations != 0 {
		t.Fatalf("point sequence allocated %v objects on an ordinary tick", allocations)
	}
}
