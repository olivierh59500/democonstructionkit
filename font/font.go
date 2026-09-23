// Package font describes bitmap fonts independently of the rendering backend.
package font

import (
	"fmt"
	"image"
	"math"
	"sort"
	"unicode"
)

// Glyph separates the visible source rectangle from the pen advance and bearing.
// An empty rectangle is useful for spaces and intentional blank characters.
type Glyph struct {
	Rect             image.Rectangle
	Advance          float64
	OffsetX, OffsetY float64
}

// Config describes an atlas. Aliases point directly to existing glyphs.
// Missing characters use Fallback when nonzero, otherwise SpaceAdvance.
type Config struct {
	Bounds                   image.Rectangle
	Glyphs                   map[rune]Glyph
	Aliases                  map[rune]rune
	LineHeight, SpaceAdvance float64
	Fallback                 rune
	Uppercase                bool
}

// Font is immutable after construction and can be shared across effects.
type Font struct{ config Config }

// New validates and copies metrics. No graphics resources are created.
func New(c Config) (*Font, error) {
	if c.Bounds.Empty() || !positive(c.LineHeight) || !positive(c.SpaceAdvance) {
		return nil, fmt.Errorf("font: bounds, line height and space advance must be positive")
	}
	glyphs := make(map[rune]Glyph, len(c.Glyphs)+1)
	for r, g := range c.Glyphs {
		if !positive(g.Advance) || !finite(g.OffsetX) || !finite(g.OffsetY) || (!g.Rect.Empty() && !g.Rect.In(c.Bounds)) {
			return nil, fmt.Errorf("font: invalid metrics for %q", r)
		}
		glyphs[r] = g
	}
	if _, ok := glyphs[' ']; !ok {
		glyphs[' '] = Glyph{Advance: c.SpaceAdvance}
	}
	if c.Fallback != 0 {
		if _, ok := glyphs[c.Fallback]; !ok {
			return nil, fmt.Errorf("font: missing fallback %q", c.Fallback)
		}
	}
	aliases := make(map[rune]rune, len(c.Aliases))
	for a, b := range c.Aliases {
		if _, ok := glyphs[b]; !ok {
			return nil, fmt.Errorf("font: alias %q refers to missing %q", a, b)
		}
		aliases[a] = b
	}
	c.Glyphs, c.Aliases = glyphs, aliases
	return &Font{config: c}, nil
}

// Grid supports sparse layouts (NUL cells), margins, gutters and custom ordering.
// Origin is relative to Bounds.Min. Zero Stride uses Cell. Zero Advance uses Cell.X.
type Grid struct {
	Bounds                            image.Rectangle
	Cell, Stride, Origin              image.Point
	Columns                           int
	Order                             string
	Blanks                            string // Explicitly supported characters with no source pixels.
	Advance, LineHeight, SpaceAdvance float64
	Uppercase                         bool
	Fallback                          rune
	Aliases                           map[rune]rune
}

// NewGrid builds regular atlas metrics without assuming ASCII ordering.
func NewGrid(g Grid) (*Font, error) {
	if g.Columns <= 0 || g.Cell.X <= 0 || g.Cell.Y <= 0 || g.Order == "" {
		return nil, fmt.Errorf("font: invalid grid")
	}
	if g.Stride == (image.Point{}) {
		g.Stride = g.Cell
	}
	if g.Stride.X < g.Cell.X || g.Stride.Y < g.Cell.Y {
		return nil, fmt.Errorf("font: overlapping grid cells")
	}
	if g.Advance == 0 {
		g.Advance = float64(g.Cell.X)
	}
	if g.SpaceAdvance == 0 {
		g.SpaceAdvance = g.Advance
	}
	if g.LineHeight == 0 {
		g.LineHeight = float64(g.Cell.Y)
	}
	c := Config{Bounds: g.Bounds, Glyphs: map[rune]Glyph{}, Aliases: g.Aliases, LineHeight: g.LineHeight, SpaceAdvance: g.SpaceAdvance, Uppercase: g.Uppercase, Fallback: g.Fallback}
	for i, r := range []rune(g.Order) {
		if r == 0 {
			continue
		}
		if _, exists := c.Glyphs[r]; exists {
			return nil, fmt.Errorf("font: duplicate grid character %q", r)
		}
		p := g.Bounds.Min.Add(g.Origin).Add(image.Pt(i%g.Columns*g.Stride.X, i/g.Columns*g.Stride.Y))
		c.Glyphs[r] = Glyph{Rect: image.Rectangle{Min: p, Max: p.Add(g.Cell)}, Advance: g.Advance}
	}
	for _, r := range g.Blanks {
		c.Glyphs[r] = Glyph{Advance: g.SpaceAdvance}
	}
	return New(c)
}

// Glyph returns resolved metrics and whether the character is supported.
// Even unsupported characters receive a useful fallback advance.
func (f *Font) Glyph(r rune) (Glyph, bool) {
	if f.config.Uppercase {
		r = unicode.ToUpper(r)
	}
	if alias, ok := f.config.Aliases[r]; ok {
		r = alias
	}
	if g, ok := f.config.Glyphs[r]; ok {
		return g, true
	}
	if f.config.Fallback != 0 {
		return f.config.Glyphs[f.config.Fallback], false
	}
	return Glyph{Advance: f.config.SpaceAdvance}, false
}

// ExactGlyph looks up a literal atlas entry without aliases, case conversion or
// fallback. It is useful when preserving a case-sensitive authored tile stream.
func (f *Font) ExactGlyph(r rune) (Glyph, bool) { g, ok := f.config.Glyphs[r]; return g, ok }

// Bounds returns the atlas bounds used to validate the font.
func (f *Font) Bounds() image.Rectangle { return f.config.Bounds }

// Characters returns the explicitly mapped characters in stable Unicode order.
// Aliases and automatic case conversion are resolved by Glyph, not duplicated here.
func (f *Font) Characters() []rune {
	chars := make([]rune, 0, len(f.config.Glyphs))
	for r := range f.config.Glyphs {
		chars = append(chars, r)
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
	return chars
}

// LineHeight is the distance between consecutive text lines.
func (f *Font) LineHeight() float64 { return f.config.LineHeight }

// Placement contains an already resolved glyph and its top-left destination.
type Placement struct {
	Rune  rune
	Glyph Glyph
	X, Y  float64
}

// Layout is reusable resolved text. Width excludes trailing letter spacing.
type Layout struct {
	Glyphs        []Placement
	Width, Height float64
}

// Layout resolves Unicode text, newlines and optional extra letter spacing.
// Negative spacing is allowed provided every character still advances forward.
func (f *Font) Layout(text string, spacing float64) (Layout, error) {
	if !finite(spacing) {
		return Layout{}, fmt.Errorf("font: spacing must be finite")
	}
	l := Layout{}
	if text == "" {
		return l, nil
	}
	x, y := 0.0, 0.0
	lineHasGlyph := false
	for _, r := range text {
		if r == '\n' {
			if lineHasGlyph {
				l.Width = math.Max(l.Width, x-spacing)
			}
			x = 0
			y += f.LineHeight()
			lineHasGlyph = false
			continue
		}
		g, _ := f.Glyph(r)
		if g.Advance+spacing <= 0 {
			return Layout{}, fmt.Errorf("font: spacing prevents forward progress for %q", r)
		}
		l.Glyphs = append(l.Glyphs, Placement{Rune: r, Glyph: g, X: x + g.OffsetX, Y: y + g.OffsetY})
		x += g.Advance + spacing
		lineHasGlyph = true
	}
	if lineHasGlyph {
		l.Width = math.Max(l.Width, x-spacing)
	}
	l.Height = y + f.LineHeight()
	return l, nil
}

func finite(v float64) bool   { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func positive(v float64) bool { return finite(v) && v > 0 }
