package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestVivaRasterWrapConfigKeepsPairedTitleBands(t *testing.T) {
	bank, err := motion.NewWrapBank(VivaRasterWrapConfig())
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 35; tick++ {
		bank.Step()
	}
	if bank.At(0) != -70 || bank.At(1) != 2 {
		t.Fatalf("before lower wrap: %v, %v", bank.At(0), bank.At(1))
	}
	bank.Step()
	if bank.At(0) != 72 || bank.At(1) != 0 {
		t.Fatalf("after lower wrap: %v, %v", bank.At(0), bank.At(1))
	}
}
