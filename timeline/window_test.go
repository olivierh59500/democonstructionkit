package timeline

import (
	"math"
	"testing"
)

func TestWindowBoundariesFadesAndSeeking(t *testing.T) {
	w := Window{Start: 10, Duration: 4, FadeIn: 1, FadeOut: 2}
	for _, tt := range []struct {
		time, alpha float64
		active      bool
	}{{9, 0, false}, {10, 0, true}, {10.5, .5, true}, {11, 1, true}, {13, .5, true}, {14, 0, false}, {10.5, .5, true}} {
		_, a, on := w.At(tt.time)
		if a != tt.alpha || on != tt.active {
			t.Fatal(tt, a, on)
		}
	}
	if _, a, on := (Window{}).At(1e9); !on || a != 1 {
		t.Fatal("unbounded layer ended")
	}
	for _, w := range []Window{{Duration: -1}, {Start: math.NaN()}, {FadeOut: 1}, {FadeIn: -1}} {
		if w.Validate() == nil {
			t.Fatal("invalid window accepted", w)
		}
	}
}
