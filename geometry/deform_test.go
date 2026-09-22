package geometry

import (
	"encoding/json"
	"math"
	"testing"
)

func TestDeformProgramOrderReferenceAndInPlace(t *testing.T) {
	stages := []DeformStage{
		{Kind: DeformScale, Vector: Vec3{2, 3, 4}},
		{Kind: DeformReference},
		{Kind: DeformWobble, Wobble: WobbleField{X: []Harmonic{{Gain: 1, Speed: 1}}, Amount: 2, Radius: 1, BaseInfluence: 1, RadialInfluence: 1}},
		{Kind: DeformTranslate, Vector: Vec3{X: 3}},
	}
	p, err := NewDeformProgram(stages)
	if err != nil {
		t.Fatal(err)
	}
	source := []Vec3{{1, 0, 0}, {0, 1, 0}}
	expected := []Vec3{{11, 0, 0}, {11, 3, 0}}
	got := append([]Vec3(nil), source...)
	if n := p.Apply(got, got, math.Pi/2); n != len(got) {
		t.Fatal(n)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("point %d: %v != %v", i, got[i], expected[i])
		}
	}
	// The program owns its harmonic arrays, including when a caller edits a config.
	stages[2].Wobble.X[0].Gain = 200
	p.Apply(got, source, math.Pi/2)
	if got[0] != expected[0] {
		t.Fatal("configuration alias")
	}
	var encoded []byte
	encoded, err = json.Marshal(stages)
	if err != nil {
		t.Fatal(err)
	}
	var restored []DeformStage
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if _, err = NewDeformProgram(restored); err != nil {
		t.Fatal(err)
	}
}

func TestDeformProgramTwistAndRipple(t *testing.T) {
	p, err := NewDeformProgram([]DeformStage{
		{Kind: DeformRipple, Ripple: RippleField{Amplitude: 2, Speed: 1, Radius: 1, CoordinateScale: 1, DirectionX: Vec3{Y: 1}}},
		{Kind: DeformTwist, Twist: TwistField{Axis: 1, Amount: math.Pi / 2, Radius: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	source := []Vec3{{0, 1, 0}}
	dst := make([]Vec3, 1)
	p.Apply(dst, source, math.Pi/2)
	if math.Abs(dst[0].X) > 1e-14 || dst[0].Y != 1 || dst[0].Z != 2 {
		t.Fatal(dst)
	}
	if err = p.SetTwistAmount(1, 0); err != nil {
		t.Fatal(err)
	}
	p.Apply(dst, source, math.Pi/2)
	if dst[0] != (Vec3{2, 1, 0}) {
		t.Fatal(dst)
	}
}

func TestDeformValidationAndBoundedWrites(t *testing.T) {
	bad := []DeformStage{{Kind: "unknown"}, {Kind: DeformTwist, Twist: TwistField{Axis: 3, Radius: 1}}, {Kind: DeformRipple}, {Kind: DeformRotate, Vector: Vec3{X: math.NaN()}}, {Kind: DeformWobble, Wobble: WobbleField{X: []Harmonic{{Speed: math.Inf(1)}}}}}
	for _, s := range bad {
		if _, err := NewDeformProgram([]DeformStage{s}); err == nil {
			t.Fatalf("accepted %v", s)
		}
	}
	p, err := NewDeformProgram([]DeformStage{{Kind: DeformTranslate, Vector: Vec3{X: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	dst := []Vec3{{9, 9, 9}, {8, 8, 8}}
	if n := p.Apply(dst, []Vec3{{1, 2, 3}}, 0); n != 1 || dst[1] != (Vec3{8, 8, 8}) {
		t.Fatal(dst)
	}
	if n := p.Apply(dst, nil, math.NaN()); n != 0 {
		t.Fatal(n)
	}
	if p.SetVector(1, Vec3{}) == nil || p.SetWobbleAmount(0, 1) == nil || p.SetTwistAmount(0, 1) == nil {
		t.Fatal("invalid update accepted")
	}
	if got := testing.AllocsPerRun(100, func() { p.Apply(dst, dst, 0) }); got != 0 {
		t.Fatalf("allocations: %v", got)
	}
}
