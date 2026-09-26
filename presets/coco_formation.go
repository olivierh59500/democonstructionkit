package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CocoLogoFormation arranges sixteen independently drawable images in a
// four-by-four grid with one shared two-harmonic movement on each axis.
func CocoLogoFormation(image *ebiten.Image, width, height float64) sprites.GroupConfig {
	return sprites.GroupConfig{
		Frames: []*ebiten.Image{image}, Count: 16,
		Origin: motion.Point{X: width / 2, Y: 72 + (height-72)/2},
		Grid:   &sprites.GridFormation{Columns: 4, Rows: 4, StepX: 200, StepY: 140},
		Translation: &motion.HarmonicTranslation{
			X: []motion.HarmonicTerm{
				{Amplitude: 100, Rate: 1.35, Phase: 1.25},
				{Amplitude: 100, Rate: 1.86, Phase: .54},
			},
			Y: []motion.HarmonicTerm{
				{Amplitude: 60, Rate: 1.72, Phase: .23, Cos: true},
				{Amplitude: 60, Rate: 1.63, Phase: .98, Cos: true},
			},
		},
		ScaleX: .5, ScaleY: .5, AnchorX: .5, AnchorY: .5,
		Opacity: .6, AlphaOnly: true,
	}
}

// MultiscreenCocoLogoFormation keeps Coco's editable grid and material while
// advancing its four shared harmonics by recurrence with periodic reanchoring.
func MultiscreenCocoLogoFormation(image *ebiten.Image, width, height float64) sprites.GroupConfig {
	c := CocoLogoFormation(image, width, height)
	harmonics := *c.Translation
	c.Translation = nil
	c.RecurrentTranslation = &motion.RecurrentTranslationConfig{
		Harmonics: harmonics, PhaseStep: .02, ReanchorEvery: 1024,
		XStepDeltas: []float64{.02 * 1.35, .02 * 1.86},
		YStepDeltas: []float64{.02 * 1.72, .02 * 1.63},
	}
	return c
}
