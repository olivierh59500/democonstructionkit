package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlySpreadpointLogoPhases retains the authored hold, reverse and forward
// sections of the 2,272-frame logo cycle. Every reset is an explicit value.
func CuddlySpreadpointLogoPhases() motion.PhaseSequenceConfig {
	pi, zero := math.Pi, 0.0
	return motion.PhaseSequenceConfig{
		Start: math.Pi, Period: 2 * math.Pi,
		Segments: []motion.PhaseSegment{
			{Samples: 300, Reset: &pi},
			{Samples: 40, Step: -math.Pi / 40},
			{Samples: 300, Reset: &zero},
			{Samples: 640, Step: -math.Pi / 40},
			{Samples: 480, Step: -math.Pi / 32},
			{Samples: 512, Step: math.Pi / 32},
		},
	}
}

// CuddlySpreadpointLogoTransform ties the horizontal position to zoom while
// the vertical position follows sine. Another image can use the same motion.
func CuddlySpreadpointLogoTransform() motion.HarmonicTransformConfig {
	return motion.HarmonicTransformConfig{
		Base:          motion.Point{X: 64, Y: 45},
		Sin:           motion.Point{Y: 40},
		ScaleCoupling: motion.Point{X: -64},
		ScaleBase:     motion.Point{X: .75, Y: .75},
		ScaleCos:      motion.Point{X: .25, Y: .25},
	}
}
