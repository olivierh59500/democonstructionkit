package scrolling

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// ProfiledConfig draws a proportional text ribbon into a bounded surface and
// samples it through a displacement table. The glyph clock and profile clock
// can advance independently. The source window is clipped before each strip
// is drawn, preserving historical partial-width rasterization.
type ProfiledConfig struct {
	Font                        *Atlas
	Text                        string
	GlyphScale, MissingAdvance  float64
	SurfaceWidth, SurfaceHeight int
	OutputWidth, OutputY        int
	StripHeight                 int
	SourceXOrigin               int
	CullMargin                  float64
	TextSpeed, ProfileSpeed     float64 // Pixels and table entries per Update.
	Profile                     []float64
	Filter                      ebiten.Filter
}

// ProfiledScroll owns its text surface, batched strip geometry and independent
// transport clocks. Font atlas images remain borrowed from the caller.
type ProfiledScroll struct {
	config                            ProfiledConfig
	text                              *Scrolling
	surface                           *ebiten.Image
	batch                             *composite.QuadBatch
	profile                           []float64
	textWidth, textX, profilePosition float64
	mapper                            Mapper
}

func NewProfiledScroll(c ProfiledConfig) (*ProfiledScroll, error) {
	if c.Font == nil || c.Font.face.Atlas == nil || c.Text == "" || c.SurfaceWidth < 1 || c.SurfaceWidth > 8192 || c.SurfaceHeight < 1 || c.SurfaceHeight > 8192 || c.OutputWidth < 1 || c.OutputWidth > c.SurfaceWidth || c.StripHeight < 1 || c.SurfaceHeight%c.StripHeight != 0 || len(c.Profile) == 0 || len(c.Profile) > 1<<20 || !finite(c.GlyphScale) || c.GlyphScale <= 0 || !finite(c.MissingAdvance) || c.MissingAdvance < 0 || !finite(c.CullMargin) || c.CullMargin < 0 || !finite(c.TextSpeed) || !finite(c.ProfileSpeed) {
		return nil, fmt.Errorf("scrolling: invalid profiled scroll configuration")
	}
	for _, value := range c.Profile {
		if !finite(value) {
			return nil, fmt.Errorf("scrolling: invalid profiled scroll offset")
		}
	}
	glyphs := make([]Glyph, 0, len(c.Text))
	for _, r := range c.Text {
		glyph := Glyph{Rune: r, Advance: c.MissingAdvance * c.GlyphScale, ScaleX: c.GlyphScale, ScaleY: c.GlyphScale}
		if img, metrics, ok := c.Font.ExactGlyph(r); ok {
			glyph.Image = img
			glyph.Advance = float64(int(metrics.Advance)) * c.GlyphScale
		}
		if !finite(glyph.Advance) || glyph.Advance < 0 || glyph.Advance > 1<<30 {
			return nil, fmt.Errorf("scrolling: profiled glyph advance exceeds budget")
		}
		glyphs = append(glyphs, glyph)
	}
	text, err := New(Config{Glyphs: glyphs})
	if err != nil {
		return nil, err
	}
	c.Profile = append([]float64(nil), c.Profile...)
	p := &ProfiledScroll{
		config: c, text: text, profile: c.Profile, textWidth: text.Length(),
		surface: ebiten.NewImage(c.SurfaceWidth, c.SurfaceHeight),
		batch:   composite.NewQuadBatch(c.SurfaceHeight / c.StripHeight),
	}
	p.batch.AlternateDiagonal = true
	p.batch.Options.Filter = c.Filter
	p.mapper = func(sample Sample, _ *ebiten.DrawImageOptions) bool {
		return sample.X < float64(p.config.SurfaceWidth)+p.config.CullMargin && sample.X+sample.Glyph.Advance > -p.config.CullMargin
	}
	return p, nil
}

func (p *ProfiledScroll) Update(kit.Frame) error {
	if p == nil || p.surface == nil {
		return nil
	}
	nextText := p.textX + p.config.TextSpeed
	nextProfile := p.profilePosition + p.config.ProfileSpeed
	if !finite(nextText) || !finite(nextProfile) {
		return fmt.Errorf("scrolling: profiled transport overflows")
	}
	p.textX = nextText
	if p.textX >= p.textWidth {
		p.textX = 0
	}
	p.profilePosition = nextProfile
	if p.profilePosition >= float64(len(p.profile)) {
		p.profilePosition -= float64(len(p.profile))
	}
	return nil
}

func (p *ProfiledScroll) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil || p.surface == nil {
		return
	}
	p.surface.Clear()
	state := IdentityState()
	state.X = float64(p.config.SurfaceWidth) - p.textX
	state.Map = p.mapper
	p.text.DrawAt(p.surface, state)
	phase := int(p.profilePosition)
	p.batch.Begin(dst, p.surface)
	for y := 0; y < p.config.SurfaceHeight; y += p.config.StripHeight {
		index := (phase + y/p.config.StripHeight) % len(p.profile)
		if index < 0 {
			index += len(p.profile)
		}
		left := int(p.profile[index]) + p.config.SourceXOrigin
		right := left + p.config.OutputWidth
		left = max(0, left)
		right = min(p.config.SurfaceWidth, right)
		if left >= right {
			continue
		}
		r := image.Rect(left, y, right, y+p.config.StripHeight)
		p.batch.Rect(r, 0, float32(p.config.OutputY+y), float32(right-left), float32(p.config.StripHeight))
	}
	p.batch.Flush()
}

func (p *ProfiledScroll) Close() error {
	if p != nil && p.surface != nil {
		p.surface.Deallocate()
		p.surface = nil
	}
	return nil
}
