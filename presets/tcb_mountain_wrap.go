package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// TCBMountainWrapConfig is the editable two-bank mountain-strip transport used
// by the standalone multiplane screen and its embedded Multiscreen version.
func TCBMountainWrapConfig() motion.WrapBankConfig {
	speeds := [...]float64{8, 7.5, 7, 6.5, 6, 5.5, 5, 4.5, 4, 3.5, 3, 2.5, 2, 1.5, 1, 0.5}
	velocity := make([]float64, 32)
	for index, speed := range speeds {
		velocity[index], velocity[index+16] = -speed, -speed
	}
	return motion.WrapBankConfig{
		Start: make([]float64, 32), Velocity: velocity,
		Lower: &motion.WrapLimit{Boundary: -256, Restart: 256, Inclusive: true, Relative: true},
	}
}
