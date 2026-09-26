package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestBilizirLoopPreservesFirstPassAndBridgesLaterCopies(t *testing.T) {
	const width = 16000.0
	config, err := BilizirScrollLoop(width)
	if err != nil {
		t.Fatal(err)
	}
	clock, err := motion.NewWrapBank(config)
	if err != nil {
		t.Fatal(err)
	}
	position, firstWrap := 0.0, -1
	for tick := 0; tick < 12000; tick++ {
		if got := clock.At(0); got != position {
			t.Fatalf("tick %d X = %g, want %g", tick, got, position)
		}
		speed := 1.0
		if tick >= 5000 && tick < 7000 {
			speed = 1.5
		} else if tick >= 7000 {
			speed = .5
		}
		velocity := -4 * speed
		if err := clock.SetVelocity(0, velocity); err != nil {
			t.Fatal(err)
		}
		before := position
		position += velocity
		if position < -width {
			position += width
			if firstWrap < 0 {
				firstWrap = tick
			}
			if position < -8 || position > 0 {
				t.Fatalf("wrapped head jumped away from the departing tail: %g", position)
			}
			if math.Abs((before+velocity+width)-position) > 1e-12 {
				t.Fatal("relative wrap lost transport overshoot")
			}
		}
		clock.Step()
	}
	if firstWrap < 4000 {
		t.Fatalf("first cycle ended too soon at tick %d", firstWrap)
	}
	if _, err := BilizirScrollLoop(0); err == nil {
		t.Fatal("accepted zero-width message")
	}
	if allocations := testing.AllocsPerRun(100, clock.Step); allocations != 0 {
		t.Fatalf("relative scroll transport allocated %v times", allocations)
	}
}
