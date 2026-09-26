package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCocoTitleMotionMatchesAuthoredPhaseAndSpeed(t *testing.T) {
	clock, err := motion.NewWaveClock(CocoTitleMotion(800))
	if err != nil {
		t.Fatal(err)
	}
	phase := .5
	if clock.At(0) != 64+800*math.Cos(phase) {
		t.Fatal("initial title pose changed")
	}
	for tick := 0; tick < 1200; tick++ {
		speed := 1.0
		if tick >= 400 && tick < 800 {
			speed = 1.5
		}
		if err := clock.SetStep(.0125 * speed); err != nil {
			t.Fatal(err)
		}
		clock.Step()
		phase += .0125 * speed
		if got, want := clock.At(0), 64+800*math.Cos(phase); got != want {
			t.Fatalf("tick %d title x = %v, want %v", tick, got, want)
		}
	}
}
