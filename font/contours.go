package font

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// ContourGlyph separates authored outlines from their pen advance. Multiple
// contours preserve holes and disconnected pieces; an empty glyph can represent
// a space. Coordinates and advance use the same caller-defined font units.
type ContourGlyph struct {
	Contours [][]geometry.Vec2
	Advance  float64
}

// ContourBankConfig supplies vector font artwork independently of a renderer.
// Closed removes one repeated endpoint from closed contours during construction;
// it never removes an unduplicated endpoint. A nonzero Fallback must name a glyph.
type ContourBankConfig struct {
	Glyphs   map[rune]ContourGlyph
	Closed   bool
	Fallback rune
}

// ContourBank owns immutable copies of vector glyphs. Glyph returns borrowed,
// read-only contour lists, so ordinary rendering requires no copying or allocation.
type ContourBank struct {
	glyphs   map[rune]ContourGlyph
	chars    []rune
	fallback rune
	vertices int
}

// NewContourBank validates and copies vector artwork. No graphics resources are
// created. Banks are bounded to 65,536 glyphs, 1,048,576 contours and points, and
// 65,536 contours per glyph or points per contour, including repeated endpoints.
func NewContourBank(c ContourBankConfig) (*ContourBank, error) {
	if len(c.Glyphs) == 0 || len(c.Glyphs) > 65536 {
		return nil, fmt.Errorf("font: invalid contour glyph count")
	}
	if c.Fallback != 0 {
		if _, exists := c.Glyphs[c.Fallback]; !exists {
			return nil, fmt.Errorf("font: missing contour fallback %q", c.Fallback)
		}
	}
	pointCount, contourCount := 0, 0
	for character, glyph := range c.Glyphs {
		if !utf8.ValidRune(character) || !finite(glyph.Advance) || glyph.Advance < 0 || len(glyph.Contours) > 65536 {
			return nil, fmt.Errorf("font: invalid contour glyph %q", character)
		}
		contourCount += len(glyph.Contours)
		if contourCount > 1<<20 {
			return nil, fmt.Errorf("font: too many font contours")
		}
		for _, contour := range glyph.Contours {
			if len(contour) > 65536 {
				return nil, fmt.Errorf("font: too many contour points for %q", character)
			}
			pointCount += len(contour)
			if pointCount > 1<<20 {
				return nil, fmt.Errorf("font: too many font contour points")
			}
			for _, point := range contour {
				if !finite(point.X) || !finite(point.Y) {
					return nil, fmt.Errorf("font: nonfinite contour point for %q", character)
				}
			}
		}
	}
	bank := &ContourBank{
		glyphs: make(map[rune]ContourGlyph, len(c.Glyphs)),
		chars:  make([]rune, 0, len(c.Glyphs)), fallback: c.Fallback,
	}
	for character, glyph := range c.Glyphs {
		copied := ContourGlyph{Advance: glyph.Advance, Contours: make([][]geometry.Vec2, len(glyph.Contours))}
		for i, contour := range glyph.Contours {
			if c.Closed && len(contour) > 1 && contour[0] == contour[len(contour)-1] {
				contour = contour[:len(contour)-1]
			}
			copied.Contours[i] = append([]geometry.Vec2(nil), contour...)
			bank.vertices += len(contour)
		}
		bank.glyphs[character] = copied
		bank.chars = append(bank.chars, character)
	}
	sort.Slice(bank.chars, func(i, j int) bool { return bank.chars[i] < bank.chars[j] })
	return bank, nil
}

// Glyph returns immutable, borrowed artwork and whether the character is
// explicitly mapped. An unsupported character receives Fallback when configured,
// with false still returned; otherwise it receives an empty, zero-advance glyph.
func (b *ContourBank) Glyph(character rune) (ContourGlyph, bool) {
	if b == nil {
		return ContourGlyph{}, false
	}
	if glyph, exists := b.glyphs[character]; exists {
		return glyph, true
	}
	if b.fallback != 0 {
		return b.glyphs[b.fallback], false
	}
	return ContourGlyph{}, false
}

// Characters returns a writable copy of the explicitly mapped characters in
// stable Unicode order. Looking up a fallback does not add an extra character.
func (b *ContourBank) Characters() []rune {
	if b == nil {
		return nil
	}
	return append([]rune(nil), b.chars...)
}

// VertexCount returns the total cached point count after endpoint normalization.
func (b *ContourBank) VertexCount() int {
	if b == nil {
		return 0
	}
	return b.vertices
}
