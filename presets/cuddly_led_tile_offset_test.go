package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCuddlyLEDTileOffsetPreservesPreloadAndWrap(t *testing.T) {
	clock, err := motion.NewWrapBank(CuddlyLEDTileOffset())
	if err != nil {
		t.Fatal(err)
	}
	offset := 0.0
	for tick := 0; tick < 2000; tick++ {
		if got := clock.At(0); got != offset {
			t.Fatalf("tick %d backdrop = %v, want %v", tick, got, offset)
		}
		offset -= 2
		if offset <= -33 {
			offset = 0
		}
		clock.Step()
		if tick == 0 && clock.At(0) != -2 {
			t.Fatal("first visible frame lost the one-tick preload")
		}
	}
	if allocations := testing.AllocsPerRun(100, clock.Step); allocations != 0 {
		t.Fatalf("tile clock allocated %v times per step", allocations)
	}
}
