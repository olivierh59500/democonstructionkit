package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// PhenomenaPhotonBounce preserves the source's overshooting photon trajectory.
// Position, floor, gravity, damping and completion threshold are editable.
func PhenomenaPhotonBounce() motion.GravityBounceConfig {
	return motion.GravityBounceConfig{
		StartPosition: 184, StartRebound: -9.50,
		Gravity: .30, Floor: 445, Damping: .70, StopRebound: -.70,
	}
}
