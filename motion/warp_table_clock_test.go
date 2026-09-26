package motion

import (
	"math"
	"testing"
)

func TestWarpTableClockPreservesSignedRowsAndIndependentVerticalPhase(t *testing.T) {
	clock, err := NewWarpTableClock(WarpTableClockConfig{
		Horizontal: []float64{-3.5, 2.5, 7.5, 0}, SourceOrigin: 64,
		VerticalStart: 1.25, VerticalStep: .1, VerticalSpatial: .2, VerticalAmplitude: 35,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ row, want int }{{-9, 64}, {-1, 64}, {0, 60}, {1, 66}, {6, 71}} {
		if got := clock.SampleX(sample.row); got != sample.want {
			t.Fatalf("row %d sample = %d, want %d", sample.row, got, sample.want)
		}
	}
	for _, column := range []int{-10, 0, 7} {
		want := math.Cos(1.25+float64(column)*.2) * 35
		if got := clock.OffsetY(column); got != want {
			t.Fatalf("column %d offset = %v, want %v", column, got, want)
		}
	}
	if err := clock.Step(1.5); err != nil || clock.Tick() != 1 || clock.Phase() != 1.4 {
		t.Fatalf("step = tick %d phase %g err %v", clock.Tick(), clock.Phase(), err)
	}
	clock.SetTick(9)
	if err := clock.SetPhase(3.25); err != nil || clock.SampleX(0) != 66 {
		t.Fatalf("seek = sample %d err %v", clock.SampleX(0), err)
	}
	clock.Reset()
	if clock.Tick() != 0 || clock.Phase() != 1.25 {
		t.Fatal("reset did not restore both phases")
	}
	if err := clock.SetPhase(math.NaN()); err == nil {
		t.Fatal("accepted nonfinite phase")
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = clock.SampleX(-3)
		_ = clock.OffsetY(11)
		_ = clock.Step(1)
	}); allocations != 0 {
		t.Fatalf("warp clock allocated %v times per frame", allocations)
	}
}
