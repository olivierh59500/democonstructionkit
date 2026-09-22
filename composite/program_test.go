package composite

import (
	"math"
	"slices"
	"testing"
)

func TestCumulativeProgramIntroLoopAndNegativeOffsets(t *testing.T) {
	p, err := NewDisplacementProgram([]int{2, 5}, []int{1, 3, 7})
	if err != nil {
		t.Fatal(err)
	}
	want := []int{2, 5, 6, 8, 12, 13, 15, 19, 20}
	got := make([]int, len(want))
	p.Fill(got, 0)
	if !slices.Equal(got, want) {
		t.Fatal(got)
	}
	for i := range want {
		if p.At(i) != want[i] {
			t.Fatal(i)
		}
	}
	for _, start := range []int{-20, -3, -1, 0, 1, 2, 4, 8, 80} {
		p.Fill(got, start)
		for i := range got {
			if got[i] != p.At(start+i) {
				t.Fatalf("start%d sample%d: %d!=%d", start, i, got[i], p.At(start+i))
			}
		}
	}
	if got := testing.AllocsPerRun(100, func() { p.Fill(got, 12345) }); got != 0 {
		t.Fatal(got)
	}
}
func TestDeltaCurveDriftAndSequentialComposition(t *testing.T) {
	first, err := (DeltaCurve{Step: 1, Extent: 4, Drift: 8}).Compile()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(first, []int{0, 2, 2, 2}) {
		t.Fatal(first)
	}
	values, err := JoinDeltaCurves([][]int{first, {1, -2, 3}}, []int{0, 1, 0})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(values, []int{0, 2, 4, 6, 7, 5, 8, 8, 10, 12, 14}) {
		t.Fatal(values)
	}
	for _, c := range []DeltaCurve{{}, {Step: -1, Extent: 2}, {Step: math.SmallestNonzeroFloat64, Extent: 2}, {Step: 2, Extent: 1, OmitLastStep: true}, {Step: 1, Extent: 3, Terms: []CurveTerm{{Amplitude: math.NaN()}}}} {
		if _, err := c.Compile(); err == nil {
			t.Fatal("invalid curve accepted")
		}
	}
	if _, err = JoinDeltaCurves([][]int{first}, []int{1}); err == nil {
		t.Fatal("invalid index accepted")
	}
}
