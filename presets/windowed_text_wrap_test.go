package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestWindowedTextWrapMatchesStrictOriginalClockWithLiveSpeedChanges(t *testing.T) {
	for _, sample := range []struct {
		name  string
		width float64
		speed float64
	}{
		{name: "TNT Crew 2", width: 17792, speed: 4},
		{name: "Union Replicants", width: 15168, speed: 6},
	} {
		t.Run(sample.name, func(t *testing.T) {
			config := WindowedTextWrap(sample.width, 640, sample.speed)
			clock, err := motion.NewWrapBank(config)
			if err != nil {
				t.Fatal(err)
			}
			oldX, speed := -640.0, sample.speed
			for frame := 0; frame < 10000; frame++ {
				if frame%413 == 0 {
					speed = float64((frame / 413) % 10)
					if err := clock.SetVelocity(0, -speed); err != nil {
						t.Fatal(err)
					}
				}
				if got := clock.At(0); got != oldX {
					t.Fatalf("frame %d X = %v, want %v", frame, got, oldX)
				}
				oldX -= speed
				if oldX < -(sample.width - 640) {
					oldX = -640
				}
				clock.Step()
			}
			if allocations := testing.AllocsPerRun(100, clock.Step); allocations != 0 {
				t.Fatalf("scroll clock allocated %v times per frame", allocations)
			}
		})
	}
}
