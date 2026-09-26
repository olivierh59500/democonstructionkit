package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// UnionDeltaLogoFlip keeps the two logo tiles until the lower scale crossing.
func UnionDeltaLogoFlip() motion.BounceToggleConfig {
	return motion.BounceToggleConfig{Start: 1, Velocity: -.02, Min: 0, Max: 1,
		Inclusive: true, ToggleLower: true, Materials: 2}
}

// UnionDeltaGoldWrap is the three-pixel vertical movement of the gold fill.
func UnionDeltaGoldWrap() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-3},
		Lower: &motion.WrapLimit{Boundary: -59, Restart: 0, Inclusive: true},
	}
}
