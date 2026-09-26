package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// CuddlyLEDTileOffset scrolls the cached backdrop by two pixels and applies
// the exact inclusive 33-pixel restart, independently of the gradient clock.
func CuddlyLEDTileOffset() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-2},
		Lower: &motion.WrapLimit{Boundary: -33, Restart: 0, Inclusive: true},
	}
}
