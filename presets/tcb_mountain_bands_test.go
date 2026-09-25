package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestTCBMountainBandsMatchStandaloneTransportAndSnapping(t *testing.T) {
	bands, err := composite.NewBands(TCBMountainBands())
	if err != nil {
		t.Fatal(err)
	}
	previous, err := motion.NewWrapBank(TCBMountainWrapConfig())
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 3000; tick++ {
		bands.Step()
		previous.Step()
		for index := 0; index < 32; index++ {
			band, ok := bands.Band(index)
			if !ok || band.PhaseX != previous.At(index) || !band.TruncatePhaseX ||
				math.Trunc(band.PhaseX)*band.MotionScaleX != float64(int(previous.At(index))*2) {
				t.Fatalf("tick %d strip %d changed phase or integer placement: %+v", tick, index, band)
			}
		}
	}
}
