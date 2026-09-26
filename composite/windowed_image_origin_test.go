package composite

import (
	"image"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestWindowedImageOriginsMatchSixRasterStripPositions(t *testing.T) {
	clock, err := motion.NewWrapBank(motion.WrapBankConfig{
		Start: []float64{174}, Velocity: []float64{-1.5},
		Lower: &motion.WrapLimit{Boundary: -974, Restart: 174, Inclusive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	offset := 174.0
	for tick := 0; tick < 5000; tick++ {
		for i := 0; i < 6; i++ {
			y := 132 + i*34
			window := WindowedImageWindow{
				Clip: image.Rect(0, y, 768, y+32), Offset: motion.Point{Y: -float64(i * 5)},
			}
			want := motion.Point{X: 0, Y: offset - float64(i*5)}
			if got := windowedImageLocalOrigin(window, 0, clock.At(0)); got != want {
				t.Fatalf("tick %d strip %d pose = %+v, want %+v", tick, i, got, want)
			}
		}
		offset -= 1.5
		if offset <= -974 {
			offset = 174
		}
		clock.Step()
	}
	window := WindowedImageWindow{Clip: image.Rect(40, 50, 100, 80), Offset: motion.Point{X: 3, Y: -7}}
	if got := windowedImageLocalOrigin(window, 2, 5); got != (motion.Point{X: 5, Y: -2}) {
		t.Fatalf("two-axis pose = %+v", got)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = windowedImageLocalOrigin(window, 2, 5)
		clock.Step()
	}); allocations != 0 {
		t.Fatalf("window pose allocated %v times per frame", allocations)
	}
}
