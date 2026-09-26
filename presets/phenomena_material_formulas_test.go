package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestPhenomenaStageColorsPreserveAuthoredArithmetic(t *testing.T) {
	formulas := PhenomenaStageColorFormulas()
	compiled := make([]*motion.FormulaProgram, 5)
	for i, expression := range []motion.FormulaExpr{
		formulas.Percent, formulas.BrightDown, formulas.WhiteMask, formulas.Green, formulas.Secondary,
	} {
		var err error
		compiled[i], err = motion.CompileFormula(expression)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []float64{0, 1, 17, 33.3, 50, 99, 100, 101, 149, 200} {
		brightness := value / 100.0
		want := []float64{brightness, 2 - brightness, (200 - value) / 100, brightness * .5, 445.25}
		for i, program := range compiled {
			got := program.AtWithSecondaryTime(value, 445.25, 0, 0, 0, 0)
			if math.Abs(got-want[i]) > 1e-15 {
				t.Fatalf("value %g formula %d = %.17g, want %.17g", value, i, got, want[i])
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		for _, program := range compiled {
			_ = program.AtWithSecondaryTime(77, 445.25, 0, 0, 0, 0)
		}
	}); allocations != 0 {
		t.Fatalf("stage color formulas allocated %v times per sample", allocations)
	}
}
