package sprites

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// BatchedSolidFieldConfig draws a FrameField with one solid quad per particle.
// FrameParticle.Image selects a material. Mask excludes particles whose
// integer-truncated origin lies inside its half-open rectangle; it does not
// clip a quad that crosses the edge. Source is a borrowed white image. When
// nil, the field creates and owns a one-pixel white image instead.
type BatchedSolidFieldConfig struct {
	Motion           motion.FrameFieldConfig
	Materials        []SolidFrame
	Mask             *image.Rectangle
	OffsetX, OffsetY float64
	Source           *ebiten.Image
	Options          *ebiten.DrawTrianglesOptions
}

type solidQuad struct {
	width, height float32
	red, green    float32
	blue, alpha   float32
}

// BatchedSolidField shares FrameField's clock and draws all visible particles
// in a single triangle batch when the population fits the index limit.
type BatchedSolidField struct {
	motion   *motion.FrameField
	quads    []solidQuad
	mask     *image.Rectangle
	offsetX  float64
	offsetY  float64
	source   *ebiten.Image
	owns     bool
	options  *ebiten.DrawTrianglesOptions
	vertices []ebiten.Vertex
	indices  []uint16
}

const solidBatchQuads = (1<<16 - 1) / 4

func NewBatchedSolidField(c BatchedSolidFieldConfig) (*BatchedSolidField, error) {
	if len(c.Materials) == 0 || len(c.Materials) > 1<<16 ||
		math.IsNaN(c.OffsetX) || math.IsInf(c.OffsetX, 0) ||
		math.IsNaN(c.OffsetY) || math.IsInf(c.OffsetY, 0) {
		return nil, fmt.Errorf("sprites: invalid batched solid field")
	}
	field, err := motion.NewFrameField(c.Motion)
	if err != nil {
		return nil, err
	}
	quads := make([]solidQuad, len(c.Materials))
	for i, material := range c.Materials {
		if material.Width < 1 || material.Height < 1 || material.Color == nil {
			return nil, fmt.Errorf("sprites: invalid solid field material %d", i)
		}
		var red, green, blue, alpha uint8
		switch c := material.Color.(type) {
		case color.RGBA:
			red, green, blue, alpha = c.R, c.G, c.B, c.A
		default:
			nrgba := color.NRGBAModel.Convert(material.Color).(color.NRGBA)
			red, green, blue, alpha = nrgba.R, nrgba.G, nrgba.B, nrgba.A
		}
		quads[i] = solidQuad{float32(material.Width), float32(material.Height),
			float32(red) / 0xff, float32(green) / 0xff,
			float32(blue) / 0xff, float32(alpha) / 0xff}
	}
	for i, particle := range field.Samples() {
		if particle.Image < 0 || particle.Image >= len(quads) {
			return nil, fmt.Errorf("sprites: invalid solid field material index %d at particle %d", particle.Image, i)
		}
	}
	batchCount := min(field.Count(), solidBatchQuads)
	f := &BatchedSolidField{
		motion: field, quads: quads, offsetX: c.OffsetX, offsetY: c.OffsetY,
		source: c.Source, vertices: make([]ebiten.Vertex, 0, batchCount*4),
		indices: make([]uint16, 0, batchCount*6),
	}
	if c.Mask != nil {
		mask := *c.Mask
		f.mask = &mask
	}
	if c.Options != nil {
		options := *c.Options
		f.options = &options
	}
	if f.source == nil {
		f.source = ebiten.NewImage(1, 1)
		f.source.Fill(color.White)
		f.owns = true
	}
	return f, nil
}

func (f *BatchedSolidField) Update(kit.Frame) error { return f.motion.Step() }

func (f *BatchedSolidField) Draw(dst *ebiten.Image) {
	if f == nil || dst == nil {
		return
	}
	vertices := f.vertices[:0]
	indices := f.indices[:0]
	for _, particle := range f.motion.Samples() {
		x, y := particle.X+f.offsetX, particle.Y+f.offsetY
		if f.mask != nil && image.Pt(int(x), int(y)).In(*f.mask) {
			continue
		}
		if particle.Image < 0 || particle.Image >= len(f.quads) {
			continue
		}
		if len(vertices)+4 > solidBatchQuads*4 {
			dst.DrawTriangles(vertices, indices, f.source, f.options)
			vertices, indices = vertices[:0], indices[:0]
		}
		quad := f.quads[particle.Image]
		base := uint16(len(vertices))
		x0, y0 := float32(x), float32(y)
		x1, y1 := x0+quad.width, y0+quad.height
		vertices = append(vertices,
			ebiten.Vertex{DstX: x0, DstY: y0, ColorR: quad.red, ColorG: quad.green, ColorB: quad.blue, ColorA: quad.alpha},
			ebiten.Vertex{DstX: x1, DstY: y0, ColorR: quad.red, ColorG: quad.green, ColorB: quad.blue, ColorA: quad.alpha},
			ebiten.Vertex{DstX: x0, DstY: y1, ColorR: quad.red, ColorG: quad.green, ColorB: quad.blue, ColorA: quad.alpha},
			ebiten.Vertex{DstX: x1, DstY: y1, ColorR: quad.red, ColorG: quad.green, ColorB: quad.blue, ColorA: quad.alpha},
		)
		indices = append(indices, base, base+1, base+2, base+1, base+3, base+2)
	}
	f.vertices, f.indices = vertices, indices
	if len(indices) > 0 {
		dst.DrawTriangles(vertices, indices, f.source, f.options)
	}
}

// Motion exposes speed and particle state for live controls.
func (f *BatchedSolidField) Motion() *motion.FrameField { return f.motion }

// Close releases only the internally created white source image.
func (f *BatchedSolidField) Close() {
	if f != nil && f.owns && f.source != nil {
		f.source.Deallocate()
		f.source = nil
	}
}
