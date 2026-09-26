package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// MultiscreenPhenomenaMainMaterials binds the common stage painter to the
// 800x600 embedded panel without constructing unreachable intro stages.
func MultiscreenPhenomenaMainMaterials(logo, raster, photonMask *ebiten.Image) composite.ScalarStagePainterConfig {
	hue := motion.ExprSecondaryTime()
	rule := composite.ScalarStageRule{From: 0, To: 0}
	return composite.ScalarStagePainterConfig{
		Backgrounds: []color.Color{color.Black},
		Passes: []composite.ScalarImagePass{
			{Rule: rule, Rect: &composite.ScalarStageRect{Width: 800, Height: 375,
				Color: color.RGBA{R: 0, G: 1, B: 17, A: 255}}, Y: 162},
			{Rule: rule, Image: logo, X: 80},
			{Rule: rule, Image: raster, Y: 129},
			{Rule: rule, Image: raster, Y: 537},
			{Rule: rule, Image: photonMask, X: 365, Y: 555,
				Tint: composite.ScalarTintConfig{Mode: composite.ScalarTintHSL, H: &hue, Saturation: 1, Lightness: .5}},
		},
	}
}
