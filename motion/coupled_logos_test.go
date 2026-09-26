package motion

import (
	"math"
	"testing"
)

func TestCoupledLogoMotionMatchesReplicantsPhasesAndDepthOrder(t *testing.T) {
	stepA, stepB := 2.0/180.0*math.Pi, 7.0/180.0*math.Pi
	motion, err := NewCoupledLogoMotion(CoupledLogoConfig{
		StepPrimary: stepA, StepSecondary: stepB, SpeedMultiplier: 1.4,
		DepthBase: 1, DepthSin: 1, DepthDivisor: 2,
		YBase: 140, YSumAmplitude: 40,
		SecondaryDepthCos: 1.0 / 8.0, SecondaryYSin: 70,
	})
	if err != nil {
		t.Fatal(err)
	}
	phaseA, phaseB, speed := 0.0, 0.0, 1.4
	var orderSeen [2]bool
	for tick := 1; tick <= 5000; tick++ {
		if tick == 1000 {
			speed = .8
		}
		if tick == 3000 {
			speed = 2
		}
		if err := motion.SetSpeedMultiplier(speed); err != nil {
			t.Fatal(err)
		}
		phaseA += stepA * speed
		phaseB += stepB * speed
		if err := motion.Step(); err != nil {
			t.Fatal(err)
		}
		depthA := (1 + math.Sin(phaseA)) / 2
		yA := 140 + (math.Cos(phaseA)+math.Sin(phaseA))*40
		depthB := depthA + math.Cos(phaseB)/8
		yB := yA + math.Sin(phaseB)*(70*depthA)
		poses := motion.Poses()
		if math.Abs(poses[0].Depth-depthA) > 1e-12 || math.Abs(poses[0].Y-yA) > 1e-12 ||
			math.Abs(poses[1].Depth-depthB) > 1e-12 || math.Abs(poses[1].Y-yB) > 1e-12 {
			t.Fatalf("tick %d: poses %+v, want (%v,%v) (%v,%v)", tick, poses, yA, depthA, yB, depthB)
		}
		order := [2]int{1, 0}
		if depthB >= depthA {
			order = [2]int{0, 1}
		}
		if got := motion.DrawOrder(); got != order {
			t.Fatalf("tick %d: order %v, want %v", tick, got, order)
		}
		orderSeen[order[0]] = true
	}
	if !orderSeen[0] || !orderSeen[1] {
		t.Fatal("test did not exercise both logo depth orders")
	}
	if got := testing.AllocsPerRun(100, func() { _ = motion.Step(); motion.Poses() }); got != 0 {
		t.Fatalf("coupled logo motion allocates %.2f objects", got)
	}
}

func BenchmarkCoupledLogoMotion(b *testing.B) {
	motion, err := NewCoupledLogoMotion(CoupledLogoConfig{StepPrimary: .03,
		StepSecondary: .1, SpeedMultiplier: 1, DepthBase: 1, DepthDivisor: 2,
		YSumAmplitude: 40})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := motion.Step(); err != nil {
			b.Fatal(err)
		}
		motion.Poses()
	}
}
