package composite

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestLevel16RasterPhasesMatchPreviousWrapBanks(t *testing.T) {
	for _, sample := range []struct {
		name                               string
		start, velocity, boundary, restart float64
		upper                              bool
	}{
		{"water", 0, 2, 220, 0, true},
		{"raster", 120, -2, -25, 120, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			config := motion.WrapBankConfig{Start: []float64{sample.start}, Velocity: []float64{sample.velocity}}
			if sample.upper {
				config.Upper = &motion.WrapLimit{Boundary: sample.boundary, Restart: sample.restart, Inclusive: true}
			} else {
				config.Lower = &motion.WrapLimit{Boundary: sample.boundary, Restart: sample.restart, Inclusive: true}
			}
			old, err := motion.NewWrapBank(config)
			if err != nil {
				t.Fatal(err)
			}
			wrap := &RasterWrap{Boundary: sample.boundary, Restart: sample.restart, Inclusive: true}
			phase := sample.start
			for tick := 0; tick < 5000; tick++ {
				if got := old.At(0); got != phase {
					t.Fatalf("tick %d phase = %g, want %g", tick, phase, got)
				}
				old.Step()
				phase = rasterNext(phase, sample.velocity, wrap)
			}
			if allocations := testing.AllocsPerRun(100, func() {
				phase = rasterNext(phase, sample.velocity, wrap)
			}); allocations != 0 {
				t.Fatalf("phase update allocated %v times", allocations)
			}
		})
	}
}
