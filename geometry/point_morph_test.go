package geometry

import (
	"math"
	"testing"
)

type morphPoints []Vec3

func (p morphPoints) Len() int                 { return len(p) }
func (p morphPoints) XYZ(index int) Vec3       { return p[index] }
func (p morphPoints) SetXYZ(index int, v Vec3) { p[index] = v }

type morphBuffer struct{ points morphPoints }

func (b *morphBuffer) Len() int                 { return len(b.points) }
func (b *morphBuffer) XYZ(index int) Vec3       { return b.points[index] }
func (b *morphBuffer) SetXYZ(index int, v Vec3) { b.points[index] = v }

func TestPointMorphKeepsIncrementalSourceAndSmoothHandoff(t *testing.T) {
	points := morphPoints{{X: 1.3, Y: -2.1, Z: 4}, {X: 5.2, Y: 6.1, Z: -7}}
	target := morphPoints{{X: 10.2, Y: 3.7, Z: 0}, {X: -8.4, Y: 9.3, Z: 11}}
	morph, err := NewPointMorph(points, target, PointMorphConfig{Frames: 7})
	if err != nil {
		t.Fatal(err)
	}
	want := append(morphPoints(nil), points...)
	for frame := 1; frame <= 7; frame++ {
		morph.Step(points)
		for i := range want {
			want[i].X += (target[i].X - []float64{1.3, 5.2}[i]) / 7
			want[i].Y += (target[i].Y - []float64{-2.1, 6.1}[i]) / 7
			want[i].Z += (target[i].Z - []float64{4, -7}[i]) / 7
			if points[i] != want[i] {
				t.Fatalf("frame %d point %d = %+v, want %+v", frame, i, points[i], want[i])
			}
		}
	}
	if !morph.Finished() || morph.Frame() != 7 {
		t.Fatal("morph did not finish at the configured frame")
	}
	morph.Step(points)
	if points[0] != want[0] || points[1] != want[1] {
		t.Fatal("finished morph advanced again")
	}
	before := append(morphPoints(nil), points...)
	second, err := NewPointMorph(points, morphPoints{{X: -10}, {Y: -20}}, PointMorphConfig{Frames: 5})
	if err != nil {
		t.Fatal(err)
	}
	if points[0] != before[0] || points[1] != before[1] {
		t.Fatal("constructing the next morph teleported the current shape")
	}
	second.Step(points)
	if points[0].X != before[0].X+(-10-before[0].X)/5 {
		t.Fatal("new morph did not continue from the previous pose")
	}
}

func TestPointMorphTailAndExactFinalOptions(t *testing.T) {
	from := morphPoints{{X: 10}, {X: 20}, {X: 30}}
	target := morphPoints{{}}
	for _, test := range []struct {
		name string
		tail MorphTail
		want morphPoints
	}{
		{"hold", MorphHold, morphPoints{{}, {X: 20}, {X: 30}}},
		{"repeat", MorphRepeat, morphPoints{{}, {}, {}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			points := append(morphPoints(nil), from...)
			morph, err := NewPointMorph(points, target, PointMorphConfig{Frames: 3, Tail: test.tail, SnapFinal: true})
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				morph.Step(points)
			}
			for i := range points {
				if points[i] != test.want[i] {
					t.Fatalf("point %d = %+v, want %+v", i, points[i], test.want[i])
				}
			}
		})
	}
	long, err := NewPointMorph(from, target, PointMorphConfig{Frames: 100_000})
	if err != nil {
		t.Fatal(err)
	}
	points := append(morphPoints(nil), from...)
	buffer := &morphBuffer{points: points}
	if allocations := testing.AllocsPerRun(100, func() { long.Step(buffer) }); allocations != 0 {
		t.Fatalf("point morph step allocates %v times", allocations)
	}
}

func TestPointMorphRejectsInvalidInput(t *testing.T) {
	points := morphPoints{{X: 1}}
	for _, config := range []PointMorphConfig{{}, {Frames: -1}, {Frames: 1, Tail: 2}} {
		if _, err := NewPointMorph(points, points, config); err == nil {
			t.Fatalf("accepted config %+v", config)
		}
	}
	if _, err := NewPointMorph(morphPoints{{X: math.NaN()}}, points, PointMorphConfig{Frames: 1}); err == nil {
		t.Fatal("accepted a nonfinite source")
	}
}
