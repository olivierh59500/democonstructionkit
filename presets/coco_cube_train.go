package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CocoCubeTrain returns editable cube count, path, spacing, speed, palette and
// per-index rotation rates for Coco's orange cube procession.
func CocoCubeTrain(width, height, cubeSize float64, count int) effects.SolidCubeTrainConfig {
	horizontalRadius := (width - cubeSize) / 2
	return effects.SolidCubeTrainConfig{
		Count: count, Cube: CocoCube(cubeSize),
		PhaseSpacing: .15, PhaseIndexOrigin: 1, PhaseStep: .04,
		X:                   motion.Wave{Amplitude: horizontalRadius, Offset: horizontalRadius, Speed: 1},
		Y:                   motion.Wave{Amplitude: 84, Offset: height / 2, Speed: 2.5, Cos: true},
		RotationSpacing:     geometry.Vec3{X: .3, Y: .2, Z: .1},
		RotationStep:        geometry.Vec3{X: .02, Y: .03, Z: .01},
		RotationIndexFactor: geometry.Vec3{X: .1, Y: .15, Z: .05},
	}
}

// MultiscreenCocoCubeTrain keeps the same path with the embedded screen's
// material and bounded sine/cosine recurrence for long-running playback.
func MultiscreenCocoCubeTrain(width, height, cubeSize float64, count int) effects.SolidCubeTrainConfig {
	c := CocoCubeTrain(width, height, cubeSize, count)
	c.Cube = MultiscreenCocoCube(cubeSize)
	c.RecurrenceInterval = 1024
	c.RecurrencePeriod = 4 * math.Pi
	return c
}
