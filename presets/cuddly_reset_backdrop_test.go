package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCuddlyResetBackdropMatchesShowcaseFrameByFrame(t *testing.T) {
	trajectory, err := motion.NewFormulaTrajectory(CuddlyResetBackdropTrajectory())
	if err != nil {
		t.Fatal(err)
	}
	index, offset := 12.0, 0.0
	for frame := 0; frame < 5000; frame++ {
		angle := 2 * math.Pi / 384 * index
		want := motion.Point{
			X: 335 - 220*math.Sin(angle)*math.Sin(offset),
			Y: 250 + 170*math.Sin(angle+math.Pi/3),
		}
		got := trajectory.At(0, 0, 0, 0)
		if math.Abs(got.X-want.X) > 1e-10 || math.Abs(got.Y-want.Y) > 1e-10 {
			t.Fatalf("frame %d pose = %+v, want %+v", frame, got, want)
		}
		index++
		offset = math.Mod(offset+math.Pi/32, 2*math.Pi)
		trajectory.Step()
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = trajectory.At(0, 0, 0, 0)
		trajectory.Step()
	}); allocations != 0 {
		t.Fatalf("trajectory allocated %v times per frame", allocations)
	}
}
