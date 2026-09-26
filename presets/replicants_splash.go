package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// ReplicantsSplash reveals one 40-pixel row every three ticks, then holds the
// complete 640-by-400 picture until the main scene starts at tick 100.
func ReplicantsSplash(image *ebiten.Image) composite.BlockRevealConfig {
	return composite.BlockRevealConfig{Image: image, CellWidth: 640, CellHeight: 40,
		Timing:     motion.SteppedRevealConfig{BlocksPerStep: 1, StepTicks: 3, DoneTick: 100},
		Background: color.RGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}}
}
