package effects

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// CRTOverlayConfig controls a normalized-coordinate CRT pass over any image.
// Frequencies are shader-coordinate units; all values can be adjusted without
// changing the scene or its scrolling transport.
type CRTOverlayConfig struct {
	Curvature, ScanlineFrequency, ScanlineAmplitude float32
	ChromaticShift, Vignette                        float32
	// NormalizeSource makes sampling independent of the source's texture-atlas
	// placement and clamps the output to valid premultiplied alpha. The default
	// retains the original shader's coordinates for existing productions.
	NormalizeSource bool
	// OutsideTransparent leaves pixels beyond the curved source transparent,
	// allowing this pass to be layered over another image. The default retains
	// the historical opaque black edge used by existing productions.
	OutsideTransparent bool
	// SourceOrigin copies the input into an independent texture with this
	// pixel-space origin before the CRT pass. It reproduces authored atlas
	// offsets without depending on Ebitengine's automatic texture packing.
	// Zero samples the supplied image directly and allocates no copy surface.
	SourceOrigin image.Point
	// Blend selects how the processed image covers the destination. The zero
	// value uses regular alpha blending; BlendCopy replaces the full pass.
	Blend ebiten.Blend
}

// CRTOverlay borrows the input on each DrawAt. It owns its shader and, when a
// source origin is configured, one bounded reusable copy surface.
type CRTOverlay struct {
	shader        *ebiten.Shader
	uniforms      map[string]any
	op            ebiten.DrawRectShaderOptions
	sourceOrigin  image.Point
	sourceSize    image.Point
	sourceBacking *ebiten.Image
	sourceView    *ebiten.Image
	copyOptions   ebiten.DrawImageOptions
}

func NewCRTOverlay(c CRTOverlayConfig) (*CRTOverlay, error) {
	for _, v := range []float32{c.Curvature, c.ScanlineFrequency, c.ScanlineAmplitude, c.ChromaticShift, c.Vignette} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, fmt.Errorf("effects: nonfinite CRT overlay parameter")
		}
	}
	if c.SourceOrigin.X < 0 || c.SourceOrigin.Y < 0 || c.SourceOrigin.X > 8192 || c.SourceOrigin.Y > 8192 {
		return nil, fmt.Errorf("effects: invalid CRT source origin")
	}
	source := crtOverlayShader
	if c.NormalizeSource {
		source = crtOverlayNormalizedShader
	}
	if c.OutsideTransparent {
		source = strings.Replace(source, "return vec4(0.0, 0.0, 0.0, 1.0)", "return vec4(0.0, 0.0, 0.0, 0.0)", 1)
	}
	shader, err := ebiten.NewShader([]byte(source))
	if err != nil {
		return nil, fmt.Errorf("effects: compile CRT overlay: %w", err)
	}
	u := map[string]any{
		"Curvature": c.Curvature, "ScanlineFrequency": c.ScanlineFrequency,
		"ScanlineAmplitude": c.ScanlineAmplitude, "ChromaticShift": c.ChromaticShift,
		"Vignette": c.Vignette,
	}
	return &CRTOverlay{
		shader: shader, uniforms: u, sourceOrigin: c.SourceOrigin,
		op: ebiten.DrawRectShaderOptions{Blend: c.Blend},
	}, nil
}

func (c *CRTOverlay) DrawAt(dst, source *ebiten.Image, x, y float64) {
	if c == nil || c.shader == nil || dst == nil || source == nil {
		return
	}
	if c.sourceOrigin != (image.Point{}) {
		b := source.Bounds()
		size := b.Size()
		if c.sourceBacking == nil || c.sourceSize != size {
			if c.sourceBacking != nil {
				c.sourceBacking.Deallocate()
			}
			c.sourceBacking = ebiten.NewImageWithOptions(
				image.Rect(0, 0, size.X+c.sourceOrigin.X, size.Y+c.sourceOrigin.Y),
				&ebiten.NewImageOptions{Unmanaged: true},
			)
			c.sourceView = c.sourceBacking.SubImage(image.Rectangle{
				Min: c.sourceOrigin, Max: c.sourceOrigin.Add(size),
			}).(*ebiten.Image)
			c.sourceSize = size
		}
		c.sourceView.Clear()
		c.copyOptions.GeoM.Reset()
		c.copyOptions.GeoM.Translate(float64(c.sourceOrigin.X-b.Min.X), float64(c.sourceOrigin.Y-b.Min.Y))
		c.sourceView.DrawImage(source, &c.copyOptions)
		source = c.sourceView
	}
	c.op.Images[0] = source
	c.op.Uniforms = c.uniforms
	c.op.GeoM.Reset()
	c.op.GeoM.Translate(x, y)
	b := source.Bounds()
	dst.DrawRectShader(b.Dx(), b.Dy(), c.shader, &c.op)
}

func (c *CRTOverlay) Close() error {
	if c != nil {
		if c.sourceBacking != nil {
			c.sourceBacking.Deallocate()
			c.sourceBacking, c.sourceView = nil, nil
		}
		if c.shader != nil {
			c.shader.Deallocate()
			c.shader = nil
		}
	}
	return nil
}

const crtOverlayShader = `
package main

var Curvature float
var ScanlineFrequency float
var ScanlineAmplitude float
var ChromaticShift float
var Vignette float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	var uv vec2
	uv = texCoord
	var dc vec2
	dc = uv - 0.5
	dc = dc * (1.0 + dot(dc, dc) * Curvature)
	uv = dc + 0.5
	if uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0 {
		return vec4(0.0, 0.0, 0.0, 1.0)
	}
	var col vec4
	col = imageSrc0At(uv)
	var scanline float
	scanline = sin(uv.y * ScanlineFrequency) * ScanlineAmplitude
	col.rgb = col.rgb - scanline
	var rShift float
	var bShift float
	rShift = imageSrc0At(uv + vec2(ChromaticShift, 0.0)).r
	bShift = imageSrc0At(uv - vec2(ChromaticShift, 0.0)).b
	col.r = rShift
	col.b = bShift
	var vignette float
	vignette = 1.0 - dot(dc, dc) * Vignette
	col.rgb = col.rgb * vignette
	return col * color
}
`

const crtOverlayNormalizedShader = `//kage:unit pixels
package main

var Curvature float
var ScanlineFrequency float
var ScanlineAmplitude float
var ChromaticShift float
var Vignette float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	origin := imageSrc0Origin()
	size := imageSrc0Size()
	uv := (texCoord - origin) / size
	dc := uv - 0.5
	dc *= 1.0 + dot(dc, dc)*Curvature
	uv = dc + 0.5
	if uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0 {
		return vec4(0.0, 0.0, 0.0, 1.0)
	}
	col := imageSrc0At(origin + uv*size)
	scanline := sin(uv.y*ScanlineFrequency) * ScanlineAmplitude
	col.rgb -= scanline
	rShift := imageSrc0At(origin + (uv+vec2(ChromaticShift, 0.0))*size).r
	bShift := imageSrc0At(origin + (uv-vec2(ChromaticShift, 0.0))*size).b
	col.r = rShift
	col.b = bShift
	col.rgb *= 1.0 - dot(dc, dc)*Vignette
	col *= color
	col.rgb = min(max(col.rgb, vec3(0.0)), vec3(col.a))
	return col
}
`
