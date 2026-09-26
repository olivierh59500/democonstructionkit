package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestBilizirWarpClockMatchesTextAndLogoAcrossSpeedChanges(t *testing.T) {
	config, err := BilizirWarpClockConfig()
	if err != nil {
		t.Fatal(err)
	}
	clock, err := motion.NewWarpTableClock(config)
	if err != nil {
		t.Fatal(err)
	}
	table, err := motion.CompileWaveTable(BilizirWaveSections()...)
	if err != nil {
		t.Fatal(err)
	}
	oldTick, oldPhase := uint64(0), 0.0
	for tick := 0; tick < 5000; tick++ {
		for _, row := range []int{-150, -40, -1, 0, 17, 99, 400} {
			index := (int(oldTick%uint64(len(table))) + row) % len(table)
			if index < 0 {
				index += len(table)
			}
			want := int(table[index] + 64)
			if got := clock.SampleX(row); got != want {
				t.Fatalf("tick %d row %d sample = %d, want %d", tick, row, got, want)
			}
		}
		for _, column := range []int{-32, -1, 0, 7, 49, 112} {
			want := math.Cos(oldPhase+float64(column)*.1) * 35
			if got := clock.OffsetY(column); math.Abs(got-want) > 1e-12 {
				t.Fatalf("tick %d column %d offset = %g, want %g", tick, column, got, want)
			}
		}
		speed := 1.0
		if tick >= 1500 && tick < 3000 {
			speed = 1.5
		} else if tick >= 3000 {
			speed = .5
		}
		oldTick++
		oldPhase += .1 * speed
		if err := clock.Step(speed); err != nil {
			t.Fatal(err)
		}
	}
	if clock.Tick() != oldTick || clock.Phase() != oldPhase {
		t.Fatalf("final clock = %d/%g, want %d/%g", clock.Tick(), clock.Phase(), oldTick, oldPhase)
	}
}
