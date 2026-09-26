package presets

import (
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// BilizirCubeMotionConfig is the image-independent part of the cube train.
// A renderer can borrow its exact path and rotation rates with other artwork.
type BilizirCubeMotionConfig struct {
	Count                                     int
	PhaseSpacing, PhaseIndexOrigin, PhaseStep float64
	X, Y                                      motion.Wave
	RotationSpacing, RotationStep             geometry.Vec3
	RotationIndexFactor                       geometry.Vec3
}

func BilizirCubeMotion(width float64, count int) BilizirCubeMotionConfig {
	radiusX := (width - 40) / 2
	return BilizirCubeMotionConfig{
		Count: count, PhaseSpacing: .15, PhaseIndexOrigin: 1, PhaseStep: .04,
		X:                   motion.Wave{Amplitude: radiusX, Offset: radiusX, Speed: 1},
		Y:                   motion.Wave{Amplitude: 84, Offset: 186, Speed: 2.5, Cos: true},
		RotationSpacing:     geometry.Vec3{X: .3, Y: .5, Z: .2},
		RotationStep:        geometry.Vec3{X: .02, Y: .03, Z: .01},
		RotationIndexFactor: geometry.Vec3{X: .1, Y: .15, Z: .05},
	}
}
