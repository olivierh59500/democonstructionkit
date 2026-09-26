package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// VectorballsSinusGrid retains the 8×8 point wave and its two phase clocks.
func VectorballsSinusGrid() geometry.SinusGridConfig {
	return geometry.SinusGridConfig{
		Columns: 8, Rows: 8, Amplitude: 200,
		SpatialNumerator: math.Pi, SpatialDivisor: 10,
		TravelStep: math.Pi / 55, EnvelopeStep: math.Pi / 60,
	}
}

// VectorballsRotors moves four paired helicopter rotor points in model space.
func VectorballsRotors() geometry.RotorConfig {
	return geometry.RotorConfig{Radii: []float64{80, 150, 210, 160}, Offset: 100, PhaseStep: .08, RequireFull: true}
}

// VectorballsYOrbit overrides the model's linear position during its entrance.
func VectorballsYOrbit() geometry.YOrbitConfig {
	return geometry.YOrbitConfig{Center: geometry.Vec3{Z: 850}, Radius: 100,
		PhaseStep: math.Pi / 144, RatioStep: .025, Min: 0, Max: 1}
}

// VectorballsBounce reproduces the decaying three-cycle Y curve and final
// forward sweep. The sample table is built once, then stepped without trig.
func VectorballsBounce() geometry.BounceCurveConfig {
	const conversion = 2.4
	return geometry.BounceCurveConfig{
		Radius: 1200, Offset: 210,
		RadiusDecay: 5.7143 / conversion, OffsetDecay: 1 / conversion,
		StepDegrees: 3 / conversion, BackwardStart: 89,
		FinalStepDegrees: 1, Loops: 3, Divisor: 3,
	}
}
