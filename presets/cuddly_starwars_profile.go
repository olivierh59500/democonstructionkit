package presets

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	motionrecipes "github.com/olivierh59500/democonstructionkit/motion/recipes"
)

// CuddlyStarwarsProfile exposes the pure profile recipe with scene presets.
func CuddlyStarwarsProfile() motion.SegmentedProfileConfig {
	return motionrecipes.CuddlyStarwarsProfile()
}
