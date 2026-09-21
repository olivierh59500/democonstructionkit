package sprites

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// Disc is a circular particle in destination pixels, including subpixel radii.
// ColorScale's zero value is white. Slice order controls overlapping particles.
type Disc struct {
	X, Y, Radius float64
	ColorScale   ebiten.ColorScale
}

// Discs batches circles without allocating vector stencil targets per particle.
// Antialias uses eight fixed coverage samples, retaining tiny distant particles.
// Geometry, colors and filtering are independent of a production or font.
type Discs struct {
	Antialias bool
	batch     *composite.QuadBatch
	shader    *ebiten.Shader
}

const discShader = `//kage:unit pixels
package main

var Antialias float

func Fragment(dstPos vec4, srcPos vec2, color vec4, custom vec4) vec4 {
    radius2 := custom.x*custom.x
    coverage := float(0)
    if Antialias > 0 {
        // Same eight-sample pattern as Ebitengine's vector antialiasing.
        offsets := [8]vec2{vec2(1,-3),vec2(-1,3),vec2(5,1),vec2(-3,-5),vec2(-5,5),vec2(-7,-1),vec2(3,7),vec2(7,-7)}
        for i := 0; i < 8; i++ {
            p := srcPos-offsets[i]/16
            if dot(p,p) <= radius2 {coverage += 0.125}
        }
    } else if dot(srcPos,srcPos) <= radius2 {coverage = 1}
    return color*coverage
}
`

func NewDiscs(capacity int) (*Discs, error) {
	shader, err := ebiten.NewShader([]byte(discShader))
	if err != nil {
		return nil, err
	}
	b := composite.NewQuadBatch(capacity)
	b.Shader = shader
	b.ShaderOptions.Uniforms = map[string]any{"Antialias": float32(1)}
	return &Discs{Antialias: true, batch: b, shader: shader}, nil
}
func (d *Discs) DrawAt(dst *ebiten.Image, points []Disc, x, y float64) {
	if d.shader == nil || dst == nil {
		return
	}
	aa := float32(0)
	if d.Antialias {
		aa = 1
	}
	d.batch.ShaderOptions.Uniforms["Antialias"] = aa
	d.batch.Begin(dst, nil)
	for _, p := range points {
		if p.Radius <= 0 {
			continue
		}
		r := float32(p.Radius)
		extent := r + 1
		corners := [4][2]float32{{-extent, -extent}, {extent, -extent}, {extent, extent}, {-extent, extent}}
		var vertices [4]ebiten.Vertex
		for i, c := range corners {
			vertices[i] = ebiten.Vertex{DstX: float32(x+p.X) + c[0], DstY: float32(y+p.Y) + c[1], SrcX: c[0], SrcY: c[1], Custom0: r, ColorR: p.ColorScale.R(), ColorG: p.ColorScale.G(), ColorB: p.ColorScale.B(), ColorA: p.ColorScale.A()}
		}
		d.batch.Quad(vertices)
	}
	d.batch.Flush()
}
func (d *Discs) Close() error {
	if d.shader != nil {
		d.shader.Deallocate()
		d.shader = nil
	}
	return nil
}
