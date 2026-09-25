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

// UnionMenuWalkParallax gives the hall and banner independent movement and
// directional wrap rules while the production retains door navigation.
func UnionMenuWalkParallax(hallLength int) motion.WalkParallaxConfig {
	return motion.WalkParallaxConfig{Layers: []motion.WalkLayer{
		{Start: 0, Speed: 5,
			Lower: &motion.WrapLimit{Boundary: -float64(hallLength), Restart: 0, Inclusive: true},
			Upper: &motion.WrapLimit{Boundary: 0, Restart: -float64(hallLength), Inclusive: true}},
		{Start: 0, Speed: 3,
			Lower: &motion.WrapLimit{Boundary: -94, Restart: -78, Inclusive: true},
			Upper: &motion.WrapLimit{Boundary: -78, Restart: -94, Inclusive: true}},
	}}
}
