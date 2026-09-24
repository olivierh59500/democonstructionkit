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
