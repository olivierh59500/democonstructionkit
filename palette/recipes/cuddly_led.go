package recipes

import (
	"image/color"

	"github.com/olivierh59500/democonstructionkit/palette"
)

// CuddlyLEDGradient preserves eleven editable color stops and center-of-row
// sampling across the original 384x2000 raster material.
func CuddlyLEDGradient() palette.UniformGradientConfig {
	return palette.UniformGradientConfig{
		Width: 384, Height: 2000, Axis: palette.GradientVertical,
		CenterSamples: true, RoundNearest: true,
		Colors: []color.RGBA{
			{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255},
			{G: 255, A: 255}, {R: 255, A: 255}, {G: 255, A: 255},
			{B: 255, A: 255}, {R: 255, A: 255}, {G: 255, A: 255},
			{B: 255, A: 255}, {G: 255, A: 255},
		},
	}
}
