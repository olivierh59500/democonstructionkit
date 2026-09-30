package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// QuantizedColorMode selects an integer color-grid operation. Passthrough uses
// an ordinary image draw and preserves every source byte. The other modes round
// the source to the configured grid before changing its color components.
type QuantizedColorMode uint8

const (
	QuantizedPassthrough QuantizedColorMode = iota
	QuantizedKeep
	QuantizedReplace
	QuantizedScale
	QuantizedFromTarget
)

// QuantizedColorConfig gives each RGB channel its greatest integer component.
// For RGB12, use {15,15,15}; {31,63,31} models RGB565. A zero array defaults to
// RGB12. AlphaThreshold discards transformed pixels below this alpha; zero
// keeps all nontransparent pixels. Passthrough bypasses that cutoff.
type QuantizedColorConfig struct {
	Levels         [3]uint16
	AlphaThreshold float32
	Blend          ebiten.Blend
}

// QuantizedColorState describes a sample without owning a clock. Scale computes
// floor(source*Numerator/Denominator). FromTarget computes
// Target-floor((Target-source)*Numerator/Denominator), reproducing integer fades
// from a color such as white. Numerator must be within 0..Denominator; the
// denominator may be 1..65536. Target uses the configured channel grid, not
// eight-bit RGB. Replace clamps Target to the grid after replacement.
type QuantizedColorState struct {
	Mode                   QuantizedColorMode
	Target                 [3]uint16
	Numerator, Denominator uint32
}

// QuantizedColor applies palette-style integer transitions to any RGBA layer.
// It borrows source and destination images and owns one shader, with no working
// canvas or CPU pixel readback. Translucent inputs are converted to straight RGB
// before the grid operation and returned as premultiplied colors.
// Instances are confined to the graphics goroutine. A timeline, music signal or
// caller can supply states without resetting another effect's transport.
type QuantizedColor struct {
	levels  [3]uint16
	shader  *ebiten.Shader
	target  []float32
	params  []float32
	options ebiten.DrawRectShaderOptions
	blend   ebiten.Blend
}

func NewQuantizedColor(c QuantizedColorConfig) (*QuantizedColor, error) {
	if c.Levels == [3]uint16{} {
		c.Levels = [3]uint16{15, 15, 15}
	}
	for _, level := range c.Levels {
		if level == 0 || level > 255 {
			return nil, fmt.Errorf("composite: color grid levels must be within 1..255")
		}
	}
	if math.IsNaN(float64(c.AlphaThreshold)) || c.AlphaThreshold < 0 || c.AlphaThreshold > 1 {
		return nil, fmt.Errorf("composite: invalid quantized-color alpha cutoff")
	}
	shader, err := ebiten.NewShader([]byte(quantizedColorShader))
	if err != nil {
		return nil, fmt.Errorf("composite: compile quantized-color shader: %w", err)
	}
	q := &QuantizedColor{levels: c.Levels, shader: shader, target: make([]float32, 3),
		params: []float32{0, 0, 1, c.AlphaThreshold}, blend: c.Blend}
	q.options.Blend = c.Blend
	q.options.Uniforms = map[string]any{
		"Levels": []float32{float32(c.Levels[0]), float32(c.Levels[1]), float32(c.Levels[2])},
		"Target": q.target, "Params": q.params,
	}
	return q, nil
}

func (q *QuantizedColor) validate(state QuantizedColorState) error {
	if state.Mode > QuantizedFromTarget {
		return fmt.Errorf("composite: unknown quantized-color mode %d", state.Mode)
	}
	if state.Mode == QuantizedScale || state.Mode == QuantizedFromTarget {
		if state.Denominator == 0 || state.Denominator > 65536 || state.Numerator > state.Denominator {
			return fmt.Errorf("composite: invalid integer color ratio")
		}
	}
	if state.Mode == QuantizedFromTarget {
		for i, target := range state.Target {
			if target > q.levels[i] {
				return fmt.Errorf("composite: fade target exceeds color grid")
			}
		}
	}
	return nil
}

// Draw renders a borrowed source at the destination origin, honoring its crop.
// Source and destination must be distinct images. State validation happens
// before drawing, leaving the destination untouched when parameters are invalid.
func (q *QuantizedColor) Draw(dst, source *ebiten.Image, state QuantizedColorState) error {
	if q == nil || q.shader == nil || dst == nil || source == nil || dst == source {
		return fmt.Errorf("composite: invalid or closed quantized-color image pass")
	}
	if err := q.validate(state); err != nil {
		return err
	}
	if state.Mode == QuantizedPassthrough {
		dst.DrawImage(source, &ebiten.DrawImageOptions{Blend: q.blend})
		return nil
	}
	q.params[0], q.params[1], q.params[2] = float32(state.Mode), float32(state.Numerator), 1
	if state.Denominator != 0 {
		q.params[2] = float32(state.Denominator)
	}
	for i, value := range state.Target {
		q.target[i] = float32(value)
	}
	q.options.Images[0] = source
	dst.DrawRectShader(source.Bounds().Dx(), source.Bounds().Dy(), q.shader, &q.options)
	return nil
}

// Close releases the shader and borrowed references, leaving input images alive.
func (q *QuantizedColor) Close() error {
	if q != nil {
		if q.shader != nil {
			q.shader.Deallocate()
			q.shader = nil
		}
		q.options.Images = [4]*ebiten.Image{}
		q.options.Uniforms = nil
		q.target, q.params = nil, nil
	}
	return nil
}

const quantizedColorShader = `//kage:unit pixels
package main

var Levels vec3
var Target vec3
var Params vec4

func Fragment(position vec4, source vec2, color vec4) vec4 {
	c := imageSrc0At(source)
	if c.a == 0 || c.a < Params.w { return vec4(0) }
	value := floor(c.rgb/c.a * Levels + vec3(0.5))
	if Params.x == 2 { value = Target }
	if Params.x == 3 { value = floor(value * Params.y / Params.z) }
	if Params.x == 4 { value = Target - floor((Target-value) * Params.y / Params.z) }
	value = clamp(value, vec3(0), Levels)
	return vec4(value/Levels*c.a, c.a)
}
`
