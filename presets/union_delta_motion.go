package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// UnionDeltaLogoFlip keeps the two logo tiles until the lower scale crossing.
func UnionDeltaLogoFlip() motion.BounceToggleConfig {
	return motion.BounceToggleConfig{Start: 1, Velocity: -.02, Min: 0, Max: 1,
		Inclusive: true, ToggleLower: true, Materials: 2}
}

// UnionDeltaGoldWrap is the three-pixel vertical movement of the gold fill.
func UnionDeltaGoldWrap() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-3},
		Lower: &motion.WrapLimit{Boundary: -59, Restart: 0, Inclusive: true},
	}
}

// UnionDeltaGoldMaterial reuses the same live gold surface under both the
// introductory title and main scrolling text through source-atop blending.
func UnionDeltaGoldMaterial(image *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{Image: image, ScaleX: 1, ScaleY: 1,
		Alpha: 1, Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop}
}
