package presets

import (
	"image/color"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyDNADiscCloud keeps the screen's Y-axis rotation, perspective and
// orange batched-disc material. The supplied model points are copied by DCK.
func CuddlyDNADiscCloud(points []geometry.Vec3) sprites.RotatingDiscCloudConfig {
	return sprites.RotatingDiscCloudConfig{
		Motion: geometry.RotatingDiscCloudConfig{
			Points: points, CenterX: 208, CenterY: 138, DepthBase: 900,
			YOffset: 16, Focal: 138 / math.Tan(20*math.Pi/180),
			AngleStep: .03, RadiusScale: 1,
		},
		Tint: color.RGBA{R: 238, G: 136, A: 255}, Antialias: true,
	}
}
