package presets

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// UnionStarballs composes two sprite materials from one depth population. The
// caller paints the logo and scroll onto the mask before the second material
// is drawn with source-atop blending. Any field, sample or output setting in
// the returned config can be edited before constructing the component.
func UnionStarballs(front, masked *ebiten.Image, randomFloat func() float64,
	paintMask func(*ebiten.Image)) (sprites.MaskedProjectedFieldConfig, error) {
	if front == nil || masked == nil || randomFloat == nil {
		return sprites.MaskedProjectedFieldConfig{}, fmt.Errorf("presets: invalid Union Starballs material or random source")
	}
	sample := func(p sprites.FieldSample, a *sprites.FieldAppearance) bool {
		if p.X < 0 || p.X > 320 || p.Y < 0 || p.Y > 200 {
			return false
		}
		size := (1 - p.Z/32) * 5 / 8
		a.ScaleX, a.ScaleY = size, size
		a.Tint.ScaleAlpha(float32(math.Floor((1-p.Z/32)*255) / 255))
		return true
	}
	output := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver}
	output.GeoM.Scale(2, 2)
	output.GeoM.Translate(64, 68)
	return sprites.MaskedProjectedFieldConfig{
		Width: 320, Height: 200, Unmanaged: true,
		Field: sprites.ProjectedFieldConfig{
			Field: sprites.FieldConfig{Count: 32, Depth: sprites.DepthRespawn, Near: 0, Far: 32,
				Spawn: func(_ int, reset bool) sprites.Point {
					p := sprites.Point{X: math.Floor(randomFloat()*49) - 25,
						Y: math.Floor(randomFloat()*49) - 25, Z: math.Floor(randomFloat()*30) + 1}
					if reset {
						p.Z = 32
					}
					return p
				}},
			View: sprites.FieldView{Camera: geometry.Camera{
				Center: geometry.Vec2{X: 160, Y: 100}, Focal: 64, Near: .001}},
			Velocity: geometry.Vec3{Z: -.2}, Delta: 1, RendererCapacity: 128,
		},
		BaseStyle: sprites.FieldStyle{DrawImages: true, Image: front,
			Blend: ebiten.BlendSourceOver, Sample: sample},
		MaskStyle: sprites.FieldStyle{DrawImages: true, Image: masked,
			Blend: ebiten.BlendSourceAtop, Sample: sample},
		PaintMask: paintMask, BaseOutput: output, MaskOutput: output,
	}, nil
}
