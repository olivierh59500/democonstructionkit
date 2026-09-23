package presets

import (
	"fmt"
	"image/color"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// NonamenoStarsConfig describes a radial projected field with independent
// initial and respawn patterns. All geometry, depth shading and trail controls
// are plain values suitable for an editor; the resulting Go callbacks are an
// implementation detail of the compiled field style.
type NonamenoStarsConfig struct {
	Count                                   int
	Speed, Width, Height, CenterX, CenterY  float64
	DepthMax, Focal, InitialTurns           float64
	SizeBase, SizeDepthLoss, MinSize        float64
	BrightnessDepthLoss                     float64
	BaseColor, TrailColor                   color.RGBA
	TrailDepthFraction, TrailWidth          float64
	InitialRadiusStride, InitialDepthStride int
	ResetAngleStride, ResetRadiusStride     int
}

// DefaultNonamenoStarsConfig reproduces the 500-point native scene.
func DefaultNonamenoStarsConfig() NonamenoStarsConfig {
	return NonamenoStarsConfig{
		Count: 500, Speed: 2, Width: 640, Height: 480, CenterX: 320, CenterY: 240,
		DepthMax: 100, Focal: 200, InitialTurns: 1,
		SizeBase: 3, SizeDepthLoss: 2.5, MinSize: .5, BrightnessDepthLoss: .7,
		BaseColor: color.RGBA{255, 255, 255, 255}, TrailColor: color.RGBA{128, 128, 128, 128},
		TrailDepthFraction: .3, TrailWidth: 1,
		InitialRadiusStride: 137, InitialDepthStride: 17,
		ResetAngleStride: 137, ResetRadiusStride: 89,
	}
}

func (c NonamenoStarsConfig) Validate() error {
	if c.Count < 0 || c.Count > 1_000_000 || c.Width <= 0 || c.Height <= 0 ||
		c.DepthMax <= 0 || c.Focal <= 0 || c.MinSize <= 0 || c.TrailWidth < 0 ||
		c.TrailDepthFraction < 0 || c.TrailDepthFraction > 1 {
		return fmt.Errorf("presets: invalid radial star field count, geometry or trail")
	}
	for _, v := range []float64{c.Speed, c.Width, c.Height, c.CenterX, c.CenterY,
		c.DepthMax, c.Focal, c.InitialTurns, c.SizeBase, c.SizeDepthLoss,
		c.MinSize, c.BrightnessDepthLoss, c.TrailDepthFraction, c.TrailWidth} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("presets: non-finite radial star field parameter")
		}
	}
	return nil
}

// NonamenoProjectedStars compiles the editable radial pattern into the shared
// projected field. Source order, reset depth and rectangle-then-trail painter
// order remain exact even when size, colors, count or speed are changed.
func NonamenoProjectedStars(c NonamenoStarsConfig) (sprites.ProjectedFieldConfig, error) {
	if err := c.Validate(); err != nil {
		return sprites.ProjectedFieldConfig{}, err
	}
	initial := make([]sprites.Point, c.Count)
	reset := make([]sprites.Point, c.Count)
	for i := range initial {
		angle := float64(i) * 2.0 * math.Pi / float64(c.Count)
		angle *= c.InitialTurns
		radius := math.Mod(float64(i*c.InitialRadiusStride), float64(c.Count)) / float64(c.Count) * c.Width / 2
		resetAngle := math.Mod(float64(i*c.ResetAngleStride), 360.0) * math.Pi / 180.0
		resetRadius := math.Mod(float64(i*c.ResetRadiusStride), c.Width/2)
		z := math.Mod(float64(i*c.InitialDepthStride), c.DepthMax)
		if z < 1 {
			z = 1
		}
		initial[i] = sprites.Point{X: math.Cos(angle) * radius, Y: math.Sin(angle) * radius, Z: z}
		reset[i] = sprites.Point{X: math.Cos(resetAngle) * resetRadius, Y: math.Sin(resetAngle) * resetRadius, Z: c.DepthMax}
	}
	style := sprites.FieldStyle{VectorRects: true}
	style.Sample = func(p sprites.FieldSample, a *sprites.FieldAppearance) bool {
		if p.X < 0 || p.X >= c.Width || p.Y < 0 || p.Y >= c.Height {
			return false
		}
		size := c.SizeBase - p.Z/c.DepthMax*c.SizeDepthLoss
		if size < c.MinSize {
			size = c.MinSize
		}
		brightness := 1.0 - p.Z/c.DepthMax*c.BrightnessDepthLoss
		a.Width, a.Height = size, size
		a.FillColor = color.RGBA{R: uint8(float64(c.BaseColor.R) * brightness), G: uint8(float64(c.BaseColor.G) * brightness), B: uint8(float64(c.BaseColor.B) * brightness), A: c.BaseColor.A}
		if p.Z < c.DepthMax*c.TrailDepthFraction && p.PreviousX != 0 && p.PreviousY != 0 {
			a.TrailWidth = c.TrailWidth
			a.TrailColor = color.RGBA{R: uint8(float64(c.TrailColor.R) * brightness), G: uint8(float64(c.TrailColor.G) * brightness), B: uint8(float64(c.TrailColor.B) * brightness), A: c.TrailColor.A}
		}
		return true
	}
	return sprites.ProjectedFieldConfig{
		Field: sprites.FieldConfig{Points: initial, Spawn: func(i int, _ bool) sprites.Point { return reset[i] },
			Depth: sprites.DepthRespawn, Near: 0, Far: c.DepthMax},
		View:     sprites.FieldView{Camera: geometry.Camera{Center: geometry.Vec2{X: c.CenterX, Y: c.CenterY}, Focal: c.Focal, Near: math.SmallestNonzeroFloat64}},
		Velocity: geometry.Vec3{Z: -c.Speed}, Delta: 1, Style: style,
	}, nil
}
