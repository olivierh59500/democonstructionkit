package effects

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// CRTOverlayConfig controls a normalized-coordinate CRT pass over any image.
// Frequencies are shader-coordinate units; all values can be adjusted without
// changing the scene or its scrolling transport.
type CRTOverlayConfig struct {
	Curvature, ScanlineFrequency, ScanlineAmplitude float32
	ChromaticShift, Vignette                        float32
}

// CRTOverlay borrows the source on each DrawAt and owns only its shader.
type CRTOverlay struct {
	shader   *ebiten.Shader
	uniforms map[string]any
	op       ebiten.DrawRectShaderOptions
}

func NewCRTOverlay(c CRTOverlayConfig) (*CRTOverlay, error) {
	for _, v := range []float32{c.Curvature, c.ScanlineFrequency, c.ScanlineAmplitude, c.ChromaticShift, c.Vignette} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, fmt.Errorf("effects: nonfinite CRT overlay parameter")
		}
	}
	shader, err := ebiten.NewShader([]byte(crtOverlayShader))
	if err != nil {
		return nil, fmt.Errorf("effects: compile CRT overlay: %w", err)
	}
	u := map[string]any{
		"Curvature": c.Curvature, "ScanlineFrequency": c.ScanlineFrequency,
		"ScanlineAmplitude": c.ScanlineAmplitude, "ChromaticShift": c.ChromaticShift,
		"Vignette": c.Vignette,
	}
	return &CRTOverlay{shader: shader, uniforms: u}, nil
}

func (c *CRTOverlay) DrawAt(dst, source *ebiten.Image, x, y float64) {
	if c == nil || c.shader == nil || dst == nil || source == nil {
		return
	}
	c.op.Images[0] = source
	c.op.Uniforms = c.uniforms
	c.op.GeoM.Reset()
	c.op.GeoM.Translate(x, y)
	b := source.Bounds()
	dst.DrawRectShader(b.Dx(), b.Dy(), c.shader, &c.op)
}

func (c *CRTOverlay) Close() error {
	if c != nil && c.shader != nil {
		c.shader.Deallocate()
		c.shader = nil
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
