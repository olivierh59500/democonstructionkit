package geometry

import (
	"math"
	"testing"
)

func TestOrderedEulerMatchesUnionFaceRotation(t *testing.T) {
	transform, err := NewOrderedEuler([3]uint8{2, 1, 0}, Vec3{X: 1, Y: -1, Z: -1})
	if err != nil {
		t.Fatal(err)
	}
	points := [...]Vec3{{X: 0.3025, Y: -139.5405, Z: 99},
		{X: -268.2185, Y: 344.7565, Z: -99}, {X: 114, Y: 258, Z: -60},
		{X: -10, Y: -212, Z: 0}, {X: 0, Y: 133, Z: 0}}
	angles := Vec3{X: -math.Pi / 2, Z: math.Pi / 3}
	step := Vec3{X: .033, Y: .032, Z: .031}
	for tick := 0; tick < 1000; tick++ {
		rx := RotateXYZ(Vec3{X: angles.X})
		ry := RotateXYZ(Vec3{Y: angles.Y})
		rz := RotateXYZ(Vec3{Z: angles.Z})
		transform.SetAngles(angles)
		for index, original := range points {
			want := rx.Apply(ry.Apply(rz.Apply(original)))
			want = Vec3{X: want.X, Y: -want.Y, Z: -want.Z}
			if got := transform.Apply(original); got != want {
				t.Fatalf("tick %d point %d = %+v, want %+v", tick, index, got, want)
			}
		}
		angles = angles.Add(step)
	}
	if got := testing.AllocsPerRun(100, func() {
		transform.SetAngles(angles)
		for _, p := range points {
			transform.Apply(p)
		}
	}); got != 0 {
		t.Fatalf("ordered Euler transform allocated %.2f objects", got)
	}
}
