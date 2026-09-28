package scrolling

import (
	"reflect"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func cuedFacadeConfig() CuedSlicesConfig {
	return CuedSlicesConfig{
		Program: SliceProgramConfig{
			Stream: SliceStreamConfig{Tokens: []SliceToken{
				{Glyph: 0, Width: 7}, {Control: "reverse"}, {Glyph: 1, Width: 5},
				{Control: "forward"}, {Glyph: 0, Width: 7},
			}, Capacity: 16, SliceWidth: 2, Repeat: true, LoopStart: 1},
			Clock: motion.CuedScrollClockConfig{InitialTextStep: 1, InitialRotationStep: .35,
				ResumeTextStep: 1, ResumeRotationStep: .35, RotationFrames: 8,
				Cues: map[string]motion.ScrollCue{
					"reverse": {PauseTicks: 7, SetRotation: true, RotationStep: -1},
					"forward": {PauseTicks: 5, SetRotation: true, RotationStep: 1},
				}},
		},
		Draw: DNADrawConfig{SliceWidth: 2, ScaleX: 1, ScaleY: 1},
	}
}

func TestCuedSliceFacadePreservesPreRollControlsAndAllocationBudget(t *testing.T) {
	config := cuedFacadeConfig()
	reference, err := NewSliceProgram(config.Program)
	if err != nil {
		t.Fatal(err)
	}
	scroll, err := New(Config{CuedSlices: &config})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	if err := reference.Warmup(11, 1); err != nil {
		t.Fatal(err)
	}
	if err := scroll.SliceProgramController().Warmup(11, 1); err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 5000; tick++ {
		if err := reference.Step(); err != nil {
			t.Fatal(err)
		}
		if err := scroll.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
			t.Fatal(err)
		}
		got := scroll.SliceProgramController()
		if got.Clock().State() != reference.Clock().State() || got.Stream().Head() != reference.Stream().Head() ||
			!reflect.DeepEqual(got.Stream().Slices(), reference.Stream().Slices()) {
			t.Fatalf("pause, rotation or strip state changed at %d", tick)
		}
	}
	if allocations := testing.AllocsPerRun(1000, func() { _ = scroll.Update(kit.Frame{}) }); allocations != 0 {
		t.Fatalf("facade update allocated %v times", allocations)
	}
	if _, err := New(Config{CuedSlices: &config, Sliced: &SlicedConfig{}}); err == nil {
		t.Fatal("ambiguous DNA clocks accepted")
	}
	config.Draw.SliceWidth = 1
	if _, err := New(Config{CuedSlices: &config}); err == nil {
		t.Fatal("rendering width differs from the insertion strip width")
	}
}
