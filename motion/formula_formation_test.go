package motion

import (
	"math"
	"testing"
)

func TestCompiledFormulaPreservesNestedWaveProducts(t *testing.T) {
	time, index, width := ExprTime(), ExprIndex(), ExprWidth()
	phase := ExprMul(ExprSub(time, ExprMul(index, ExprConst(.1))), ExprConst(2))
	expr := ExprMul(ExprMul(ExprSin(phase), width), ExprAdd(ExprSin(time), ExprCos(time)))
	program, err := CompileFormula(expr)
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ t, i, w float64 }{{0, 0, 288}, {.5, 1, 288}, {2.3, 11, 288}, {10, 4, 120}} {
		want := math.Sin((sample.t-sample.i*.1)*2) * sample.w * (math.Sin(sample.t) + math.Cos(sample.t))
		if got := program.At(sample.t, sample.i, sample.w, 0, 12); got != want {
			t.Fatalf("sample %+v = %.16g, want %.16g", sample, got, want)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = program.At(2.3, 11, 288, 0, 12) }); allocs != 0 {
		t.Fatalf("formula evaluation allocates %v times", allocs)
	}
}

func TestFormulaCompilerRejectsBadTrees(t *testing.T) {
	for _, expr := range []FormulaExpr{
		{Op: FormulaAdd, Args: []FormulaExpr{ExprConst(1)}},
		{Op: FormulaConstant, Value: math.NaN()},
		{Op: FormulaOp("unknown")},
	} {
		if _, err := CompileFormula(expr); err == nil {
			t.Fatalf("accepted invalid expression: %+v", expr)
		}
	}
}
