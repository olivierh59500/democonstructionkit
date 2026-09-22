package composite

import (
	"math"
	"testing"
)

func TestBackgroundCopies(t *testing.T) {
	for _, c := range []struct {
		name                             string
		origin, extent, period, min, max float64
		positions                        []float64
	}{
		{"single", 3, 8, 0, 0, 16, []float64{3}},
		{"outside", -8, 8, 0, 0, 16, nil},
		{"touching", 16, 8, 0, 0, 16, nil},
		{"regular", 0, 8, 8, 0, 16, []float64{0, 8}},
		{"negative", -3, 8, 8, 0, 16, []float64{-3, 5, 13}},
		{"positive", 5, 8, 8, 0, 16, []float64{-3, 5, 13}},
		{"long elapsed", -8000003, 8, 8, 0, 16, []float64{-3, 5, 13}},
		{"viewport origin", 3, 8, 8, 20, 36, []float64{19, 27, 35}},
		{"gaps", 1, 3, 8, 0, 16, []float64{1, 9}},
		{"overlap", 0, 12, 8, 0, 16, []float64{-8, 0, 8}},
		{"fractional", -.25, 3.5, 3.5, 0, 7, []float64{-.25, 3.25, 6.75}},
	} {
		t.Run(c.name, func(t *testing.T) {
			origin, first, last := backgroundCopies(c.origin, c.extent, c.period, c.min, c.max)
			if max(0, last-first+1) != len(c.positions) {
				t.Fatalf("copies %d..%d, want %d", first, last, len(c.positions))
			}
			for i, want := range c.positions {
				got := origin + float64(first+i)*c.period
				if got != want {
					t.Fatalf("copy %d at %g, want %g", i, got, want)
				}
			}
		})
	}
}

func TestBackgroundConfiguration(t *testing.T) {
	for _, c := range []BackgroundConfig{
		{PeriodX: -1}, {PeriodY: -1}, {ScaleX: -1}, {ScaleY: -1},
		{PeriodX: math.Inf(1)}, {ScaleY: math.NaN()}, {ParallaxX: math.NaN()},
	} {
		if _, err := NewBackground(c); err == nil {
			t.Fatalf("invalid config accepted: %+v", c)
		}
	}
	b, err := NewBackground(BackgroundConfig{})
	if err != nil || b.config.ScaleX != 1 || b.config.ScaleY != 1 {
		t.Fatal("zero scales should default to one", err)
	}
	c := DefaultBackgroundConfig()
	if c.ParallaxX != 1 || c.ParallaxY != 1 {
		t.Fatal("default camera should follow both axes")
	}
}

func BenchmarkBackgroundCopies(b *testing.B) {
	for i := 0; i < b.N; i++ {
		backgroundCopies(-float64(i)*4, 640, 640, 0, 640)
	}
}

func TestBackgroundRejectsUnboundedCopyRanges(t *testing.T) {
	for _, period := range []float64{1e-6, 1e-100, math.SmallestNonzeroFloat64} {
		if _, _, _, ok := backgroundCopyRange(0, 32, period, 0, 640, 16384); ok {
			t.Fatalf("unbounded density accepted: %g", period)
		}
	}
	if _, _, _, ok := backgroundCopyRange(0, 32, 16, 0, 640, 16384); !ok {
		t.Fatal("ordinary overlaps rejected")
	}
}
