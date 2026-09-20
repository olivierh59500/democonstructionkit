// Package composite renders configurable image instances and exact strip passes.
// Sprites, logos, rasters and scrolling surfaces use the same drawing operations.
package composite

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// Instance represents one sprite, logo or raster fragment. Source is optional;
// nil uses the entire image. Options expose Ebitengine's complete transform,
// filtering, color and blending controls without imposing a center anchor.
type Instance struct {
	Image   *ebiten.Image
	Source  *image.Rectangle
	Options ebiten.DrawImageOptions
}

func (i Instance) Draw(dst *ebiten.Image) {
	if i.Image == nil {
		return
	}
	src := i.Image
	if i.Source != nil {
		r := i.Source.Intersect(src.Bounds())
		if r.Empty() {
			return
		}
		src = src.SubImage(r).(*ebiten.Image)
	}
	dst.DrawImage(src, &i.Options)
}

// Images draws independent instances in caller-specified order. Generate can
// animate arbitrary instance counts, frames, pivots, paths and depth ordering.
type Images struct {
	Instances []Instance
	Generate  func(kit.Frame, func(Instance))
	frame     kit.Frame
}

func (s *Images) Update(f kit.Frame) error { s.frame = f; return nil }
func (s *Images) Draw(dst *ebiten.Image) {
	for _, i := range s.Instances {
		i.Draw(dst)
	}
	if s.Generate != nil {
		s.Generate(s.frame, func(i Instance) { i.Draw(dst) })
	}
}

type Axis uint8

const (
	Rows Axis = iota
	Columns
)

// Strip describes a source crop and an independent destination transform.
// Source is absolute in the input image. DrawImage preserves integer cropping
// before scaling; this differs from interpolating a continuously warped mesh.
type Strip struct {
	Source  image.Rectangle
	Options ebiten.DrawImageOptions
}
type StripMap func(index int, source image.Rectangle, frame kit.Frame) Strip

// Strips draws scanlines/columns of any image: text, logo, sprite or raster asset.
// Thickness, first/count, source selection and destination placement are explicit.
// Mapper can select a different source row or stretch a band to arbitrary height.
type Strips struct {
	Axis                    Axis
	Thickness, First, Count int
	Map                     StripMap
}

func (s Strips) Draw(dst, source *ebiten.Image, frame kit.Frame) {
	if source == nil || s.Thickness <= 0 {
		return
	}
	bounds := source.Bounds()
	extent := bounds.Dy()
	if s.Axis == Columns {
		extent = bounds.Dx()
	}
	count := s.Count
	if count == 0 {
		count = (extent + s.Thickness - 1) / s.Thickness
	}
	for i := s.First; i < s.First+count; i++ {
		r := bounds
		op := ebiten.DrawImageOptions{}
		if s.Axis == Columns {
			r.Min.X = bounds.Min.X + i*s.Thickness
			r.Max.X = min(bounds.Max.X, r.Min.X+s.Thickness)
			op.GeoM.Translate(float64(i*s.Thickness), 0)
		} else {
			r.Min.Y = bounds.Min.Y + i*s.Thickness
			r.Max.Y = min(bounds.Max.Y, r.Min.Y+s.Thickness)
			op.GeoM.Translate(0, float64(i*s.Thickness))
		}
		strip := Strip{Source: r, Options: op}
		if s.Map != nil {
			strip = s.Map(i, r, frame)
		}
		Instance{Image: source, Source: &strip.Source, Options: strip.Options}.Draw(dst)
	}
}

// Pass renders an effect into a reusable surface, then applies an image operation.
// Nest passes to preserve an original row-then-column (or column-then-row) pipeline.
type Pass struct {
	Source    kit.Effect
	Operation func(dst, source *ebiten.Image, frame kit.Frame)
	surface   *ebiten.Image
	frame     kit.Frame
}

func NewPass(source kit.Effect, width, height int, operation func(*ebiten.Image, *ebiten.Image, kit.Frame)) (*Pass, error) {
	if source == nil || width <= 0 || height <= 0 || operation == nil {
		return nil, fmt.Errorf("composite: invalid pass")
	}
	return &Pass{Source: source, Operation: operation, surface: ebiten.NewImageWithOptions(image.Rect(0, 0, width, height), &ebiten.NewImageOptions{Unmanaged: true})}, nil
}
func (p *Pass) Update(f kit.Frame) error { p.frame = f; return p.Source.Update(f) }
func (p *Pass) Draw(dst *ebiten.Image) {
	p.surface.Clear()
	p.Source.Draw(p.surface)
	p.Operation(dst, p.surface, p.frame)
}
func (p *Pass) Close() error { p.surface.Deallocate(); return kit.Close(p.Source) }

// Layer clips/transforms/composites any effect, allowing independent ordering and
// blend modes. Width/height are the effect's native canvas, not a global kit size.
func Layer(source kit.Effect, width, height int, options func(kit.Frame) ebiten.DrawImageOptions) (*Pass, error) {
	return NewPass(source, width, height, func(dst, src *ebiten.Image, f kit.Frame) {
		op := ebiten.DrawImageOptions{}
		if options != nil {
			op = options(f)
		}
		dst.DrawImage(src, &op)
	})
}
