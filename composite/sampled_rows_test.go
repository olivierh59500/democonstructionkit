package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type testRowSampler struct{ invalid bool }

func (s *testRowSampler) Sample(time float64, copyIndex, row int) motion.SampledRowPose {
	if s.invalid {
		return motion.SampledRowPose{X: math.NaN()}
	}
	return motion.SampledRowPose{SourceY: float64(row), SourceHeight: 1,
		X: time + float64(copyIndex*10+row), Y: float64(row), ScaleX: 1, ScaleY: 1}
}

func TestSampledRowsUsesAbsoluteSceneClockAndRetainsLastValidPose(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	defer image.Deallocate()
	program := &testRowSampler{}
	rows, err := NewSampledRows(SampledRowsConfig{
		Image: image, Program: program, Rows: 2, Copies: 2,
		SourceWidth: 4, UseTime: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := rows.Update(kit.Frame{Time: 804}); err != nil {
		t.Fatal(err)
	}
	for i, want := range []float64{804, 805, 814, 815} {
		if got := rows.Poses()[i].X; got != want {
			t.Fatalf("pose %d X = %v, want %v", i, got, want)
		}
	}
	program.invalid = true
	if err := rows.Update(kit.Frame{Time: 900}); err == nil {
		t.Fatal("accepted a nonfinite row pose")
	}
	if rows.Time() != 804 || rows.Poses()[0].X != 804 {
		t.Fatal("invalid update discarded the last valid frame")
	}
}
