package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// GrodanBackgroundPair returns two editable repeated pictures and the source
// screen's gated, boundary-coupled movement program.
func GrodanBackgroundPair(green, pink *ebiten.Image) effects.GatedBackgroundPairConfig {
	background := composite.BackgroundConfig{PeriodX: 640, PeriodY: 400, CopiesX: 3, CopiesY: 2}
	return effects.GatedBackgroundPairConfig{
		Images:      [2]*ebiten.Image{green, pink},
		Backgrounds: [2]composite.BackgroundConfig{background, background},
		Motion: motion.GatedBackgroundPairConfig{
			GateStep: .1, GateOpen: 10, GateReset: 20,
			FirstX: motion.ThresholdAxis{Start: 0, Velocity: 1, Lower: -1280, Upper: 0,
				VelocityBelow: 16, VelocityAbove: -16},
			FirstY: motion.ThresholdAxis{Start: 0, Velocity: 1, Lower: -400, Upper: 0,
				VelocityBelow: 1, VelocityAbove: -1},
			SecondY: motion.ThresholdAxis{Start: 0, Velocity: 1, Lower: -400, Upper: 0,
				VelocityBelow: 2, VelocityAbove: -2},
			SecondXMin: -710, SecondXMax: 0,
			SecondXBelowVelocity: 16, SecondXAboveVelocity: -16,
		},
	}
}
