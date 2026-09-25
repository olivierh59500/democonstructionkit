package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestTCBMountainWrapConfigPreservesFractionalStripOffsets(t *testing.T) {
	config := TCBMountainWrapConfig()
	bank, err := motion.NewWrapBank(config)
	if err != nil {
		t.Fatal(err)
	}
	var legacy [32]float64
	for tick := 0; tick < 1024; tick++ {
		bank.Step()
		for index := range legacy {
			legacy[index] += config.Velocity[index]
			if legacy[index] <= -256 {
				legacy[index] += 256
			}
			if got := bank.At(index); got != legacy[index] {
				t.Fatalf("tick %d strip %d = %v, want %v", tick, index, got, legacy[index])
			}
		}
	}
}
