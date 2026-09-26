package presets

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// ReplicantsLogoPair preserves the two interleaved zoom paths while leaving
// artwork, phase speeds, frame counts and depth ordering editable.
func ReplicantsLogoPair(repLogo, tcbLogo *ebiten.Image) sprites.CoupledLogoPairConfig {
	return sprites.CoupledLogoPairConfig{
		Motion: motion.CoupledLogoConfig{
			StepPrimary: 2.0 / 180.0 * math.Pi, StepSecondary: 7.0 / 180.0 * math.Pi,
			SpeedMultiplier: 1.4,
			DepthBase:       1, DepthSin: 1, DepthDivisor: 2,
			YBase: 140, YSumAmplitude: 40,
			SecondaryDepthCos: 1.0 / 8.0, SecondaryYSin: 70,
		},
		Primary: sprites.QuantizedLogoLayer{Image: repLogo, FrameCount: 35,
			Gain: 35, HideFirst: true},
		Secondary: sprites.QuantizedLogoLayer{Image: tcbLogo, FrameCount: 40,
			Bias: 4, Gain: 32, HideFirst: true},
		CenterX: 320,
	}
}
