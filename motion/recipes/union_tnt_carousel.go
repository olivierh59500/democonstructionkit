package recipes

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// UnionTNTCarouselMotion returns the five authored object poses with editable
// entrance and recession thresholds. Geometry and palettes are independent.
func UnionTNTCarouselMotion() motion.ModelCarouselConfig {
	standard := motion.Vector3{X: .02, Y: .02, Z: .02}
	return motion.ModelCarouselConfig{
		Models: []motion.ModelCarouselEntry{
			{RotationStep: standard},
			{Rotation: motion.Vector3{Z: math.Pi}, Camera: motion.Vector3{Z: 100}, RotationStep: standard},
			{RotationStep: standard},
			{Rotation: motion.Vector3{X: -math.Pi / 2}, RotationStep: motion.Vector3{X: .033, Y: .032, Z: .031}},
			{Rotation: motion.Vector3{X: -math.Pi / 2, Z: math.Pi / 3}, Camera: motion.Vector3{Y: 70, Z: 100}, RotationStep: motion.Vector3{Z: .02}},
		},
		Initial: 1, StartCamera: motion.Vector3{Z: 10},
		EntranceStep: 10, EntranceAt: 700, ExitStep: 100, ExitAt: 10000,
	}
}
