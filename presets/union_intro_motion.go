package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// UnionIntroBackdropMotion repeats the 96-pixel source period after 24 ticks.
func UnionIntroBackdropMotion() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{4},
		Upper: &motion.WrapLimit{Boundary: 96, Restart: 0, Inclusive: true},
	}
}

// UnionIntroLogoTransform lets any borrowed image follow the screen's
// horizontal sine while keeping its authored Y and unit scale.
func UnionIntroLogoTransform() motion.HarmonicTransformConfig {
	return motion.HarmonicTransformConfig{
		Base:      motion.Point{X: 256, Y: 43},
		Sin:       motion.Point{X: 190},
		ScaleBase: motion.Point{X: 1, Y: 1},
	}
}

func UnionIntroLogoPhaseStep() float64 { return .05 }
