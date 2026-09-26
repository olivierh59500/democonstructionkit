package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestBilizirPlainAndWarpedLogoSharePhaseWithoutPositionJump(t *testing.T) {
	clock, err := motion.NewWaveClock(BilizirLogoClock())
	if err != nil {
		t.Fatal(err)
	}
	plain, err := motion.NewHarmonicTransform(BilizirPlainLogoMotion(800, 480))
	if err != nil {
		t.Fatal(err)
	}
	warped, err := motion.NewHarmonicTransform(BilizirWarpedLogoMotion(800, 480, 64))
	if err != nil {
		t.Fatal(err)
	}
	phase := 0.0
	margin := 64.0
	for tick := 0; tick < 5000; tick++ {
		wantPlain := 160 + math.Sin(phase)*160
		wantWarped := 160 + math.Sin(phase)*(160-margin)
		if got := plain.At(clock.Phase()).X; math.Abs(got-wantPlain) > 1e-10 {
			t.Fatalf("tick %d plain logo X = %g, want %g", tick, got, wantPlain)
		}
		if got := warped.At(clock.Phase()).X; math.Abs(got-wantWarped) > 1e-10 {
			t.Fatalf("tick %d warped logo X = %g, want %g", tick, got, wantWarped)
		}
		if tick == 2400 {
			margin = 75
			warped, err = motion.NewHarmonicTransform(BilizirWarpedLogoMotion(800, 480, margin))
			if err != nil {
				t.Fatal(err)
			}
			wantWarped = 160 + math.Sin(phase)*85
			if got := warped.At(clock.Phase()).X; math.Abs(got-wantWarped) > 1e-10 {
				t.Fatalf("changed margin reset phase: got %g, want %g", got, wantWarped)
			}
		}
		speed := 1.0
		if tick >= 1500 && tick < 3000 {
			speed = 1.5
		} else if tick >= 3000 {
			speed = .5
		}
		phase += .05 * speed
		if err := clock.SetStep(BilizirLogoPhaseStep() * speed); err != nil {
			t.Fatal(err)
		}
		clock.Step()
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = plain.At(clock.Phase())
		_ = warped.At(clock.Phase())
	}); allocations != 0 {
		t.Fatalf("logo pose sampling allocated %v times", allocations)
	}
}
