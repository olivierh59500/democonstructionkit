package presets

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyStarwarsStars keeps the seeded 400-point field, camera travel and
// three authored depth shades editable without a screen-local update loop.
func CuddlyStarwarsStars(randomFloat func() float64) (sprites.ProjectedFieldConfig, error) {
	if randomFloat == nil {
		return sprites.ProjectedFieldConfig{}, fmt.Errorf("presets: missing Starwars random source")
	}
	return sprites.ProjectedFieldConfig{
		Field: sprites.FieldConfig{Count: 400, Depth: sprites.DepthWrap, Near: 0, Far: 130,
			Spawn: func(i int, _ bool) sprites.Point {
				return sprites.Point{X: math.Floor(randomFloat()*320) - 160,
					Y: math.Floor(randomFloat()*200) - 100, Z: float64(i) * (130.0 / 400)}
			}},
		View: sprites.FieldView{Camera: geometry.Camera{
			Center: geometry.Vec2{X: 160, Y: 100}, Focal: 128,
			Near: math.SmallestNonzeroFloat64}, Angle: 180},
		ViewVelocity: geometry.Vec3{Z: -1.5}, AngleStep: -.02,
		RendererCapacity: 400,
		Style: sprites.FieldStyle{Antialias: true, Sample: func(p sprites.FieldSample,
			a *sprites.FieldAppearance) bool {
			shade := float32(1)
			if p.Z > 130.0/3 {
				shade = float32(170*257) / 65535
			}
			if p.Z > 260.0/3 {
				shade = float32(85*257) / 65535
			}
			a.Tint.Scale(shade, shade, shade, 1)
			return true
		}},
	}, nil
}
