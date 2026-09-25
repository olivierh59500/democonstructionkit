package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// VivaLogoFormation is the editable ten-logo harmonic train. verticalAmplitude
// retains the different authored spacing in standalone Viva and Multiscreen.
func VivaLogoFormation(image *ebiten.Image, width, height, verticalAmplitude float64) sprites.RecurrentFormationConfig {
	centerX := width/4 - 16
	return sprites.RecurrentFormationConfig{
		Image: image, Count: 10,
		Origin:     motion.Point{X: centerX, Y: 24 + height/4 - 16},
		XAmplitude: centerX, YAmplitude: verticalAmplitude, YSecondaryAmplitude: verticalAmplitude,
		XDivisor: 25, XSecondaryDivisor: 300, YDivisor: 37, YSecondaryDivisor: 17,
		XIndexStep: .2, XSecondaryStep: 1.0 / 60.0,
		YIndexStep: 5.0 / 37.0, YSecondaryStep: 5.0 / 17.0,
		YSecondaryCos: true,
		ScaleX:        2, ScaleY: 2, OutputScaleX: 2, OutputScaleY: 2,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}
