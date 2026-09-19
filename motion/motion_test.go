package motion

import "testing"

func TestWrapAndTableAcrossBoundaries(t *testing.T) {
	if Wrap(-101, 10) != 9 || Wrap(101, 10) != 1 {
		t.Fatal("wrap failed")
	}
	table := Table{Values: []float64{0, 10}, Rate: 1}
	if table.At(-.5) != 5 || table.At(2.5) != 5 {
		t.Fatal("table discontinuity")
	}
}
func TestTweenBoundaries(t *testing.T) {
	for _, e := range []Ease{Linear, Smooth, Sine, ElasticIn, ElasticOut} {
		if Tween(2, 8, 3, 4, 2, e) != 2 || Tween(2, 8, 3, 4, 7, e) != 8 {
			t.Fatal("incorrect endpoints")
		}
	}
	if Tween(2, 8, 3, 0, 3, nil) != 8 {
		t.Fatal("instant cut failed")
	}
}
