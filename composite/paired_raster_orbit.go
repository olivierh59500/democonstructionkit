package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type RasterPair struct{ Left, Right *ebiten.Image }

// PairedRasterOrbitConfig binds phase-selected pairs to borrowed images. The
// phase program controls both the painter order and Y position of each pair.
type PairedRasterOrbitConfig struct {
	Motion         motion.PairedPhaseConfig
	Pairs          []RasterPair
	LeftX, RightX  float64
	ScaleX, ScaleY float64
	Filter         ebiten.Filter
	Blend          ebiten.Blend
}

// PairedRasterOrbit draws each selected pair in exact phase-pass order without
// creating an intermediate surface or allocating geometry during Update/Draw.
type PairedRasterOrbit struct {
	motion *motion.PairedPhaseProgram
	config PairedRasterOrbitConfig
}

func NewPairedRasterOrbit(c PairedRasterOrbitConfig) (*PairedRasterOrbit, error) {
	if len(c.Pairs) == 0 || len(c.Pairs) > 256 ||
		math.IsNaN(c.LeftX) || math.IsInf(c.LeftX, 0) ||
		math.IsNaN(c.RightX) || math.IsInf(c.RightX, 0) ||
		math.IsNaN(c.ScaleX) || math.IsInf(c.ScaleX, 0) || c.ScaleX == 0 ||
		math.IsNaN(c.ScaleY) || math.IsInf(c.ScaleY, 0) || c.ScaleY == 0 {
		return nil, fmt.Errorf("composite: invalid paired raster orbit")
	}
	for _, pair := range c.Pairs {
		if pair.Left == nil || pair.Right == nil {
			return nil, fmt.Errorf("composite: missing raster pair image")
		}
	}
	for _, pass := range c.Motion.Passes {
		if pass.Material >= len(c.Pairs) {
			return nil, fmt.Errorf("composite: raster pair index out of range")
		}
	}
	program, err := motion.NewPairedPhaseProgram(c.Motion)
	if err != nil {
		return nil, err
	}
	c.Pairs = append([]RasterPair(nil), c.Pairs...)
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendSourceOver
	}
	return &PairedRasterOrbit{motion: program, config: c}, nil
}

func (r *PairedRasterOrbit) Step()                  { r.motion.Step() }
func (r *PairedRasterOrbit) Update(kit.Frame) error { r.Step(); return nil }

func (r *PairedRasterOrbit) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil {
		return
	}
	for _, sample := range r.motion.Samples() {
		pair := r.config.Pairs[sample.Material]
		r.drawImage(dst, pair.Left, r.config.LeftX, sample.Y, sample.Alpha)
		r.drawImage(dst, pair.Right, r.config.RightX, sample.Y, sample.Alpha)
	}
}

func (r *PairedRasterOrbit) drawImage(dst, src *ebiten.Image, x, y, alpha float64) {
	op := ebiten.DrawImageOptions{Filter: r.config.Filter, Blend: r.config.Blend}
	op.GeoM.Translate(-float64(src.Bounds().Dx()/2), -float64(src.Bounds().Dy()/2))
	op.GeoM.Scale(r.config.ScaleX, r.config.ScaleY)
	op.GeoM.Rotate(0)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleAlpha(float32(alpha))
	Instance{Image: src, Options: op}.Draw(dst)
}

func (r *PairedRasterOrbit) Motion() *motion.PairedPhaseProgram { return r.motion }
