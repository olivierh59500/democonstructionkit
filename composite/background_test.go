package composite

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
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
		{CopiesX: -1}, {CopiesY: -1}, {CopiesX: 1<<20 + 1},
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

func TestBackgroundLimitedCopies(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		origin, extent, period, min, max float64
		budget, copies                   int
		positions                        []float64
		wantOK                           bool
	}{
		{"two copies", 0, 4, 5, 0, 16, 16, 2, []float64{0, 5}, true},
		{"partial first", -6, 4, 5, 0, 16, 16, 3, []float64{-1, 4}, true},
		{"all outside", -100, 4, 5, 0, 16, 16, 3, nil, true},
		{"dense but bounded", 0, 32, 1e-100, 0, 640, 16, 3, []float64{0, 1e-100, 2e-100}, true},
		{"bounded over budget", 0, 32, 1e-100, 0, 640, 2, 3, nil, false},
		{"no distant rebase", -10000000, 4, 5, 0, 16, 16, 3, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			origin, first, last, ok := backgroundCopyRangeLimit(tc.origin, tc.extent, tc.period, tc.min, tc.max, tc.budget, tc.copies)
			if ok != tc.wantOK {
				t.Fatalf("range accepted = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got := max(0, last-first+1); got != len(tc.positions) {
				t.Fatalf("copies = %d, want %d", got, len(tc.positions))
			}
			for i, want := range tc.positions {
				got := origin + float64(first+i)*tc.period
				if got != want {
					t.Fatalf("copy %d at %g, want %g", i, got, want)
				}
			}
		})
	}
}

func TestBackgroundLayerVelocity(t *testing.T) {
	renderer, err := NewBackground(DefaultBackgroundConfig())
	if err != nil {
		t.Fatal(err)
	}
	layer := &BackgroundLayer{Renderer: renderer, Pose: BackgroundPose{X: 10, Y: 3, CameraX: 8}, VelocityX: -12, VelocityY: 7}
	if err := layer.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	if got := layer.poseAt(); got != (BackgroundPose{X: -14, Y: 17, CameraX: 8}) {
		t.Fatalf("velocity pose = %+v", got)
	}
	layer.Sample = func(f kit.Frame) BackgroundPose { return BackgroundPose{X: f.Time * 3, CameraY: 4} }
	if got := layer.poseAt(); got != (BackgroundPose{X: -18, Y: 14, CameraY: 4}) {
		t.Fatalf("sampled pose = %+v", got)
	}
	layer.VelocityY = math.Inf(1)
	if err := layer.Update(kit.Frame{Time: 2}); err == nil {
		t.Fatal("nonfinite velocity accepted")
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
