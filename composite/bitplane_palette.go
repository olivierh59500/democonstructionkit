package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// BitplanePaletteConfig describes a palette lookup over one to six binary image
// planes. Plane zero is the least significant bit. Each source must have the
// configured dimensions and a zero origin; its alpha selects a bit. A zero
// threshold selects 0.5. Palette entries may be translucent.
type BitplanePaletteConfig struct {
	Width, Height int
	Planes        int
	Threshold     float32
	Palette       []color.NRGBA
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
		return nil, fmt.Errorf("composite: invalid bitplane alpha threshold")
	}
	b := &BitplanePalette{width: c.Width, height: c.Height, planes: c.Planes,
		threshold: c.Threshold, palette: make([]float32, 64*4)}
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
		b.packOptions.Uniforms = map[string]any{"Threshold": b.threshold}
	}
	b.options.Blend = c.Blend
	b.options.Uniforms = map[string]any{"Palette": b.palette, "Threshold": b.threshold}
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
	if b == nil || b.shader == nil || dst == nil {
		return fmt.Errorf("composite: bitplane target is nil or closed")
	}
	if len(planes) != b.planes {
		return fmt.Errorf("composite: expected %d bitplane images, got %d", b.planes, len(planes))
	}
	for _, plane := range planes {
		if plane == nil || plane.Bounds().Min.X != 0 || plane.Bounds().Min.Y != 0 ||
			plane.Bounds().Dx() != b.width || plane.Bounds().Dy() != b.height {
			return fmt.Errorf("composite: bitplane image bounds must be (0,0)-(%d,%d)", b.width, b.height)
		}
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
	return nil
}

const bitplaneDirectShader = `//kage:unit pixels
package main

var Threshold float
var Palette [64]vec4

func Fragment(position vec4, source vec2, color vec4) vec4 {
	index := int(step(Threshold, imageSrc0At(source).a) +
		2*step(Threshold, imageSrc1At(source).a) +
		4*step(Threshold, imageSrc2At(source).a) +
		8*step(Threshold, imageSrc3At(source).a) + 0.5)
	return Palette[index]
}
`

const bitplanePackShader = `//kage:unit pixels
package main

var Threshold float

func Fragment(position vec4, source vec2, color vec4) vec4 {
	return vec4(step(Threshold, imageSrc0At(source).a),
		step(Threshold, imageSrc1At(source).a),
		step(Threshold, imageSrc2At(source).a),
		step(Threshold, imageSrc3At(source).a))
}
`

const bitplanePackedShader = `//kage:unit pixels
package main

var Threshold float
var Palette [64]vec4

func Fragment(position vec4, source vec2, color vec4) vec4 {
	bits := imageSrc0At(source)
	index := int(bits.r + 2*bits.g + 4*bits.b + 8*bits.a +
		16*step(Threshold, imageSrc1At(source).a) +
		32*step(Threshold, imageSrc2At(source).a) + 0.5)
	return Palette[index]
}
`
