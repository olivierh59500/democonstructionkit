package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// VivaRasterWrapConfig is the editable two-lane title-raster transport shared
// by the standalone Viva screen and its Multiscreen presentation.
func VivaRasterWrapConfig() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0, 72}, Velocity: []float64{-2},
		Lower: &motion.WrapLimit{Boundary: -72, Restart: 72, Inclusive: true},
	}
}
