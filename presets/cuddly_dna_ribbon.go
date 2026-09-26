package presets

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyDNATwist keeps the source's two-sided strip phases and occlusion
// thresholds. Changing images or any returned parameter creates a new ribbon.
func CuddlyDNATwist() motion.TwistingRibbonConfig {
	return motion.TwistingRibbonConfig{
		Width: 320, StripWidth: 16, IndexStep: 6,
		PhaseStep: 8, PhaseWrap: 3840, FarStart: 1280,
		NearAmplitude: 15, FarAmplitude: 30, NearScale: 1.5, FarScale: 1,
		AngleMultiplier: 8, NearDivisor: 1280, FarDivisor: 2560,
		BaseAngle: math.Pi + .4, FaceGap: 1.12, BaseY: 30,
		BackVisibleBelow: 1.5, BackVisibleAbove: 3.6,
		FrontVisibleAbove: .5, FrontVisibleBelow: 4.6,
		UpperClipStart: .5, UpperClipEnd: 1.5,
		LowerClipStart: 3.6, LowerClipEnd: 4.6,
		MinimumHeight: .075,
	}
}

// CuddlyDNARibbon configures the borrowed front/back text surfaces. Holding
// the first Update preserves the production's draw-then-advance first frame.
func CuddlyDNARibbon(front, back *ebiten.Image) composite.TwistingRibbonConfig {
	return composite.TwistingRibbonConfig{
		Front: front, Back: back, Motion: CuddlyDNATwist(),
		SourceHeight: 25, Filter: ebiten.FilterNearest,
		FirstUpdateHolds: true,
	}
}
