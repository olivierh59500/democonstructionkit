package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// UnionDeltaWordSlide enters, holds and exits once before the main scroller.
func UnionDeltaWordSlide() motion.EnterHoldExitConfig {
	return motion.EnterHoldExitConfig{
		StartX: -640, EnterVelocity: 6, EnterBoundary: 32, EnterInclusive: true,
		ExitVelocity: -6, ExitBoundary: -640, ExitInclusive: true,
		HoldTicks: 200,
	}
}
