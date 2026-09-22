package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// MagnifierOptions controls a circular magnification overlay. Coordinates are
// relative to the source image's upper-left corner, which is drawn at (0, 0).
// Crop uses absolute source image coordinates; nil samples the entire source.
// Falloff is the exponent of the radial transition from Zoom at the center to
// 1 at the edge. Zero keeps the same zoom throughout the circle. Feather is an
// inward edge fade in pixels. Opacity is in [0, 1].
type MagnifierOptions struct {
	CenterX, CenterY float64
	Radius, Zoom     float64
	Falloff, Opacity float64
	Feather          float64
	Crop             *image.Rectangle
	Filter           ebiten.Filter
}

// DefaultMagnifierOptions returns a visible, smoothly curved lens. A zero-value
// options struct draws nothing, making an inactive effect cheap to represent.
func DefaultMagnifierOptions() MagnifierOptions {
	return MagnifierOptions{Radius: 64, Zoom: 2, Falloff: 1, Opacity: 1, Feather: 1}
}

// Magnifier draws a live image through a reusable GPU lens. Draw only shades
// the lens bounding box and never reads pixels back to the CPU. The source must
// be distinct from the destination (including their backing images). Render a
// scene to a composite.Pass or another reusable surface before applying it.
// Draw leaves the destination outside the lens untouched; draw the source first
// if it should also form the background. Multiple lenses can share the source.
type Magnifier struct {
	shader   *ebiten.Shader
	uniforms map[string]any
	geometry [4]float32
	style    [4]float32
	crop     [4]float32
	vertices [4]ebiten.Vertex
}

func NewMagnifier() (*Magnifier, error) {
	s, err := ebiten.NewShader([]byte(magnifierShader))
	if err != nil {
		return nil, fmt.Errorf("composite: compile magnifier: %w", err)
	}
	m := &Magnifier{shader: s}
	m.uniforms = map[string]any{"Geometry": m.geometry[:], "Style": m.style[:], "Crop": m.crop[:]}
	return m, nil
}

// Draw composites the lens onto dst. Invalid or invisible options are a no-op.
// FilterLinear enables bilinear sampling; other filters use nearest sampling.
// Draw is intended for the same goroutine as other Ebitengine drawing calls.
func (m *Magnifier) Draw(dst, source *ebiten.Image, options MagnifierOptions) {
	if m == nil || m.shader == nil || dst == nil || source == nil || dst == source {
		return
	}
	for _, v := range [...]float64{options.CenterX, options.CenterY, options.Radius, options.Zoom, options.Falloff, options.Opacity, options.Feather} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > math.MaxFloat32 {
			return
		}
	}
	if options.Radius <= 0 || options.Zoom <= 0 || options.Opacity <= 0 {
		return
	}
	b := source.Bounds()
	crop := b
	if options.Crop != nil {
		crop = crop.Intersect(*options.Crop)
	}
	if crop.Empty() {
		return
	}
	m.geometry = [4]float32{float32(options.CenterX), float32(options.CenterY), float32(options.Radius), float32(options.Zoom)}
	m.style = [4]float32{float32(max(0, options.Falloff)), float32(min(1, options.Opacity)), float32(max(0, min(options.Radius, options.Feather))), 0}
	if options.Filter == ebiten.FilterLinear {
		m.style[3] = 1
	}
	m.crop = [4]float32{float32(crop.Min.X - b.Min.X), float32(crop.Min.Y - b.Min.Y), float32(crop.Max.X - b.Min.X), float32(crop.Max.Y - b.Min.Y)}
	left, top := math.Floor(options.CenterX-options.Radius), math.Floor(options.CenterY-options.Radius)
	right, bottom := math.Ceil(options.CenterX+options.Radius), math.Ceil(options.CenterY+options.Radius)
	// Clip geometry before submitting it; this also keeps very large radii from
	// producing needlessly large rasterization bounds on mobile GPUs.
	db := dst.Bounds()
	left, top = max(left, float64(db.Min.X)), max(top, float64(db.Min.Y))
	right, bottom = min(right, float64(db.Max.X)), min(bottom, float64(db.Max.Y))
	if left >= right || top >= bottom {
		return
	}
	for i, p := range [4][2]float64{{left, top}, {right, top}, {right, bottom}, {left, bottom}} {
		m.vertices[i] = ebiten.Vertex{DstX: float32(p[0]), DstY: float32(p[1]), SrcX: float32(p[0]) + float32(b.Min.X), SrcY: float32(p[1]) + float32(b.Min.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1}
	}
	op := ebiten.DrawTrianglesShaderOptions{Uniforms: m.uniforms}
	op.Images[0] = source
	dst.DrawTrianglesShader(m.vertices[:], magnifierIndices[:], m.shader, &op)
}

func (m *Magnifier) Close() error {
	if m != nil && m.shader != nil {
		m.shader.Deallocate()
		m.shader = nil
	}
	return nil
}

var magnifierIndices = [...]uint16{0, 1, 2, 0, 2, 3}

const magnifierShader = `//kage:unit pixels
package main

var Geometry vec4
var Style vec4
var Crop vec4

func sample(p vec2) vec4 {
	// Clamping each sample prevents color leakage from adjacent atlas regions.
	p = clamp(p, Crop.xy+vec2(.5), Crop.zw-vec2(.5))
	return imageSrc0At(imageSrc0Origin()+p)
}

func Fragment(dst vec4, uv vec2, color vec4) vec4 {
	p := uv-imageSrc0Origin()
	d := p-Geometry.xy
	r := length(d)/Geometry.z
	if r >= 1 { return vec4(0) }
	weight := float(1)
	if Style.x > 0 { weight = pow(max(0, 1-r), Style.x) }
	zoom := 1+(Geometry.w-1)*weight
	q := Geometry.xy+d/zoom
	if q.x < Crop.x || q.y < Crop.y || q.x >= Crop.z || q.y >= Crop.w { return vec4(0) }
	c := sample(q)
	if Style.w > .5 {
		base := floor(q-vec2(.5))+vec2(.5)
		f := q-base
		c = mix(mix(sample(base), sample(base+vec2(1,0)), f.x), mix(sample(base+vec2(0,1)), sample(base+vec2(1,1)), f.x), f.y)
	}
	alpha := Style.y
	if Style.z > 0 { alpha *= clamp((1-r)*Geometry.z/Style.z, 0, 1) }
	return c*alpha
}
`
