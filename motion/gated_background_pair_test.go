package motion

import (
	"math"
	"testing"
)

func grodanBackgroundMotionConfig() GatedBackgroundPairConfig {
	return GatedBackgroundPairConfig{
		GateStep: .1, GateOpen: 10, GateReset: 20,
		FirstX: ThresholdAxis{Start: 0, Velocity: 1, Lower: -1280, Upper: 0,
			VelocityBelow: 16, VelocityAbove: -16},
		FirstY: ThresholdAxis{Start: 0, Velocity: 1, Lower: -400, Upper: 0,
			VelocityBelow: 1, VelocityAbove: -1},
		SecondY: ThresholdAxis{Start: 0, Velocity: 1, Lower: -400, Upper: 0,
			VelocityBelow: 2, VelocityAbove: -2},
		SecondXMin: -710, SecondXMax: 0,
		SecondXBelowVelocity: 16, SecondXAboveVelocity: -16,
	}
}

func TestGatedBackgroundPairMatchesGrodanBoundaryAndGateOrder(t *testing.T) {
	pair, err := NewGatedBackgroundPair(grodanBackgroundMotionConfig())
	if err != nil {
		t.Fatal(err)
	}
	moveX, moveY, vx, vy, gate := 0.0, 0.0, 1.0, 1.0, 0.0
	x, y, dx, dy := 0.0, 0.0, 0.0, 1.0
	gateCycles, directionChanges := 0, 0
	for tick := 1; tick <= 50000; tick++ {
		gate += .1
		if moveY < -400 {
			vy = 1
		}
		if moveY > 0 {
			vy = -1
		}
		moveY += vy
		if gate > 10 {
			if moveX < -1280 {
				vx = 16
			}
			if moveX > 0 {
				vx = -16
			}
			moveX += vx
		}
		if gate > 20 {
			gate = 0
			gateCycles++
		}
		if y < -400 {
			dy, dx = 2, 16
			directionChanges++
		}
		if y > 0 {
			dy, dx = -2, -16
			directionChanges++
		}
		x += dx
		if x < -710 {
			x = -710
		}
		if x > 0 {
			x = 0
		}
		y += dy
		if err := pair.Step(); err != nil {
			t.Fatal(err)
		}
		poses := pair.Poses()
		for i, got := range [...]float64{poses[0].X, poses[0].Y, poses[1].X, poses[1].Y, pair.GatePhase()} {
			want := [...]float64{moveX, moveY, x, y, gate}[i]
			if math.Abs(got-want) > 1e-12 {
				t.Fatalf("tick %d coordinate %d: %.12f != %.12f", tick, i, got, want)
			}
		}
	}
	if gateCycles < 10 || directionChanges < 10 {
		t.Fatal("test did not cover repeated gates and bounces")
	}
	if got := testing.AllocsPerRun(100, func() { _ = pair.Step(); pair.Poses() }); got != 0 {
		t.Fatalf("background pair allocates %.2f objects", got)
	}
}
