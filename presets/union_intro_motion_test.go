package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestUnionIntroBackdropAndLogoMatchOriginalTicks(t *testing.T) {
	background, err := motion.NewWrapBank(UnionIntroBackdropMotion())
	if err != nil {
		t.Fatal(err)
	}
	logo, err := motion.NewHarmonicTransform(UnionIntroLogoTransform())
	if err != nil {
		t.Fatal(err)
	}
	for frame := 0; frame < 5000; frame++ {
		wantBackdrop := float64((frame % 24) * 4)
		if got := background.At(0); got != wantBackdrop {
			t.Fatalf("frame %d backdrop Y = %v, want %v", frame, got, wantBackdrop)
		}
		pose := logo.At(float64(frame) * UnionIntroLogoPhaseStep())
		wantX := 256 + 190*math.Sin(float64(frame)*.05)
		if math.Abs(pose.X-wantX) > 1e-10 || pose.Y != 43 || pose.ScaleX != 1 || pose.ScaleY != 1 {
			t.Fatalf("frame %d logo pose = %+v, want X=%v", frame, pose, wantX)
		}
		background.Step()
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = logo.At(3.5)
		background.Step()
	}); allocations != 0 {
		t.Fatalf("intro motion allocated %v times per frame", allocations)
	}
}
