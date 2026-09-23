package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// BitmapSpec is an image-independent atlas recipe, including partial columns.
// Zero Columns derives the sheet width; FractionalColumns retains partial cells.
// Unlike integer Font metrics, fractional cells are sampled without rounding.
type BitmapSpec struct {
	Width, Height     float64
	Columns           int
	FractionalColumns bool
	First             rune
	Order             string
}

// Grid binds an atlas recipe to a borrowed image and its sampling filter.
func (s BitmapSpec) Grid(img *ebiten.Image, filter ebiten.Filter) (BitmapGrid, error) {
	if img == nil || !finite(s.Width) || !finite(s.Height) || s.Width <= 0 || s.Height <= 0 || s.Columns < 0 {
		return BitmapGrid{}, fmt.Errorf("scrolling: invalid bitmap atlas recipe")
	}
	columns := s.Columns
	if columns == 0 {
		columns = int(float64(img.Bounds().Dx()) / s.Width)
	}
	if columns < 1 {
		return BitmapGrid{}, fmt.Errorf("scrolling: bitmap atlas contains no columns")
	}
	span := 0.0
	if s.FractionalColumns {
		span = float64(img.Bounds().Dx()) / s.Width
	}
	return BitmapGrid{Image: img, Width: s.Width, Height: s.Height, Columns: columns, ColumnSpan: span, First: s.First, Order: s.Order, Filter: filter}, nil
}

// BitmapText caches character lookup, including unsupported blank positions.
// It borrows its atlas and creates neither textures nor per-draw allocations.
type BitmapText struct {
	windowBatch *composite.QuadBatch
	grid        BitmapGrid
	regions     []composite.Region
	valid       []bool
	advance     float64
}

// NewBitmapText resolves source regions once; zero advance uses the cell width.
func NewBitmapText(grid BitmapGrid, text string, advance float64) (*BitmapText, error) {
	if grid.Image == nil || grid.Columns < 1 || !finite(grid.Width) || !finite(grid.Height) || grid.Width <= 0 || grid.Height <= 0 || !finite(advance) || advance < 0 {
		return nil, fmt.Errorf("scrolling: invalid bitmap text")
	}
	if advance == 0 {
		advance = grid.Width
	}
	result := &BitmapText{grid: grid, advance: advance}
	for _, ch := range text {
		region, valid := grid.Region(ch)
		result.regions = append(result.regions, region)
		result.valid = append(result.valid, valid)
	}
	return result, nil
}
func (t *BitmapText) Len() int       { return len(t.regions) }
func (t *BitmapText) Width() float64 { return float64(t.Len()) * t.advance }
func (t *BitmapText) DrawAt(dst *ebiten.Image, x, y, sx, sy float64) {
	for i := range t.regions {
		t.DrawGlyph(dst, i, x+float64(i)*t.advance*sx, y, sx, sy)
	}
}

// DrawGlyph draws an already resolved character at an independently chosen pose.
func (t *BitmapText) DrawGlyph(dst *ebiten.Image, index int, x, y, sx, sy float64) {
	if index < 0 || index >= len(t.regions) || !t.valid[index] || dst == nil {
		return
	}
	op := ebiten.DrawImageOptions{Filter: t.grid.Filter}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x, y)
	composite.DrawRegion(dst, t.grid.Image, t.regions[index], &op)
}

// BitmapParagraph compiles all lines once while retaining per-line alignment.
// MaxCharacters clips long lines after alignment, without allocating substrings.
type BitmapParagraphConfig struct {
	Font                                                        BitmapGrid
	Lines                                                       []string
	Width, LineAdvance, GlyphAdvance, AlignmentAdvance, OffsetX float64
	Align                                                       Alignment
	MaxCharacters                                               int
}
type BitmapParagraph struct {
	config  BitmapParagraphConfig
	lines   []*BitmapText
	origins []float64
}

// NewBitmapParagraph compiles a reusable paragraph without creating a texture.
func NewBitmapParagraph(c BitmapParagraphConfig) (*BitmapParagraph, error) {
	if !finite(c.Width) || c.Width < 0 || !finite(c.LineAdvance) || c.LineAdvance <= 0 || !finite(c.OffsetX) || !finite(c.AlignmentAdvance) || c.AlignmentAdvance < 0 || c.MaxCharacters < 0 || c.Align > AlignRight {
		return nil, fmt.Errorf("scrolling: invalid bitmap paragraph")
	}
	p := &BitmapParagraph{config: c}
	for _, line := range c.Lines {
		text, err := NewBitmapText(c.Font, line, c.GlyphAdvance)
		if err != nil {
			return nil, err
		}
		width := text.Width()
		if c.AlignmentAdvance > 0 {
			width = float64(text.Len()) * c.AlignmentAdvance
		}
		x := c.OffsetX
		if c.Align == AlignCenter {
			x += (c.Width - width) / 2
		}
		if c.Align == AlignRight {
			x += c.Width - width
		}
		p.lines = append(p.lines, text)
		p.origins = append(p.origins, x)
	}
	p.config.Lines = nil
	return p, nil
}
func (p *BitmapParagraph) Len() int { return len(p.lines) }
func (p *BitmapParagraph) DrawWindow(dst *ebiten.Image, first, count int, x, y float64) {
	for line := 0; line < count; line++ {
		index := first + line
		if index < 0 || index >= len(p.lines) {
			continue
		}
		text := p.lines[index]
		n := text.Len()
		if p.config.MaxCharacters > 0 {
			n = min(n, p.config.MaxCharacters)
		}
		for glyph := 0; glyph < n; glyph++ {
			text.DrawGlyph(dst, glyph, x+p.origins[index]+float64(glyph)*text.advance, y+float64(line)*p.config.LineAdvance, 1, 1)
		}
	}
}

// Scrolling builds regular Scrolling glyphs from an integer atlas recipe.
// Fractional recipes intentionally use BitmapText or recycled transport instead.
func (g BitmapGrid) Scrolling(text string) (*Scrolling, error) {
	if g.Image == nil || g.Columns < 1 {
		return nil, fmt.Errorf("scrolling: invalid bitmap grid")
	}
	if g.Width != math.Trunc(g.Width) || g.Height != math.Trunc(g.Height) || g.Width <= 0 || g.Height <= 0 {
		return nil, fmt.Errorf("scrolling: integer cells required for glyph scrolling")
	}
	var glyphs []Glyph
	images := make(map[rune]*ebiten.Image)
	for _, ch := range text {
		img, cached := images[ch]
		if !cached {
			region, valid := g.Region(ch)
			if valid {
				rect := image.Rect(int(region.X), int(region.Y), int(region.X+region.Width), int(region.Y+region.Height))
				if rect.In(g.Image.Bounds()) {
					img = g.Image.SubImage(rect).(*ebiten.Image)
				}
			}
			images[ch] = img
		}
		glyphs = append(glyphs, Glyph{Image: img, Rune: ch, Advance: g.Width, ScaleX: 1, ScaleY: 1})
	}
	filter := g.Filter
	return New(Config{Glyphs: glyphs, Map: func(_ Sample, options *ebiten.DrawImageOptions) bool { options.Filter = filter; return true }})
}
