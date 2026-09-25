package presets

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// UnionMenuPanoramaWrap is the repeating two-pixel panorama strip.
func UnionMenuPanoramaWrap() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-2},
		Lower: &motion.WrapLimit{Boundary: -540, Restart: -28, Inclusive: true},
	}
}

// UnionMenuCoverWidth is the 840-pixel uncover wipe, shrinking five pixels
// per logical tick and holding at zero.
func UnionMenuCoverWidth() motion.LinearTickConfig {
	return motion.LinearTickConfig{Start: 840, Step: -5, Min: 0, Max: 840}
}

// UnionMenuPaletteCycle retains the source's first update at tick four and
// later changes every three ticks, independently of the palette's colors.
func UnionMenuPaletteCycle(colors int) timeline.PacedIndexConfig {
	return timeline.PacedIndexConfig{Count: colors, First: 4, Every: 3}
}

// UnionMenuCharacterCycle advances one character image per five movements.
func UnionMenuCharacterCycle() timeline.PacedIndexConfig {
	return timeline.PacedIndexConfig{Count: 8, First: 5, Every: 5}
}

// UnionMenuLogoBounce keeps a long hold followed by a bounded vertical logo
// squeeze and rebound. The first movement begins with five wait ticks left.
func UnionMenuLogoBounce() motion.HoldBounceConfig {
	return motion.HoldBounceConfig{Min: 0, Max: 1, Velocity: -.04, HoldTicks: 1000, LeadTicks: 5}
}
