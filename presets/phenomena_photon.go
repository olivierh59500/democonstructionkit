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

// PhenomenaPhotonHueCycle keeps the main-screen hue's strict 360-degree wrap.
// The same WrapBank can drive another sprite, raster or logo color cue.
func PhenomenaPhotonHueCycle() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{1.0 / 3.0},
		Upper: &motion.WrapLimit{Boundary: 360, Restart: 0},
	}
}
