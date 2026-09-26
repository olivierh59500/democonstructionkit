package presets

import (
	"github.com/olivierh59500/democonstructionkit/effects"
)

// BilizirCubeTrain configures the twelve independent orange cubes with the
// source's phase spacing, X/Y trajectory and per-index XYZ rotation rates.
// Count, materials, path and all harmonics remain editable before construction.
func BilizirCubeTrain(width float64, count int) effects.SolidCubeTrainConfig {
	motion := BilizirCubeMotion(width, count)
	return effects.SolidCubeTrainConfig{
		Count: motion.Count, Cube: BilizirCube(20),
		PhaseSpacing: motion.PhaseSpacing, PhaseIndexOrigin: motion.PhaseIndexOrigin, PhaseStep: motion.PhaseStep,
		X: motion.X, Y: motion.Y,
		RotationSpacing: motion.RotationSpacing, RotationStep: motion.RotationStep,
		RotationIndexFactor: motion.RotationIndexFactor,
	}
}
