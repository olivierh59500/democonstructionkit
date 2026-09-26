package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestBilizirCubeMotionKeepsTwelveIndependentPhasesAndRotations(t *testing.T) {
	config := BilizirCubeMotion(800, 12)
	if config.Count != 12 {
		t.Fatalf("cube count = %d, want 12", config.Count)
	}
	var sourcePhases, sharedPhases [12]float64
	var sourceRotations, sharedRotations [12]geometry.Vec3
	for i := 0; i < 12; i++ {
		sourcePhases[i] = .15 * float64(i+1)
		sharedPhases[i] = config.PhaseSpacing * (float64(i) + config.PhaseIndexOrigin)
		sourceRotations[i] = geometry.Vec3{X: float64(i) * .3, Y: float64(i) * .5, Z: float64(i) * .2}
		sharedRotations[i] = geometry.Vec3{X: float64(i) * config.RotationSpacing.X,
			Y: float64(i) * config.RotationSpacing.Y, Z: float64(i) * config.RotationSpacing.Z}
	}
	for tick := 0; tick < 5000; tick++ {
		for i := 0; i < 12; i++ {
			wantX := 380 + 380*math.Sin(sourcePhases[i])
			wantY := 186 + 84*math.Cos(sourcePhases[i]*2.5)
			gotX, gotY := config.X.At(0, sharedPhases[i]), config.Y.At(0, sharedPhases[i])
			if math.Abs(gotX-wantX) > 1e-10 || math.Abs(gotY-wantY) > 1e-10 {
				t.Fatalf("tick %d cube %d pose = (%g,%g), want (%g,%g)", tick, i, gotX, gotY, wantX, wantY)
			}
			got, want := sharedRotations[i], sourceRotations[i]
			if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 || math.Abs(got.Z-want.Z) > 1e-12 {
				t.Fatalf("tick %d cube %d rotation = %+v, want %+v", tick, i, got, want)
			}
		}
		speed := 1.0
		if tick >= 1500 && tick < 3000 {
			speed = 1.5
		} else if tick >= 3000 {
			speed = .5
		}
		for i := 0; i < 12; i++ {
			index := float64(i)
			sourcePhases[i] += .04 * speed
			sharedPhases[i] += config.PhaseStep * speed
			sourceRotations[i].X += .02 * speed * (1 + index*.1)
			sourceRotations[i].Y += .03 * speed * (1 + index*.15)
			sourceRotations[i].Z += .01 * speed * (1 + index*.05)
			sharedRotations[i].X += config.RotationStep.X * speed * (1 + index*config.RotationIndexFactor.X)
			sharedRotations[i].Y += config.RotationStep.Y * speed * (1 + index*config.RotationIndexFactor.Y)
			sharedRotations[i].Z += config.RotationStep.Z * speed * (1 + index*config.RotationIndexFactor.Z)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = config.X.At(0, sharedPhases[0])
		_ = config.Y.At(0, sharedPhases[0])
	}); allocations != 0 {
		t.Fatalf("cube pose sampling allocated %v times", allocations)
	}
}
