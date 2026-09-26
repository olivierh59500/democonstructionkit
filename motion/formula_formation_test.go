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

func TestFormulaFloorPreservesHalfUpPixelSnapping(t *testing.T) {
	phase := ExprAdd(ExprDiv(ExprTime(), ExprConst(71)), ExprDiv(ExprIndex(), ExprConst(47)))
	wave := ExprMul(ExprConst(151), ExprSin(ExprMul(ExprConst(5), phase)))
	expr := ExprFloor(ExprAdd(wave, ExprConst(.5)))
	program, err := CompileFormula(expr)
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 400; tick++ {
		for index := 0; index < 20; index++ {
			phase := float64(tick)/71 + float64(index)/47
			want := math.Floor(151*math.Sin(5*phase) + .5)
			if got := program.At(float64(tick), float64(index), 0, 0, 20); got != want {
				t.Fatalf("tick %d index %d = %v, want %v", tick, index, got, want)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { program.At(37, 12, 0, 0, 20) }); allocations != 0 {
		t.Fatalf("floor formula allocated %v times", allocations)
	}
}

func TestFormulaFormationMatchesSpreadpointBallPath(t *testing.T) {
	phase := ExprAdd(ExprDiv(ExprTime(), ExprConst(71)), ExprDiv(ExprIndex(), ExprConst(47)))
	axis := func(base, amplitude, rate float64) FormulaExpr {
		return ExprAdd(ExprConst(base), ExprFloor(ExprAdd(
			ExprMul(ExprConst(amplitude), ExprSin(ExprMul(ExprConst(rate), phase))), ExprConst(.5))))
	}
	formation, err := NewFormulaFormation(FormulaFormationConfig{
		X: axis(203, 151, 5), Y: axis(79, 50, 8),
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 400; tick++ {
		for index := 0; index < 20; index++ {
			p := float64(tick)/71 + float64(index)/47
			want := Point{X: 203 + math.Floor(151*math.Sin(5*p)+.5),
				Y: 79 + math.Floor(50*math.Sin(8*p)+.5)}
			if got := formation.At(float64(tick), index, 0, 0, 20); got != want {
				t.Fatalf("tick %d ball %d = %+v, want %+v", tick, index, got, want)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { formation.At(37, 12, 0, 0, 20) }); allocations != 0 {
		t.Fatalf("ball formation allocated %v times per sample", allocations)
	}
}
