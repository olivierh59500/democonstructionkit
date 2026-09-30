package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// PaletteGridAxis maps source-space pixels to palette cells. CellSize defaults
// to one. Offset is applied first. A positive FirstSpan gives cell zero a
// different initial length, followed by equal CellSize spans. Indices optionally
// remaps those spans to arbitrary palette cells; coordinates clamp at the ends.
type PaletteGridAxis struct {
	CellSize, Offset, FirstSpan float32
	Indices                     []int
}

// PaletteGridConfig gives every cell two independently editable colors. A
// control image selects or mixes them using its stored premultiplied channel.
// Threshold zero blends continuously; a positive threshold selects a binary
// bank. An optional body image replaces its covered pixels with BodyColor.
// Sources stay borrowed; only the small color bank and shader are owned.
type PaletteGridConfig struct {
	Width, Height, Columns, Rows int
	X, Y                         PaletteGridAxis
	ControlChannel, BodyChannel  BitplaneChannel
	ControlThreshold             float32
	BodyThreshold                float32
	BodyColor                    *color.NRGBA
	Blend                        ebiten.Blend
}

// PaletteGridState offsets the control/body samples independently without
// moving the palette grid. Positive offsets read farther right/down.
type PaletteGridState struct {
	ControlOffset, BodyOffset [2]float32
}

// PaletteGrid applies spatial color banks to masks, logos, scrolls or scenery
// in one GPU pass. A single column with one-pixel rows makes a row-palette floor
// or raster material. Wider cells and an optional body mask make shadowed color
// grids. Uploads contain only two colors per cell, never a full-stage raster.
type PaletteGrid struct {
	width, height, columns, rows int
	body                         bool
	shader                       *ebiten.Shader
	palette                      *ebiten.Image
	pixels                       []byte
	offsets                      []float32
	bodyColor                    []float32
	vertices                     [4]ebiten.Vertex
	options                      ebiten.DrawTrianglesShaderOptions
}

func normalizePaletteAxis(a PaletteGridAxis, cells int) (PaletteGridAxis, []float32, []float32, error) {
	if a.CellSize == 0 {
		a.CellSize = 1
	}
	for _, value := range []float32{a.CellSize, a.Offset, a.FirstSpan} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return a, nil, nil, fmt.Errorf("composite: nonfinite palette axis")
		}
	}
	if a.CellSize <= 0 || a.FirstSpan < 0 || len(a.Indices) > 256 {
		return a, nil, nil, fmt.Errorf("composite: invalid palette axis spans")
	}
	mapping := make([]float32, max(1, len(a.Indices)))
	for i, index := range a.Indices {
		if index < 0 || index >= cells {
			return a, nil, nil, fmt.Errorf("composite: palette cell index is out of range")
		}
		mapping[i] = float32(index)
	}
	last := cells - 1
	if len(a.Indices) > 0 {
		last = len(a.Indices) - 1
	}
	return a, []float32{a.Offset, a.FirstSpan, a.CellSize, float32(last)}, mapping, nil
}

func NewPaletteGrid(c PaletteGridConfig) (*PaletteGrid, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 || int64(c.Width)*int64(c.Height) > 16*1024*1024 ||
		c.Columns < 1 || c.Columns > 2048 || c.Rows < 1 || c.Rows > 8192 || int64(c.Columns)*int64(c.Rows) > 512*1024 ||
		c.ControlChannel > BitplaneBlue || c.BodyChannel > BitplaneBlue {
		return nil, fmt.Errorf("composite: invalid palette grid dimensions or channels")
	}
	if c.BodyThreshold == 0 {
		c.BodyThreshold = .5
	}
	if !bandFinite(float64(c.ControlThreshold)) || c.ControlThreshold < 0 || c.ControlThreshold > 1 ||
		!bandFinite(float64(c.BodyThreshold)) || c.BodyThreshold <= 0 || c.BodyThreshold > 1 {
		return nil, fmt.Errorf("composite: invalid palette grid thresholds")
	}
	x, xaxis, xmap, err := normalizePaletteAxis(c.X, c.Columns)
	if err != nil {
		return nil, err
	}
	y, yaxis, ymap, err := normalizePaletteAxis(c.Y, c.Rows)
	if err != nil {
		return nil, err
	}
	shader, err := ebiten.NewShader([]byte(fmt.Sprintf(paletteGridShader, len(xmap), len(ymap))))
	if err != nil {
		return nil, fmt.Errorf("composite: compile palette grid shader: %w", err)
	}
	p := &PaletteGrid{width: c.Width, height: c.Height, columns: c.Columns, rows: c.Rows,
		body: c.BodyColor != nil, shader: shader, palette: render.NewSurface(c.Columns*2, c.Rows),
		pixels: make([]byte, c.Columns*c.Rows*8), offsets: make([]float32, 4), bodyColor: make([]float32, 4)}
	p.vertices = [4]ebiten.Vertex{render.Vertex(0, 0, 0, 0, color.White), render.Vertex(float64(c.Width), 0, float64(c.Width), 0, color.White),
		render.Vertex(0, float64(c.Height), 0, float64(c.Height), color.White), render.Vertex(float64(c.Width), float64(c.Height), float64(c.Width), float64(c.Height), color.White)}
	channels := make([]float32, 8)
	for i, channel := range [2]BitplaneChannel{c.ControlChannel, c.BodyChannel} {
		component := 3
		if channel != BitplaneAlpha {
			component = int(channel) - 1
		}
		channels[4*i+component] = 1
	}
	params := []float32{c.ControlThreshold, c.BodyThreshold, 0}
	if p.body {
		params[2] = 1
		p.setBodyColor(*c.BodyColor)
	}
	mapped := []float32{0, 0}
	if len(x.Indices) != 0 {
		mapped[0] = 1
	}
	if len(y.Indices) != 0 {
		mapped[1] = 1
	}
	p.options.Blend = c.Blend
	p.options.Uniforms = map[string]any{"XAxis": xaxis, "YAxis": yaxis, "XMap": xmap, "YMap": ymap, "Mapped": mapped,
		"Columns": float32(c.Columns), "Params": params, "Channels": channels, "Offsets": p.offsets, "BodyColor": p.bodyColor}
	return p, nil
}

func (p *PaletteGrid) setBodyColor(c color.NRGBA) {
	a := float32(c.A) / 255
	p.bodyColor[0], p.bodyColor[1], p.bodyColor[2], p.bodyColor[3] = float32(c.R)/255*a, float32(c.G)/255*a, float32(c.B)/255*a, a
}

// SetBodyColor changes an enabled replacement material without changing masks.
func (p *PaletteGrid) SetBodyColor(c color.NRGBA) error {
	if p == nil || p.shader == nil || !p.body {
		return fmt.Errorf("composite: palette body material is absent or closed")
	}
	p.setBodyColor(c)
	return nil
}

// SetColors copies row-major first/second cell banks into one small RGBA image.
// Inputs may be reused immediately. Translucent colors are premultiplied before
// upload; updates do not allocate an image or read any mask pixels back.
func (p *PaletteGrid) SetColors(first, second []color.NRGBA) error {
	if p == nil || p.shader == nil || len(first) != p.columns*p.rows || len(second) != len(first) {
		return fmt.Errorf("composite: invalid or closed palette grid color banks")
	}
	for row := 0; row < p.rows; row++ {
		for col := 0; col < p.columns; col++ {
			for bank, c := range [2]color.NRGBA{first[row*p.columns+col], second[row*p.columns+col]} {
				at := (row*p.columns*2 + col + bank*p.columns) * 4
				r, g, b, a := c.RGBA()
				p.pixels[at], p.pixels[at+1], p.pixels[at+2], p.pixels[at+3] = uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
			}
		}
	}
	p.palette.WritePixels(p.pixels)
	return nil
}

var paletteGridIndices = [...]uint16{0, 1, 2, 1, 3, 2}

// Draw combines stage-sized borrowed control/body images at the destination
// origin. Body may be nil when disabled. Sampling outside a mask is zero.
func (p *PaletteGrid) Draw(dst, control, body *ebiten.Image, state PaletteGridState) error {
	if p == nil || p.shader == nil || dst == nil || control == nil || dst == control || p.body && (body == nil || dst == body) {
		return fmt.Errorf("composite: invalid or closed palette grid draw")
	}
	for _, img := range []*ebiten.Image{control, body} {
		if img == nil {
			continue
		}
		if img.Bounds().Min.X != 0 || img.Bounds().Min.Y != 0 || img.Bounds().Dx() != p.width || img.Bounds().Dy() != p.height {
			return fmt.Errorf("composite: palette masks must match the configured stage")
		}
	}
	for _, offset := range [2][2]float32{state.ControlOffset, state.BodyOffset} {
		for _, value := range offset {
			if !bandFinite(float64(value)) {
				return fmt.Errorf("composite: palette mask offset must be finite")
			}
		}
	}
	p.offsets[0], p.offsets[1], p.offsets[2], p.offsets[3] = state.ControlOffset[0], state.ControlOffset[1], state.BodyOffset[0], state.BodyOffset[1]
	p.options.Images = [4]*ebiten.Image{control, body, p.palette}
	dst.DrawTrianglesShader(p.vertices[:], paletteGridIndices[:], p.shader, &p.options)
	return nil
}

func (p *PaletteGrid) Close() error {
	if p == nil {
		return nil
	}
	if p.palette != nil {
		p.palette.Deallocate()
		p.palette = nil
	}
	if p.shader != nil {
		p.shader.Deallocate()
		p.shader = nil
	}
	p.options.Images = [4]*ebiten.Image{}
	p.options.Uniforms = nil
	p.pixels, p.offsets, p.bodyColor = nil, nil, nil
	return nil
}

const paletteGridShader = `//kage:unit pixels
package main

var XAxis vec4
var YAxis vec4
var XMap [%d]float
var YMap [%d]float
var Mapped vec2
var Columns float
var Params vec3
var Channels [2]vec4
var Offsets [2]vec2
var BodyColor vec4

func cell(coordinate float, axis vec4) float {
	v := coordinate + axis.x
	index := floor(v / axis.z)
	if axis.y > 0 {
		index = 0
		if v >= axis.y { index = 1 + floor((v-axis.y)/axis.z) }
	}
	return clamp(index, 0, axis.w)
}

func Fragment(position vec4, source vec2, color vec4) vec4 {
	if Params.z > 0 && dot(imageSrc1At(source + Offsets[1]), Channels[1]) >= Params.y {
		return BodyColor
	}
	p := source - imageSrc0Origin()
	x, y := cell(p.x, XAxis), cell(p.y, YAxis)
	if Mapped.x > 0 { x = XMap[int(x)] }
	if Mapped.y > 0 { y = YMap[int(y)] }
	first := imageSrc2At(imageSrc0Origin() + vec2(x, y) + vec2(0.5))
	second := imageSrc2At(imageSrc0Origin() + vec2(x+Columns, y) + vec2(0.5))
	value := dot(imageSrc0At(source + Offsets[0]), Channels[0])
	if Params.x > 0 { value = step(Params.x, value) }
	return mix(first, second, value)
}
`
