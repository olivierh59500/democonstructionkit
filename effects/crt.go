package effects

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

// CRTConfig controls the optional Go-like Kage shader. Zero values disable each
// component, giving a neutral pass-through without changing the source effect.
type CRTConfig struct{ Curvature, Scanlines, Chromatic, Vignette float32 }
type CRT struct {
	Source   kit.Effect
	Config   CRTConfig
	canvas   *ebiten.Image
	shader   *ebiten.Shader
	uniforms map[string]any
}

func NewCRT(source kit.Effect, width, height int, c CRTConfig) (*CRT, error) {
	if source == nil || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("effects: invalid CRT surface")
	}
	shader, err := ebiten.NewShader([]byte(crtShader))
	if err != nil {
		return nil, fmt.Errorf("effects: compile CRT: %w", err)
	}
	return &CRT{Source: source, Config: c, canvas: render.NewSurface(width, height), shader: shader, uniforms: map[string]any{}}, nil
}
func (c *CRT) Update(f kit.Frame) error { return c.Source.Update(f) }
func (c *CRT) Draw(dst *ebiten.Image) {
	c.canvas.Clear()
	c.Source.Draw(c.canvas)
	c.uniforms["Curvature"] = c.Config.Curvature
	c.uniforms["Scanlines"] = c.Config.Scanlines
	c.uniforms["Chromatic"] = c.Config.Chromatic
	c.uniforms["Vignette"] = c.Config.Vignette
	op := ebiten.DrawRectShaderOptions{Uniforms: c.uniforms}
	op.Images[0] = c.canvas
	b := c.canvas.Bounds()
	dst.DrawRectShader(b.Dx(), b.Dy(), c.shader, &op)
}
func (c *CRT) Close() error { c.canvas.Deallocate(); c.shader.Deallocate(); return kit.Close(c.Source) }

const crtShader = `//kage:unit pixels
package main

var Curvature float
var Scanlines float
var Chromatic float
var Vignette float

func Fragment(dst vec4, uv vec2, color vec4) vec4 {
	origin, size := imageSrc0Origin(), imageSrc0Size()
	p := (uv-origin)/size*2-1
	p *= 1+Curvature*dot(p,p)
	q := (p+1)*size/2+origin
	if abs(p.x)>1 || abs(p.y)>1 { return vec4(0) }
	c := imageSrc0At(q)
	c.r = imageSrc0At(q+vec2(Chromatic,0)).r
	c.b = imageSrc0At(q-vec2(Chromatic,0)).b
	shade := (1-Scanlines*(.5+.5*cos(q.y*3.14159265)))*clamp(1-Vignette*dot(p,p),0,1)
	return vec4(c.rgb*shade,c.a)*color
}
`
