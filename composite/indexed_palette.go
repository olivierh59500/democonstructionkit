package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// IndexedPaletteConfig maps a stored image component to a live palette.
// Channel defaults to alpha; red is customary for opaque encoded indices.
// Scale zero defaults to len(Palette)-1. Offset is added before nearest-integer
// rounding; indices clamp at each end. SourceAlpha optionally multiplies the
// palette result by source alpha. Palette alpha remains independently editable.
type IndexedPaletteConfig struct {
	Palette       []color.NRGBA
	Channel       BitplaneChannel
	Scale, Offset float32
	SourceAlpha   bool
	Blend         ebiten.Blend
}

// IndexedPalette owns only a shader and reusable color uniforms. Source images
// stay borrowed, including cropped/atlased inputs. SetPalette changes colors
// without decoding or reuploading the indexed artwork.
type IndexedPalette struct {
	count   int
	shader  *ebiten.Shader
	colors  []float32
	options ebiten.DrawRectShaderOptions
}

func NewIndexedPalette(c IndexedPaletteConfig) (*IndexedPalette, error) {
	if len(c.Palette) < 1 || len(c.Palette) > 256 || c.Channel > BitplaneBlue ||
		math.IsNaN(float64(c.Scale)) || math.IsInf(float64(c.Scale), 0) || c.Scale < 0 ||
		math.IsNaN(float64(c.Offset)) || math.IsInf(float64(c.Offset), 0) {
		return nil, fmt.Errorf("composite: invalid indexed palette or encoding")
	}
	if c.Scale == 0 {
		c.Scale = float32(len(c.Palette) - 1)
	}
	shader, err := ebiten.NewShader([]byte(fmt.Sprintf(indexedPaletteShader, len(c.Palette))))
	if err != nil {
		return nil, fmt.Errorf("composite: compile indexed palette: %w", err)
	}
	p := &IndexedPalette{count: len(c.Palette), shader: shader, colors: make([]float32, len(c.Palette)*4)}
	channel := make([]float32, 4)
	component := 3
	if c.Channel != BitplaneAlpha {
		component = int(c.Channel) - 1
	}
	channel[component] = 1
	sourceAlpha := float32(0)
	if c.SourceAlpha {
		sourceAlpha = 1
	}
	p.options.Blend = c.Blend
	p.options.Uniforms = map[string]any{"Palette": p.colors, "Channel": channel,
		"Encoding": []float32{c.Scale, c.Offset, float32(p.count - 1), sourceAlpha}}
	if err := p.SetPalette(c.Palette); err != nil {
		p.Close()
		return nil, err
	}
	return p, nil
}

func (p *IndexedPalette) SetPalette(colors []color.NRGBA) error {
	if p == nil || p.shader == nil || len(colors) != p.count {
		return fmt.Errorf("composite: invalid or closed indexed color bank")
	}
	for i, c := range colors {
		a := float32(c.A) / 255
		p.colors[i*4], p.colors[i*4+1], p.colors[i*4+2], p.colors[i*4+3] = float32(c.R)/255*a, float32(c.G)/255*a, float32(c.B)/255*a, a
	}
	return nil
}

func (p *IndexedPalette) Draw(dst, source *ebiten.Image) error {
	if p == nil {
		return fmt.Errorf("composite: invalid or closed indexed palette draw")
	}
	return p.DrawWith(dst, source, p.options)
}

// DrawWith places and tints the indexed image in one shader draw. GeoM,
// ColorScale and blending are copied for this call; the lookup owns Images and
// Uniforms, so those caller fields are ignored. A later Draw retains its original
// identity placement and configured blend rather than inheriting these options.
func (p *IndexedPalette) DrawWith(dst, source *ebiten.Image, options ebiten.DrawRectShaderOptions) error {
	if p == nil || p.shader == nil || dst == nil || source == nil || dst == source {
		return fmt.Errorf("composite: invalid or closed indexed palette draw")
	}
	options.Images = [4]*ebiten.Image{source}
	options.Uniforms = p.options.Uniforms
	dst.DrawRectShader(source.Bounds().Dx(), source.Bounds().Dy(), p.shader, &options)
	return nil
}

// Close leaves all borrowed indexed images alive and is safe to repeat.
func (p *IndexedPalette) Close() error {
	if p != nil {
		if p.shader != nil {
			p.shader.Deallocate()
			p.shader = nil
		}
		p.options.Images = [4]*ebiten.Image{}
		p.options.Uniforms = nil
		p.colors = nil
	}
	return nil
}

const indexedPaletteShader = `//kage:unit pixels
package main

var Palette [%d]vec4
var Channel vec4
var Encoding vec4

func Fragment(position vec4, source vec2, color vec4) vec4 {
	c := imageSrc0At(source)
	index := int(clamp(floor(dot(c, Channel)*Encoding.x+Encoding.y+0.5), 0, Encoding.z))
	result := Palette[index]
	if Encoding.w > 0 { result *= c.a }
	return result * color
}
`
