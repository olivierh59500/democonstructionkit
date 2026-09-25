package composite

import (
	"math"
	"testing"
)

func TestRowWarpKeepsIndependentStripAndVerticalClocks(t *testing.T) {
	wave := []float64{4, 12, -7}
	warp, err := NewRowWarp(RowWarpConfig{
		Mode: RowWarpSourceX, Thickness: 2, SourceX: 64, SourceWidth: 640,
		Wave: wave, WaveStep: 1, VerticalBase: 30, VerticalAmplitude: 30,
		VerticalDivisor: 20, VerticalStep: 1.2,
	})
	if err != nil {
		t.Fatal(err)
	}
	wave[1] = 100
	if warp.Horizontal(0) != 4 || warp.Vertical() != 60 {
		t.Fatal("initial row warp phase changed")
	}
	for tick := 1; tick < 1200; tick++ {
		if err := warp.Step(); err != nil {
			t.Fatal(err)
		}
		if got, want := warp.Horizontal(0), []float64{4, 12, -7}[tick%3]; got != want {
			t.Fatalf("tick %d horizontal = %v, want %v", tick, got, want)
		}
		if got, want := warp.Vertical(), 30+30*math.Cos(warp.verticalPhase/20); got != want {
			t.Fatalf("tick %d vertical = %v, want %v", tick, got, want)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = warp.Step() }); allocations != 0 {
		t.Fatalf("row warp step allocates %v times", allocations)
	}
}

func TestRowWarpRejectsInvalidSourceSampling(t *testing.T) {
	for _, config := range []RowWarpConfig{
		{Mode: RowWarpSourceX, Thickness: 2, SourceWidth: 0},
		{Mode: RowWarpDestinationX, Thickness: 0},
		{Mode: RowWarpDestinationX, Thickness: 2, Wave: []float64{math.NaN()}},
	} {
		if _, err := NewRowWarp(config); err == nil {
			t.Fatalf("accepted invalid row warp: %+v", config)
		}
	}
}
