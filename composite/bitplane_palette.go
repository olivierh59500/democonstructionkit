package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// BitplaneChannel selects the stored color component that supplies a plane bit.
// Alpha is the default for silhouette masks; red suits opaque monochrome art.
type BitplaneChannel uint8

const (
	BitplaneAlpha BitplaneChannel = iota
	BitplaneRed
	BitplaneGreen
	BitplaneBlue
)

// BitplanePaletteConfig describes a palette lookup over one to six binary image
// planes. Plane zero is the least significant bit. Each source must have the
// configured dimensions and a zero origin. Channels defaults to alpha for every
// plane; when supplied, it must contain one selection per plane. A zero
// threshold selects 0.5. Palette entries may be translucent. Color channels
// sample the source image's premultiplied values, before palette lookup.
type BitplanePaletteConfig struct {
	Width, Height int
	Planes        int
	Threshold     float32
	Palette       []color.NRGBA
	Channels      []BitplaneChannel
	Blend         ebiten.Blend
}

// BitplanePalette combines caller-owned mask images without reading pixels back
// from the GPU. One to four planes use one draw pass. Five or six planes use a
// retained RGBA packing surface and a second pass to stay within Ebitengine's
// four-source limit. The palette can change independently of the mask images.
// Draw and SetPalette belong on the graphics goroutine.
type BitplanePalette struct {
	width, height int
	planes        int
	threshold     float32
	palette       []float32
	channels      []float32
	offsets       []float32
	shader        *ebiten.Shader
	packShader    *ebiten.Shader
	packed        *ebiten.Image
	options       ebiten.DrawRectShaderOptions
	packOptions   ebiten.DrawRectShaderOptions
}

func NewBitplanePalette(c BitplanePaletteConfig) (*BitplanePalette, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 8192 || c.Height > 8192 ||
		int64(c.Width)*int64(c.Height) > 16*1024*1024 || c.Planes < 1 || c.Planes > 6 ||
		len(c.Palette) != 1<<c.Planes {
		return nil, fmt.Errorf("composite: invalid bitplane dimensions, count or palette")
	}
	if c.Threshold == 0 {
		c.Threshold = 0.5
	}
	if math.IsNaN(float64(c.Threshold)) || c.Threshold <= 0 || c.Threshold > 1 {
		return nil, fmt.Errorf("composite: invalid bitplane threshold")
	}
	if len(c.Channels) != 0 && len(c.Channels) != c.Planes {
		return nil, fmt.Errorf("composite: expected one channel per bitplane")
	}
	for _, channel := range c.Channels {
		if channel > BitplaneBlue {
			return nil, fmt.Errorf("composite: unknown bitplane channel %d", channel)
		}
	}
	b := &BitplanePalette{width: c.Width, height: c.Height, planes: c.Planes,
		threshold: c.Threshold, palette: make([]float32, 64*4), channels: make([]float32, 6*4), offsets: make([]float32, 6*2)}
	for i := 0; i < 6; i++ {
		channel := BitplaneAlpha
		if i < len(c.Channels) {
			channel = c.Channels[i]
		}
		component := 3
		if channel != BitplaneAlpha {
			component = int(channel) - 1
		}
		b.channels[4*i+component] = 1
	}
	var err error
	if c.Planes <= 4 {
		b.shader, err = ebiten.NewShader([]byte(bitplaneDirectShader))
	} else {
		b.packShader, err = ebiten.NewShader([]byte(bitplanePackShader))
		if err == nil {
			b.shader, err = ebiten.NewShader([]byte(bitplanePackedShader))
		}
	}
	if err != nil {
		b.Close()
		return nil, fmt.Errorf("composite: compile bitplane shader: %w", err)
	}
	if c.Planes > 4 {
		b.packed = render.NewSurface(c.Width, c.Height)
		b.packOptions.Blend = ebiten.BlendCopy
		b.packOptions.Uniforms = map[string]any{"Threshold": b.threshold, "Channels": b.channels, "Offsets": b.offsets}
	}
	b.options.Blend = c.Blend
	b.options.Uniforms = map[string]any{"Palette": b.palette, "Threshold": b.threshold, "Channels": b.channels, "Offsets": b.offsets}
	if err := b.SetPalette(c.Palette); err != nil {
		b.Close()
		return nil, err
	}
	return b, nil
}

// SetPalette replaces the current color bank without reallocating GPU images.
// Input colors are copied into a premultiplied shader uniform.
func (b *BitplanePalette) SetPalette(colors []color.NRGBA) error {
	if b == nil || b.shader == nil {
		return fmt.Errorf("composite: bitplane palette is closed")
	}
	if len(colors) != 1<<b.planes {
		return fmt.Errorf("composite: expected %d bitplane colors, got %d", 1<<b.planes, len(colors))
	}
	for i, c := range colors {
		a := float32(c.A) / 255
		b.palette[4*i], b.palette[4*i+1], b.palette[4*i+2], b.palette[4*i+3] =
			float32(c.R)/255*a, float32(c.G)/255*a, float32(c.B)/255*a, a
	}
	return nil
}

// Draw composes the current source bank at the destination origin. The planes
// slice must contain exactly the configured number of non-nil images. Borrowed
// masks may be changed between draws, including masks from a retained ring.
func (b *BitplanePalette) Draw(dst *ebiten.Image, planes []*ebiten.Image) error {
	return b.DrawOffsets(dst, planes, nil)
}

// DrawOffsets samples each plane at an independent pixel offset before palette
// lookup. A positive X/Y reads pixels to the right/below, moving the visible
// material left/up. Nil offsets select zero for every plane; otherwise supply
// one finite {X, Y} pair per plane. The same borrowed image may supply several
// planes with different offsets or channels. Sampling outside a source is
// transparent; it does not wrap. Draw resets all offsets to zero on its next
// call. The existing one/two-pass path and image budget are unchanged.
func (b *BitplanePalette) DrawOffsets(dst *ebiten.Image, planes []*ebiten.Image, offsets [][2]float32) error {
	if b == nil || b.shader == nil || dst == nil {
		return fmt.Errorf("composite: bitplane target is nil or closed")
	}
	if len(planes) != b.planes {
		return fmt.Errorf("composite: expected %d bitplane images, got %d", b.planes, len(planes))
	}
	if len(offsets) != 0 && len(offsets) != b.planes {
		return fmt.Errorf("composite: expected %d bitplane offsets, got %d", b.planes, len(offsets))
	}
	for _, offset := range offsets {
		for _, component := range offset {
			if math.IsNaN(float64(component)) || math.IsInf(float64(component), 0) {
				return fmt.Errorf("composite: bitplane offset must be finite")
			}
		}
	}
	for _, plane := range planes {
		if plane == nil || plane.Bounds().Min.X != 0 || plane.Bounds().Min.Y != 0 ||
			plane.Bounds().Dx() != b.width || plane.Bounds().Dy() != b.height {
			return fmt.Errorf("composite: bitplane image bounds must be (0,0)-(%d,%d)", b.width, b.height)
		}
	}
	clear(b.offsets)
	for i, offset := range offsets {
		b.offsets[2*i], b.offsets[2*i+1] = offset[0], offset[1]
	}
	if b.packed == nil {
		b.options.Images = [4]*ebiten.Image{}
		copy(b.options.Images[:], planes)
	} else {
		b.packOptions.Images = [4]*ebiten.Image{}
		copy(b.packOptions.Images[:], planes[:4])
		b.packed.DrawRectShader(b.width, b.height, b.packShader, &b.packOptions)
		b.options.Images = [4]*ebiten.Image{b.packed, planes[4]}
		if b.planes == 6 {
			b.options.Images[2] = planes[5]
		}
	}
	dst.DrawRectShader(b.width, b.height, b.shader, &b.options)
	return nil
}

// Close releases owned shaders and the optional packing surface. It leaves all
// caller-owned masks untouched and may be called more than once.
func (b *BitplanePalette) Close() error {
	if b == nil {
		return nil
	}
	if b.packed != nil {
		b.packed.Deallocate()
		b.packed = nil
	}
	if b.packShader != nil {
		b.packShader.Deallocate()
		b.packShader = nil
	}
	if b.shader != nil {
		b.shader.Deallocate()
		b.shader = nil
	}
	b.options.Images = [4]*ebiten.Image{}
	b.packOptions.Images = [4]*ebiten.Image{}
	b.options.Uniforms = nil
	b.packOptions.Uniforms = nil
	b.palette = nil
	b.channels = nil
	b.offsets = nil
	return nil
}

const bitplaneDirectShader = `//kage:unit pixels
package main

var Threshold float
var Palette [64]vec4
var Channels [6]vec4
var Offsets [6]vec2

func Fragment(position vec4, source vec2, color vec4) vec4 {
	index := int(step(Threshold, dot(imageSrc0At(source + Offsets[0]), Channels[0])) +
		2*step(Threshold, dot(imageSrc1At(source + Offsets[1]), Channels[1])) +
		4*step(Threshold, dot(imageSrc2At(source + Offsets[2]), Channels[2])) +
		8*step(Threshold, dot(imageSrc3At(source + Offsets[3]), Channels[3])) + 0.5)
	return Palette[index]
}
`

const bitplanePackShader = `//kage:unit pixels
package main

var Threshold float
var Channels [6]vec4
var Offsets [6]vec2

func Fragment(position vec4, source vec2, color vec4) vec4 {
	return vec4(step(Threshold, dot(imageSrc0At(source + Offsets[0]), Channels[0])),
		step(Threshold, dot(imageSrc1At(source + Offsets[1]), Channels[1])),
		step(Threshold, dot(imageSrc2At(source + Offsets[2]), Channels[2])),
		step(Threshold, dot(imageSrc3At(source + Offsets[3]), Channels[3])))
}
`

const bitplanePackedShader = `//kage:unit pixels
package main

var Threshold float
var Palette [64]vec4
var Channels [6]vec4
var Offsets [6]vec2

func Fragment(position vec4, source vec2, color vec4) vec4 {
	bits := imageSrc0At(source)
	index := int(bits.r + 2*bits.g + 4*bits.b + 8*bits.a +
		16*step(Threshold, dot(imageSrc1At(source + Offsets[4]), Channels[4])) +
		32*step(Threshold, dot(imageSrc2At(source + Offsets[5]), Channels[5])) + 0.5)
	return Palette[index]
}
`
