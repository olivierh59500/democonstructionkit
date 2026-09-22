package motion

import (
	"math"
	"testing"
)

func TestBSpline32LinearControlPointsAndExtrapolation(t *testing.T) {
	keys := []float32{0, 0, 10, 1, 1, 20, 2, 2, 30, 3, 3, 40, 4, 4, 50}
	spline, err := NewBSpline32(keys, 2)
	if err != nil {
		t.Fatal(err)
	}
	keys[1] = 1000
	var output [2]float32
	for _, at := range []float32{-1, 0, .5, 1, 1.5, 2, 3} {
		if !spline.Sample(output[:], at) {
			t.Fatal("sample rejected", at)
		}
		if math.Abs(float64(output[0]-(at+1))) > 1e-5 || math.Abs(float64(output[1]-(at+2)*10)) > 1e-4 {
			t.Fatalf("at %g: %v", at, output)
		}
	}
	if spline.Components() != 2 {
		t.Fatal("component count changed")
	}
	if allocations := testing.AllocsPerRun(100, func() { spline.Sample(output[:], .25) }); allocations != 0 {
		t.Fatalf("sample allocations = %g", allocations)
	}
	before := output
	if spline.Sample(output[:1], 1) || spline.Sample(output[:], float32(math.NaN())) || output != before {
		t.Fatal("invalid sample must preserve output")
	}
}

func TestBSpline32RejectsInvalidKeys(t *testing.T) {
	for _, keys := range [][]float32{
		nil, {0, 1, 2, 3, 4}, {0, 1, 0, 2, 1, 3, 2, 4},
		{0, 1, 1, 2, 2, 3, 3, float32(math.Inf(1))},
	} {
		if _, err := NewBSpline32(keys, 1); err == nil {
			t.Fatal("accepted malformed B-spline", keys)
		}
	}
}
