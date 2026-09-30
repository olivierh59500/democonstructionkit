package presets

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// VectorballsProjectedObject supplies the demonstration's camera and angular
// steps to any of the five editable point-object families. All dimensions,
// fill, ball index, wave and projection fields may be changed before creation.
func VectorballsProjectedObject(name string, fill sprites.Fill, segments int, size float64, image int) (sprites.ProjectedObjectConfig, error) {
	if image == -1 {
		image = 120
		if name == "sphere" {
			image = 10
			if segments > 16 {
				image = 8
			}
		}
	}
	if image < 0 || image > 120 {
		return sprites.ProjectedObjectConfig{}, fmt.Errorf("presets: vectorball index must be within [0,120]")
	}
	c := sprites.ProjectedObjectConfig{
		RotationStep: geometry.Vec3{X: .012, Y: .017, Z: .005},
		Position:     geometry.Vec3{Z: 850}, Scale: .55,
		Focal: 1450, CenterX: 320, CenterY: 193,
		// With Focal/(Focal+Z), larger Z is farther away. Submit it first so
		// foreground balls cover background balls at projected crossings.
		YUp: true, AscendingDepth: false,
	}
	switch name {
	case "cube":
		c.Cube = &sprites.CubeConfig{Size: size, Segments: segments, Fill: fill, Image: image}
	case "pyramid":
		c.Pyramid = &sprites.PyramidConfig{Width: size, Height: size, Segments: segments, Fill: fill, Image: image}
	case "plane", "flag":
		c.Plane = &sprites.PlaneConfig{Width: size, Height: size * .65, Columns: segments, Rows: segments, Image: image}
		if name == "flag" {
			c.Flag = &sprites.Flag{Width: size, PinLeft: true,
				Wave: motion.Wave{Amplitude: size * .15, Spatial: 7 / size, Speed: 3}, RowPhase: 2 / size}
		}
	case "sphere":
		if segments < 1 || segments > 32 {
			return sprites.ProjectedObjectConfig{}, fmt.Errorf("presets: sphere density must be within [1,32]")
		}
		c.Sphere = &sprites.SphereConfig{Radius: size / 2, Count: segments * segments * 4, Image: image}
		c.Batch = true
		if segments >= 12 {
			c.CullPositiveModelZ = true
			c.ScaleImages = true
		}
	default:
		return sprites.ProjectedObjectConfig{}, fmt.Errorf("presets: unknown vectorball object %q", name)
	}
	return c, nil
}
