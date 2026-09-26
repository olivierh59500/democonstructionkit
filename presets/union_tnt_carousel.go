package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	motionrecipes "github.com/olivierh59500/democonstructionkit/motion/recipes"
)

// UnionTNTCarouselMotion is the editable five-object entry and handoff clock.
func UnionTNTCarouselMotion() motion.ModelCarouselConfig {
	return motionrecipes.UnionTNTCarouselMotion()
}

// UnionTNTMeshCarousel binds caller-owned model data to the original camera,
// Euler order, mirrored axes and per-face materials.
func UnionTNTMeshCarousel(models []effects.SolidMeshModel) effects.SolidMeshCarouselConfig {
	return effects.SolidMeshCarouselConfig{
		Models: models, Motion: UnionTNTCarouselMotion(),
		Camera: geometry.Camera{Center: geometry.Vec2{X: 320, Y: 200},
			Focal: 200 / math.Tan(25*math.Pi/360), Near: 1},
		RotationOrder: [3]uint8{2, 1, 0},
		Mirror:        geometry.Vec3{X: 1, Y: -1, Z: -1},
		CameraSign:    geometry.Vec3{X: -1, Y: 1, Z: 1},
		CullBackFaces: true,
	}
}
