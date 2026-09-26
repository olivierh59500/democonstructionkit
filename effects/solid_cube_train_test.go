package effects_test

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/presets"
)

func TestCocoCubeTrainMatchesIndependentSourcePhases(t *testing.T) {
	train, err := effects.NewSolidCubeTrain(presets.CocoCubeTrain(800, 600, 40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer train.Close()
	var phases [12]float64
	var rotations [12][3]float64
	for i := range phases {
		phases[i] = .15 * float64(i+1)
	}
	for tick := 0; tick < 1200; tick++ {
		speed := 1.0
		if tick >= 400 && tick < 800 {
			speed = 1.5
		}
		if tick >= 800 && tick < 820 {
			speed = 0
		}
		if err := train.SetSpeed(speed); err != nil {
			t.Fatal(err)
		}
		if err := train.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		for i := range phases {
			phases[i] += .04 * speed
			index := float64(i)
			rotations[i][0] += .02 * speed * (1 + index*.1)
			rotations[i][1] += .03 * speed * (1 + index*.15)
			rotations[i][2] += .01 * speed * (1 + index*.05)
			position, rotation, ok := train.Pose(i)
			if !ok {
				t.Fatalf("tick %d cube %d is missing", tick, i)
			}
			wantX := 380 + 380*math.Sin(phases[i])
			wantY := 300 + 84*math.Cos(phases[i]*2.5)
			if math.Abs(position.X-wantX) > 1e-10 || math.Abs(position.Y-wantY) > 1e-10 ||
				math.Abs(rotation.X-(index*.3+rotations[i][0])) > 1e-10 ||
				math.Abs(rotation.Y-(index*.2+rotations[i][1])) > 1e-10 ||
				math.Abs(rotation.Z-(index*.1+rotations[i][2])) > 1e-10 {
				t.Fatalf("tick %d cube %d pose (%+v, %+v) differs", tick, i, position, rotation)
			}
		}
	}
}

func TestSolidCubeTrainAcceptsCustomPathAndMaterials(t *testing.T) {
	material := effects.DefaultSolidCubeConfig(20)
	config := effects.SolidCubeTrainConfig{
		Count: 2, CubeConfigs: []effects.SolidCubeConfig{material, material},
		PhaseStart: .25, PhaseSpacing: .5, PhaseStep: .1,
		Path: func(index int, phase float64) motion.Point {
			return motion.Point{X: float64(index) * 10, Y: phase}
		},
	}
	train, err := effects.NewSolidCubeTrain(config)
	if err != nil {
		t.Fatal(err)
	}
	defer train.Close()
	if err := train.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	position, _, ok := train.Pose(1)
	if !ok || position.X != 10 || math.Abs(position.Y-.85) > 1e-12 {
		t.Fatalf("custom path pose %+v, visible %v", position, ok)
	}
	config.CubeConfigs = config.CubeConfigs[:1]
	if _, err := effects.NewSolidCubeTrain(config); err == nil {
		t.Fatal("accepted fewer cube materials than instances")
	}
}

func TestMultiscreenCubeTrainKeepsReanchoredRecurrence(t *testing.T) {
	config := presets.MultiscreenCocoCubeTrain(800, 600, 40, 12)
	train, err := effects.NewSolidCubeTrain(config)
	if err != nil {
		t.Fatal(err)
	}
	defer train.Close()
	phase := .15
	pathSin, pathCos := math.Sincos(phase)
	bobSin, bobCos := math.Sincos(phase * 2.5)
	pathStepSin, pathStepCos := math.Sincos(.04)
	bobStepSin, bobStepCos := math.Sincos(.1)
	for tick := 1; tick <= 10_000; tick++ {
		if err := train.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		phase += .04
		if tick&1023 == 0 {
			phase = math.Mod(phase, 4*math.Pi)
			pathSin, pathCos = math.Sincos(phase)
			bobSin, bobCos = math.Sincos(phase * 2.5)
		} else {
			pathSin, pathCos = pathSin*pathStepCos+pathCos*pathStepSin, pathCos*pathStepCos-pathSin*pathStepSin
			bobSin, bobCos = bobSin*bobStepCos+bobCos*bobStepSin, bobCos*bobStepCos-bobSin*bobStepSin
		}
		position, _, ok := train.Pose(0)
		if !ok || math.Abs(position.X-(380+380*pathSin)) > 1e-11 ||
			math.Abs(position.Y-(300+84*bobCos)) > 1e-11 {
			t.Fatalf("tick %d cached cube position %+v differs", tick, position)
		}
	}
	config.Path = func(int, float64) motion.Point { return motion.Point{} }
	if _, err := effects.NewSolidCubeTrain(config); err == nil {
		t.Fatal("accepted a custom path with harmonic recurrence")
	}
}
