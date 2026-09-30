package composite

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// GradientBarsConfig lays out independent vertical bars from a shared baseline.
// Level returns the current height in destination pixels and is sampled once
// per column during each Draw. Step places columns; Width leaves configurable
// gaps. Top and Bottom color the filled quad's vertices independently.
//
// Texture is borrowed; nil creates one owned white pixel. Empty Source uses the
// texture bounds. OutlineWidth defaults to one, and a nil OutlineColor selects
// white. Outlines use four inset strips with no corner overlap; bars at most
// twice the outline width are omitted rather than producing inverted strips.
type GradientBarsConfig struct {
	Columns                  int
	X, Baseline, Step, Width float64
	Level                    func(column int) float64
	Texture                  *ebiten.Image
	Source                   image.Rectangle
	Top, Bottom              color.NRGBA
	OutlineWidth             float64
	OutlineColor             *color.NRGBA
	Blend                    ebiten.Blend
	Filter                   ebiten.Filter
	Address                  ebiten.Address
}

// GradientBars owns reusable geometry storage while borrowing its level source
// and optional material. Drawing never updates levels or reads back pixels.
type GradientBars struct {
	config       GradientBarsConfig
	texture      *ebiten.Image
	ownedTexture bool
	batch        *render.Batch
	fill, stroke [4]ebiten.Vertex
}

func NewGradientBars(c GradientBarsConfig) (*GradientBars, error) {
	if c.OutlineWidth == 0 {
		c.OutlineWidth = 1
	}
	if c.Columns < 1 || c.Columns > 65536 || c.Level == nil || c.Step <= 0 || c.Width <= 0 || c.OutlineWidth <= 0 ||
		!barCoordinate(c.X) || !barCoordinate(c.Baseline) || !barCoordinate(c.Step) || !barCoordinate(c.Width) ||
		!barCoordinate(c.OutlineWidth) || !barCoordinate(c.X+float64(c.Columns-1)*c.Step+c.Width) {
		return nil, fmt.Errorf("composite: invalid gradient bar layout")
	}
	texture := c.Texture
	owned := texture == nil
	bounds := image.Rect(0, 0, 1, 1)
	if !owned {
		bounds = texture.Bounds()
	}
	if c.Source.Empty() {
		c.Source = bounds
	}
	if !c.Source.In(bounds) {
		return nil, fmt.Errorf("composite: gradient bar source outside its material")
	}
	if owned {
		texture = ebiten.NewImage(1, 1)
		texture.Fill(color.White)
	}
	outline := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if c.OutlineColor != nil {
		outline = *c.OutlineColor
	}
	c.Texture, c.OutlineColor = nil, nil
	g := &GradientBars{config: c, texture: texture, ownedTexture: owned, batch: render.NewBatch(min(20000, c.Columns*8))}
	g.batch.Options.Blend, g.batch.Options.Filter, g.batch.Options.Address = c.Blend, c.Filter, c.Address
	u, v := float64(c.Source.Min.X), float64(c.Source.Min.Y)
	u2, v2 := float64(c.Source.Max.X), float64(c.Source.Max.Y)
	g.fill = [4]ebiten.Vertex{
		render.Vertex(0, 0, u, v, c.Top), render.Vertex(0, 0, u2, v, c.Top),
		render.Vertex(0, 0, u2, v2, c.Bottom), render.Vertex(0, 0, u, v2, c.Bottom),
	}
	g.stroke = [4]ebiten.Vertex{
		render.Vertex(0, 0, u, v, outline), render.Vertex(0, 0, u2, v, outline),
		render.Vertex(0, 0, u2, v2, outline), render.Vertex(0, 0, u, v2, outline),
	}
	return g, nil
}

func (g *GradientBars) Draw(dst *ebiten.Image) { g.draw(dst, false) }

func (g *GradientBars) DrawOutline(dst *ebiten.Image) { g.draw(dst, true) }

func (g *GradientBars) draw(dst *ebiten.Image, outlined bool) {
	if g == nil || g.texture == nil || dst == nil || dst == g.texture {
		return
	}
	g.batch.Begin(dst, g.texture)
	for column := 0; column < g.config.Columns; column++ {
		height := g.config.Level(column)
		top := g.config.Baseline - height
		if height <= 0 || !barCoordinate(height) || !barCoordinate(top) {
			continue
		}
		x, width := g.config.X+float64(column)*g.config.Step, g.config.Width
		if !outlined {
			g.batch.Quad(barQuad(g.fill, x, top, width, height))
			continue
		}
		stroke := g.config.OutlineWidth
		if width <= 2*stroke || height <= 2*stroke {
			continue
		}
		g.batch.Quad(barQuad(g.stroke, x, top, width, stroke))
		g.batch.Quad(barQuad(g.stroke, x, top+height-stroke, width, stroke))
		g.batch.Quad(barQuad(g.stroke, x, top+stroke, stroke, height-2*stroke))
		g.batch.Quad(barQuad(g.stroke, x+width-stroke, top+stroke, stroke, height-2*stroke))
	}
	g.batch.Flush()
}

func barQuad(vertices [4]ebiten.Vertex, x, y, width, height float64) [4]ebiten.Vertex {
	vertices[0].DstX, vertices[0].DstY = float32(x), float32(y)
	vertices[1].DstX, vertices[1].DstY = float32(x+width), float32(y)
	vertices[2].DstX, vertices[2].DstY = float32(x+width), float32(y+height)
	vertices[3].DstX, vertices[3].DstY = float32(x), float32(y+height)
	return vertices
}

func barCoordinate(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -math.MaxFloat32 && value <= math.MaxFloat32
}

// Close is idempotent and never deallocates a caller-owned material.
func (g *GradientBars) Close() error {
	if g != nil && g.texture != nil {
		if g.ownedTexture {
			g.texture.Deallocate()
		}
		g.texture, g.batch = nil, nil
		g.config.Level = nil
	}
	return nil
}
