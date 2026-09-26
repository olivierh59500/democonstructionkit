package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCuddlySpreadpointLogoMatchesBothOriginalCycles(t *testing.T) {
	phase, err := motion.NewPhaseSequence(CuddlySpreadpointLogoPhases())
	if err != nil {
		t.Fatal(err)
	}
	orbit, err := motion.NewHarmonicTransform(CuddlySpreadpointLogoTransform())
	if err != nil {
		t.Fatal(err)
	}
	var original []float64
	angle := math.Pi
	for _, segment := range []struct {
		count int
		step  float64
		reset *float64
	}{
		{300, 0, scalar(math.Pi)}, {40, -math.Pi / 40, nil},
		{300, 0, scalar(0)}, {640, -math.Pi / 40, nil},
		{480, -math.Pi / 32, nil}, {512, math.Pi / 32, nil},
	} {
		for i := 0; i < segment.count; i++ {
			if segment.reset != nil {
				angle = *segment.reset
			} else {
				angle += segment.step
			}
			original = append(original, math.Mod(angle, 2*math.Pi))
		}
	}
	if phase.Len() != len(original) || phase.Len() != 2272 {
		t.Fatalf("phase length = %d, want %d", phase.Len(), len(original))
	}
	for frame := 0; frame < 2*len(original); frame++ {
		oldAngle := original[frame%len(original)]
		if got := phase.Current(); got != oldAngle || phase.AtTick(frame) != oldAngle {
			t.Fatalf("frame %d angle = %g, want %g", frame, got, oldAngle)
		}
		y, z := math.Sin(oldAngle), math.Cos(oldAngle)/4+.75
		wantX, wantY := 64-64*z, 45+40*y
		pose := orbit.At(phase.Current())
		if math.Abs(pose.X-wantX) > 1e-10 || math.Abs(pose.Y-wantY) > 1e-10 ||
			math.Abs(pose.ScaleX-z) > 1e-10 || math.Abs(pose.ScaleY-z) > 1e-10 {
			t.Fatalf("frame %d pose = %+v, want x=%g y=%g scale=%g", frame, pose, wantX, wantY, z)
		}
		phase.Step()
	}
	if phase.Index() != 0 || phase.AtTick(-1) != original[len(original)-1] {
		t.Fatal("phase sequence did not wrap or sample negative ticks")
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = orbit.At(phase.Current())
		phase.Step()
	}); allocations != 0 {
		t.Fatalf("logo pose allocated %v times per frame", allocations)
	}
}

func scalar(value float64) *float64 { return &value }
