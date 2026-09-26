package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// GlyphPagesConfig binds a font to a reusable, font-independent page cycle.
// Offset and Scale place the complete effect on any destination layer. With
// UseAliases false, glyphs are selected literally from the configured atlas.
type GlyphPagesConfig struct {
	Cycle      *motion.GlyphPageCycle
	Font       *scrolling.Atlas
	OffsetX    float64
	OffsetY    float64
	Scale      float64
	Filter     ebiten.Filter
	Blend      ebiten.Blend
	UseAliases bool
}

// GlyphPages draws cached atlas glyphs in stable depth order. It borrows the
// font and cycle, so independent screens may share an atlas without copying it.
type GlyphPages struct{ config GlyphPagesConfig }

func NewGlyphPages(c GlyphPagesConfig) (*GlyphPages, error) {
	if c.Cycle == nil || c.Font == nil || math.IsNaN(c.OffsetX) || math.IsInf(c.OffsetX, 0) ||
		math.IsNaN(c.OffsetY) || math.IsInf(c.OffsetY, 0) || math.IsNaN(c.Scale) || math.IsInf(c.Scale, 0) || c.Scale < 0 {
		return nil, fmt.Errorf("sprites: invalid glyph page renderer")
	}
	if c.Scale == 0 {
		c.Scale = 1
	}
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendSourceOver
	}
	return &GlyphPages{config: c}, nil
}

func (p *GlyphPages) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil {
		return
	}
	c := p.config
	lineHeight := c.Font.Metrics().LineHeight()
	lookup := c.Font.ExactGlyph
	if c.UseAliases {
		lookup = c.Font.Glyph
	}
	for _, index := range c.Cycle.DrawOrder() {
		pose := c.Cycle.Glyph(index)
		if pose.Position.Z <= 0 {
			continue
		}
		glyph, metrics, _ := lookup(pose.Rune)
		if glyph == nil {
			continue
		}
		scale := pose.Position.Z * c.Scale
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(c.OffsetX+pose.Position.X*c.Scale-metrics.Advance*scale/2,
			c.OffsetY+pose.Position.Y*c.Scale-lineHeight*scale/2)
		dst.DrawImage(glyph, &op)
	}
}
