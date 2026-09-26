package composite

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestScalarStageRuleKeepsExactBoundaryAndDirection(t *testing.T) {
	showDim := ScalarStageRule{From: 2, To: 2, Compare: timeline.ScalarLessEqual, Threshold: 100}
	showMask := ScalarStageRule{From: 2, To: 2, Compare: timeline.ScalarGreater, Threshold: 100}
	for _, value := range []float64{0, 50, 100, 101, 200} {
		if showDim.Matches(2, value, 1) != (value <= 100) || showMask.Matches(2, value, 1) != (value > 100) {
			t.Fatalf("logo boundary %g selected wrong material", value)
		}
	}
	if showDim.Matches(1, 50, 1) || showMask.Matches(3, 150, 1) {
		t.Fatal("pass escaped its stage")
	}
	logoOut := ScalarStageRule{From: 8, To: 8, DirectionSign: 1}
	if !logoOut.Matches(8, 50, 1) || logoOut.Matches(8, 50, -1) || logoOut.Matches(8, 50, 0) {
		t.Fatal("outro direction changed logo draw order")
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = showDim.Matches(2, 77, 1)
	}); allocations != 0 {
		t.Fatalf("stage rule allocated %v times per sample", allocations)
	}
}
