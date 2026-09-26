package presets

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestVivaRotozoomStageBoundaries(t *testing.T) {
	program, err := NewVivaRotozoom(DefaultVivaRotozoomConfig(768, 540))
	if err != nil {
		t.Fatal(err)
	}
	if err := program.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	p := program.Repetition()
	if p.CenterX != 380 || p.CenterY != 270 || p.Zoom != 1 || p.Rotation != 0 || p.PhaseX != 6144 || p.PhaseY != 4320 {
		t.Fatalf("first entry pose = %+v", p)
	}
	for tick := 2; tick <= 383; tick++ {
		program.Update(kit.Frame{})
	}
	if program.entryAngles != 0 || program.stage != 0 {
		t.Fatalf("entry rotated early at tick 383: %+v", program)
	}
	program.Update(kit.Frame{})
	if program.entryAngles != 1 || program.stage != 0 {
		t.Fatalf("first rotation boundary = %+v", program)
	}
	for tick := 385; tick <= 428; tick++ {
		program.Update(kit.Frame{})
	}
	if program.stage != 1 || program.entryAngles != 45 || program.Repetition().Zoom != .5 {
		t.Fatalf("orbit handoff = %+v, pose = %+v", program, program.Repetition())
	}
	for tick := 429; tick <= 678; tick++ {
		program.Update(kit.Frame{})
	}
	if program.stage != 2 || math.Abs(program.orbit-2) > 1e-12 {
		t.Fatalf("zoom handoff = %+v", program)
	}
}

func TestVivaRotozoomPhaseOptions(t *testing.T) {
	config := DefaultVivaRotozoomConfig(320, 200)
	config.OrbitPhase, config.ZoomPhase, config.RotationPhase = .7, .3, .2
	program, err := NewVivaRotozoom(config)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 205; i++ {
		program.Update(kit.Frame{})
	}
	if program.stage != 1 || program.Repetition().CenterX == config.Width/2 || program.Repetition().Zoom == config.ZoomBase {
		t.Fatalf("phase offsets did not affect the moving pose: %+v", program.Repetition())
	}
	config.EntryPanStep = 0
	if _, err := NewVivaRotozoom(config); err == nil {
		t.Fatal("accepted a stalled entrance")
	}
}

func TestCocoRotozoomStartsInHarmonicStageAndKeepsAllClocks(t *testing.T) {
	config := CocoRotozoom(800, 600)
	program, err := NewVivaRotozoom(config)
	if err != nil {
		t.Fatal(err)
	}
	if program.stage != 3 {
		t.Fatal("Coco entered Viva's staged introduction")
	}
	xPhase, zPhase, rPhase := 0.0, 0.0, 0.0
	for tick := 1; tick <= 750; tick++ {
		speed := 1.0
		if tick >= 300 {
			speed = 1.5
		}
		if err := program.SetSpeedMultiplier(speed); err != nil {
			t.Fatal(err)
		}
		xPhase += .008 * speed
		zPhase += .003 * speed
		rPhase += .005 * speed
		if err := program.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		wantZoom := .5 + math.Abs(math.Sin(zPhase)*2.5)
		wantRotation := 360.0 / 4.0 * math.Cos(rPhase*4-math.Cos(rPhase-.01)) * .3 * math.Pi / 180
		wantX := 400 + 200*math.Cos(xPhase*4-math.Cos(xPhase-.1))
		wantY := 300 + (600.0/2.7)*-math.Sin(xPhase*2.3-math.Cos(xPhase-.1))
		pose := program.Repetition()
		if program.stage != 3 || math.Abs(pose.CenterX-wantX) > 1e-11 ||
			math.Abs(pose.CenterY-wantY) > 1e-11 || math.Abs(pose.Zoom-wantZoom) > 1e-11 ||
			math.Abs(pose.Rotation-wantRotation) > 1e-11 ||
			pose.PhaseX != 1600 || pose.PhaseY != 1200 || pose.Color != ([4]float32{.5, .5, .5, 1}) {
			t.Fatalf("tick %d Coco pose differs: %+v", tick, pose)
		}
	}
}

func TestVivaRotozoomPausePreservesStageAndPose(t *testing.T) {
	config := DefaultVivaRotozoomConfig(320, 200)
	program, err := NewVivaRotozoom(config)
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 160; tick++ {
		if err := program.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := program.SetSpeedMultiplier(0); err != nil {
		t.Fatal(err)
	}
	stage, entryAngles, pose := program.stage, program.entryAngles, program.Repetition()
	for tick := 0; tick < 100; tick++ {
		if err := program.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	if program.stage != stage || program.entryAngles != entryAngles || program.Repetition() != pose {
		t.Fatalf("pause changed rotozoom pose: stage %d -> %d, pose %+v -> %+v", stage, program.stage, pose, program.Repetition())
	}
}
