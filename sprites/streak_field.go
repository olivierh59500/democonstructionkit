package sprites

import (
	"fmt"
	"image/color"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// StreakField compiles the complete legacy streak motion into a ProjectedField.
// It shares depth transport, projected history and the vector painter with
// other particle materials. NewStreaks remains available for existing callers.
// All StreakConfig values remain editable, including movement off the center.
func StreakField(c StreakConfig) (ProjectedFieldConfig, error) {
	if !dimension(c.Width) || !dimension(c.Height) || !dimension(c.Focal) ||
		c.Count < 0 || c.Count > 100000 || c.Random == nil ||
		!finiteField(c.Speed) || !finiteField(c.CenterX) || !finiteField(c.CenterY) {
		return ProjectedFieldConfig{}, fmt.Errorf("sprites: invalid streak field configuration")
	}
	if c.Color == nil {
		c.Color = color.White
	}
	depth := (c.Width + c.Height) / 2
	dx := float64(int(c.CenterX-c.Width/2) >> 4)
	dy := float64(int(c.CenterY-c.Height/2) >> 4)
	style := FieldStyle{VectorLines: true, VectorTrailPaint: c.Color, Antialias: true}
	style.Sample = func(p FieldSample, a *FieldAppearance) bool {
		if !p.Connected || p.PreviousX <= 0 || p.PreviousX >= c.Width ||
			p.PreviousY <= 0 || p.PreviousY >= c.Height {
			return false
		}
		width := (1 - p.Z/depth) * 2
		if width <= 0 {
			return false
		}
		a.TrailWidth = width
		return true
	}
	return ProjectedFieldConfig{
		Field: FieldConfig{Count: c.Count, Depth: DepthSingleWrap, Near: 0, Far: depth,
			WrapX: c.Width, WrapY: c.Height,
			Spawn: func(_ int, _ bool) Point {
				return Point{X: c.Random()*c.Width*2 - c.Width,
					Y: c.Random()*c.Height*2 - c.Height,
					Z: math.Floor(c.Random()*depth + .5)}
			}},
		View: FieldView{Camera: geometry.Camera{
			Center: geometry.Vec2{X: c.Width / 2, Y: c.Height / 2},
			Focal:  c.Focal, Near: math.SmallestNonzeroFloat64,
		}, DivideFirst: true},
		Velocity: geometry.Vec3{X: dx, Y: dy, Z: -c.Speed}, Delta: 1,
		Style: style, RendererCapacity: c.Count, ColdStart: true,
	}, nil
}
