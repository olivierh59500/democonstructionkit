package composite

import (
	"math"
	"testing"
)

func TestThinRowRetainsTotalCoverage(t *testing.T) {
	for _, height := range []float64{.01, .2, .7, 1, 3.25} {
		for _, top := range []float64{-2.3, 0, 1.9, 4.5} {
			sum := 0.0
			for y := int(math.Floor(top)) - 1; y <= int(math.Ceil(top+height)); y++ {
				c := rowCoverage(top, height, y)
				if c < 0 || c > 1 {
					t.Fatalf("invalid coverage %g", c)
				}
				sum += c
			}
			if math.Abs(sum-height) > 1e-12 {
				t.Fatalf("height %g lost coverage: %g", height, sum)
			}
		}
	}
}
