package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

const bilizirLogoPhaseStep = .05

// BilizirLogoClock is the shared phase source for plain and warped artwork.
func BilizirLogoClock() motion.WaveClockConfig {
	return motion.WaveClockConfig{Step: bilizirLogoPhaseStep}
}

func BilizirLogoPhaseStep() float64 { return bilizirLogoPhaseStep }

// BilizirPlainLogoMotion keeps the complete horizontal sine amplitude.
func BilizirPlainLogoMotion(screenWidth, logoWidth float64) motion.HarmonicTransformConfig {
	center := (screenWidth - logoWidth) / 2
	return motion.HarmonicTransformConfig{
		Base: motion.Point{X: center}, Sin: motion.Point{X: center},
		ScaleBase: motion.Point{X: 1, Y: 1},
	}
}

// BilizirWarpedLogoMotion reduces travel by its source padding and clamps an
// oversized logo to the center instead of allowing the warped edge to escape.
func BilizirWarpedLogoMotion(screenWidth, logoWidth, margin float64) motion.HarmonicTransformConfig {
	center := (screenWidth - logoWidth) / 2
	amplitude := math.Max(0, center-margin)
	return motion.HarmonicTransformConfig{
		Base: motion.Point{X: center}, Sin: motion.Point{X: amplitude},
		ScaleBase: motion.Point{X: 1, Y: 1},
	}
}
