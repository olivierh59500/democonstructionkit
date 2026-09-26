package presets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// UnionDiskCopierRasterWindows keeps six phase-offset raster views in their
// authored order without allocating six intermediate strip surfaces.
func UnionDiskCopierRasterWindows(raster *ebiten.Image) composite.WindowedImageBankConfig {
	windows := make([]composite.WindowedImageWindow, 6)
	for i := range windows {
		y := 132 + i*34
		windows[i] = composite.WindowedImageWindow{
			Clip:   image.Rect(0, y, 768, y+32),
			Offset: motion.Point{Y: -float64(i * 5)},
		}
	}
	return composite.WindowedImageBankConfig{
		Image: raster, Windows: windows, ScaleX: 1.3, ScaleY: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
		MotionY: &motion.WrapBankConfig{
			Start: []float64{174}, Velocity: []float64{-1.5},
			Lower: &motion.WrapLimit{Boundary: -974, Restart: 174, Inclusive: true},
		},
	}
}
