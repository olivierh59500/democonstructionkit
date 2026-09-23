package effects

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// TimedCRTConfig extends a bitmap CRT material with animated scanlines,
// chromatic fringe, green phosphor glow and flicker. TimeStep is seconds per
// Update; passing another clock through SetTime allows manual cue alignment.
type TimedCRTConfig struct {
	Curvature, ScanlineFrequency, ScanlineAmplitude, ScanlineTimeRate float32
	ChromaticShift, Vignette, GlowShift, GlowGain                     float32
	FlickerBase, FlickerAmplitude, FlickerRate                        float32
	TimeStep                                                          float64
}

// TimedCRTOverlay owns one shader and borrows its source image for each draw.
type TimedCRTOverlay struct {
	shader   *ebiten.Shader
	uniforms map[string]any
	op       ebiten.DrawRectShaderOptions
	step     float64
	time     float64
}

func NewTimedCRTOverlay(c TimedCRTConfig) (*TimedCRTOverlay, error) {
	for _, value := range []float32{c.Curvature, c.ScanlineFrequency, c.ScanlineAmplitude, c.ScanlineTimeRate, c.ChromaticShift, c.Vignette, c.GlowShift, c.GlowGain, c.FlickerBase, c.FlickerAmplitude, c.FlickerRate} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("effects: nonfinite timed CRT parameter")
		}
	}
	if math.IsNaN(c.TimeStep) || math.IsInf(c.TimeStep, 0) {
		return nil, fmt.Errorf("effects: nonfinite timed CRT step")
	}
	shader, err := ebiten.NewShader([]byte(timedCRTShader))
	if err != nil {
		return nil, fmt.Errorf("effects: compile timed CRT: %w", err)
	}
	return &TimedCRTOverlay{
		shader: shader, step: c.TimeStep,
		uniforms: map[string]any{
			"Time": float32(0), "Curvature": c.Curvature,
			"ScanlineFrequency": c.ScanlineFrequency, "ScanlineAmplitude": c.ScanlineAmplitude,
			"ScanlineTimeRate": c.ScanlineTimeRate, "ChromaticShift": c.ChromaticShift,
			"Vignette": c.Vignette, "GlowShift": c.GlowShift, "GlowGain": c.GlowGain,
			"FlickerBase": c.FlickerBase, "FlickerAmplitude": c.FlickerAmplitude,
			"FlickerRate": c.FlickerRate,
		},
	}, nil
}

func (c *TimedCRTOverlay) Update(kit.Frame) error {
	return c.SetTime(c.time + c.step)
}
func (c *TimedCRTOverlay) SetTime(seconds float64) error {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) || math.IsInf(float64(float32(seconds)), 0) {
		return fmt.Errorf("effects: invalid timed CRT clock")
	}
	c.time = seconds
	c.uniforms["Time"] = float32(seconds)
	return nil
}
func (c *TimedCRTOverlay) Time() float64 { return c.time }

func (c *TimedCRTOverlay) DrawAt(dst, source *ebiten.Image, x, y float64) {
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

func (c *TimedCRTOverlay) Close() error {
	if c != nil && c.shader != nil {
		c.shader.Deallocate()
		c.shader = nil
	}
	return nil
}

const timedCRTShader = `
package main

var Time float
var Curvature float
var ScanlineFrequency float
var ScanlineAmplitude float
var ScanlineTimeRate float
var ChromaticShift float
var Vignette float
var GlowShift float
var GlowGain float
var FlickerBase float
var FlickerAmplitude float
var FlickerRate float

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
	scanline = sin(uv.y * ScanlineFrequency + Time * ScanlineTimeRate) * ScanlineAmplitude
	col.rgb = col.rgb - scanline
	var rShift float
	var bShift float
	rShift = imageSrc0At(uv + vec2(ChromaticShift, 0.0)).r
	bShift = imageSrc0At(uv - vec2(ChromaticShift, 0.0)).b
	col.r = rShift
	col.b = bShift
	var glow float
	glow = imageSrc0At(uv + vec2(GlowShift, GlowShift)).g * GlowGain
	col.g = col.g + glow
	var vignette float
	vignette = 1.0 - dot(dc, dc) * Vignette
	col.rgb = col.rgb * vignette
	var flicker float
	flicker = FlickerBase + sin(Time * FlickerRate) * FlickerAmplitude
	col.rgb = col.rgb * flicker
	return col * color
}
`
