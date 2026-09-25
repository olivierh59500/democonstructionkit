package composite

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestRasterTitleKeepsTwoPhasesAndModeSurfaceBudgets(t *testing.T) {
	title, raster, dst := ebiten.NewImage(528, 36), ebiten.NewImage(24, 144), ebiten.NewImage(800, 600)
	defer title.Deallocate()
	defer raster.Deallocate()
	defer dst.Deallocate()
	for _, mode := range []RasterTitleMode{RasterTitleCanvas, RasterTitleDirect} {
		config := RasterTitleConfig{
			Title: title, Raster: raster, Mode: mode,
			Motion: motion.WrapBankConfig{
				Start: []float64{0, 72}, Velocity: []float64{-2},
				Lower: &motion.WrapLimit{Boundary: -72, Restart: 72, Inclusive: true},
			},
			RasterScaleX: 24, RasterScaleY: 1, ThirdOffsetY: 72,
			Clip: image.Rect(0, 14, 800, 86), UnderlayWidth: 800, UnderlayHeight: 72,
		}
		effect, err := NewRasterTitle(config)
		if err != nil {
			t.Fatal(err)
		}
		if (effect.canvas != nil) != (mode == RasterTitleCanvas) {
			t.Fatal("raster title allocated an unexpected GPU surface")
		}
		phases := [2]float64{0, 72}
		for tick := 0; tick < 400; tick++ {
			effect.Step()
			for index := range phases {
				phases[index] -= 2
				if phases[index] <= -72 {
					phases[index] = 72
				}
				if effect.Motion().At(index) != phases[index] {
					t.Fatalf("mode %d tick %d phase %d = %v, want %v", mode, tick, index, effect.Motion().At(index), phases[index])
				}
			}
		}
		effect.DrawAt(dst, 64, 14)
		if err := effect.Close(); err != nil {
			t.Fatal(err)
		}
	}
	mirrored, err := NewRasterTitle(RasterTitleConfig{
		Title: title, Raster: raster, Mode: RasterTitleCanvas,
		Motion:       motion.WrapBankConfig{Start: []float64{0, 72}, Velocity: []float64{-2}, Lower: &motion.WrapLimit{Boundary: -72, Restart: 72}},
		RasterScaleX: -1, RasterScaleY: 1, ThirdOffsetY: -36,
	})
	if err != nil {
		t.Fatalf("rejected signed raster variation: %v", err)
	}
	mirrored.Close()
}
