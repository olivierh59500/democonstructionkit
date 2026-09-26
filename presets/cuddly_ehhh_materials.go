package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyEhhhInnerRollWrap keeps the authored eight-position raster cycle.
func CuddlyEhhhInnerRollWrap() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-2},
		Lower: &motion.WrapLimit{Boundary: -16, Restart: 0, Inclusive: true},
	}
}

// CuddlyEhhhMiddleRaster fills the live middle text without another surface.
func CuddlyEhhhMiddleRaster(image *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{Image: image,
		ScaleX: 1, ScaleY: 1, Alpha: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop}
}
