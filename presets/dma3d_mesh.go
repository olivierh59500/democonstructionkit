package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// DMA3DMorphMesh returns the three authored fourteen-point forms, twenty-four
// colored faces and exact 120/240-tick morph schedule. All fields may be edited
// before constructing the reusable renderer.
func DMA3DMorphMesh(source *ebiten.Image) effects.MorphingMeshConfig {
	vec := func(x, y, z float64) geometry.Vec3 { return geometry.Vec3{X: x, Y: y, Z: z} }
	shapes := [][]geometry.Vec3{
		{
			vec(0, 0, 200), vec(-100, 25, 100), vec(100, 25, 100), vec(100, -25, 100),
			vec(-100, -25, 100), vec(200, 0, 0), vec(100, 25, -100), vec(100, -25, -100),
			vec(-200, 0, 0), vec(-100, 25, -100), vec(-100, -25, -100), vec(0, 0, -200),
			vec(0, 25, 0), vec(0, -25, 0),
		},
		{
			vec(0, 0, 100), vec(-100, 100, 100), vec(100, 100, 100), vec(100, -100, 100),
			vec(-100, -100, 100), vec(100, 0, 0), vec(100, 100, -100), vec(100, -100, -100),
			vec(-100, 0, 0), vec(-100, 100, -100), vec(-100, -100, -100), vec(0, 0, -100),
			vec(0, 100, 0), vec(0, -100, 0),
		},
		{
			vec(0, 0, 200), vec(-100, 100, 100), vec(100, 100, 100), vec(100, -100, 100),
			vec(-100, -100, 100), vec(200, 0, 0), vec(100, 100, -100), vec(100, -100, -100),
			vec(-200, 0, 0), vec(-100, 100, -100), vec(-100, -100, -100), vec(0, 0, -200),
			vec(0, 200, 0), vec(0, -200, 0),
		},
	}
	face := func(a, b, c, material int) geometry.MorphFace {
		return geometry.MorphFace{Vertices: [3]int{a, b, c}, Material: material}
	}
	faces := []geometry.MorphFace{
		face(0, 2, 1, 0), face(0, 3, 2, 1), face(0, 4, 3, 0), face(0, 1, 4, 1),
		face(5, 6, 2, 1), face(5, 7, 6, 0), face(5, 3, 7, 1), face(5, 2, 3, 0),
		face(8, 1, 9, 1), face(8, 9, 10, 0), face(8, 10, 4, 1), face(8, 4, 1, 0),
		face(11, 6, 7, 1), face(11, 7, 10, 0), face(11, 10, 9, 1), face(11, 9, 6, 0),
		face(12, 1, 2, 1), face(12, 2, 6, 0), face(12, 6, 9, 1), face(12, 9, 1, 0),
		face(13, 7, 3, 0), face(13, 10, 7, 1), face(13, 4, 10, 0), face(13, 3, 4, 1),
	}
	return effects.MorphingMeshConfig{
		Geometry: geometry.MorphingMeshConfig{
			Shapes: shapes, Faces: faces, MorphTicks: 120, CycleTicks: 240,
			RotationStep: geometry.Vec3{X: .01, Y: .02, Z: .04},
			CenterX:      320, CenterY: 240, Focal: 900, CameraZ: 1000,
		},
		Materials: []effects.MorphMaterial{
			{Color: color.RGBA{R: 0x00, G: 0xaa, B: 0x00, A: 0x80}, Blend: ebiten.BlendLighter},
			{Color: color.RGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xe6}, Blend: ebiten.BlendSourceOver},
		},
		CullPositive: true, Filter: ebiten.FilterLinear, Source: source,
	}
}
