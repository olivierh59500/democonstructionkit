package geometry

import (
	"math"
	"testing"
)

func TestRotateXYZScaledPreservesVectorballMatrixOrder(t *testing.T) {
	for frame := 0; frame < 5000; frame++ {
		angles := Vec3{X: float64(frame)*.017 - 2, Y: float64(frame)*-.023 + .4, Z: float64(frame)*.031 - 1}
		scale := .35 + float64(frame%7)*.125
		sinX, cosX := math.Sincos(angles.X)
		sinY, cosY := math.Sincos(angles.Y)
		sinZ, cosZ := math.Sincos(angles.Z)
		want := Rotation{
			cosY * cosZ * scale,
			(sinX*sinY*cosZ - cosX*sinZ) * scale,
			(cosX*sinY*cosZ + sinX*sinZ) * scale,
			cosY * sinZ * scale,
			(sinX*sinY*sinZ + cosX*cosZ) * scale,
			(cosX*sinY*sinZ - sinX*cosZ) * scale,
			-sinY * scale,
			sinX * cosY * scale,
			cosX * cosY * scale,
		}
		got := RotateXYZScaled(angles, scale)
		for coefficient := range got {
			if math.Abs(got[coefficient]-want[coefficient]) > 1e-12*math.Max(1, math.Abs(want[coefficient])) {
				t.Fatalf("frame %d coefficient %d = %.17g, want %.17g", frame, coefficient, got[coefficient], want[coefficient])
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = RotateXYZScaled(Vec3{X: .2, Y: .4, Z: .6}, .35)
	}); allocations != 0 {
		t.Fatalf("scaled rotation allocated %v times", allocations)
	}
}
