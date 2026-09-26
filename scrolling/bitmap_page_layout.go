package scrolling

import (
	"fmt"
	"math"
	"unicode"
)

// BitmapPageLine places a static line using one atlas order. Unsupported
// characters still consume Advance, preserving authored word spacing.
type BitmapPageLine struct {
	Text           string
	X, Y           float64
	Advance        float64
	ScaleX, ScaleY float64
}

type BitmapPageGlyph struct {
	Index          int
	X, Y           float64
	ScaleX, ScaleY float64
}

// CompileBitmapPageLayout resolves character order and positions once. Its
// output can feed the retained page renderer or a caller's custom material.
func CompileBitmapPageLayout(order string, lines []BitmapPageLine, uppercase bool) ([]BitmapPageGlyph, error) {
	characters := []rune(order)
	if len(characters) == 0 || len(characters) > 1<<20 || len(lines) > 1<<16 {
		return nil, fmt.Errorf("scrolling: invalid bitmap page alphabet or line count")
	}
	index := make(map[rune]int, len(characters))
	for i, character := range characters {
		if uppercase {
			character = unicode.ToUpper(character)
		}
		if _, exists := index[character]; exists {
			return nil, fmt.Errorf("scrolling: duplicate bitmap page character")
		}
		index[character] = i
	}
	var placements []BitmapPageGlyph
	for _, line := range lines {
		for _, value := range [...]float64{line.X, line.Y, line.Advance, line.ScaleX, line.ScaleY} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("scrolling: nonfinite bitmap page placement")
			}
		}
		if line.Advance <= 0 || line.ScaleX <= 0 || line.ScaleY <= 0 || len(line.Text) > 1<<20 {
			return nil, fmt.Errorf("scrolling: invalid bitmap page line")
		}
		position := line.X
		for _, character := range line.Text {
			if uppercase {
				character = unicode.ToUpper(character)
			}
			if glyph, ok := index[character]; ok {
				if len(placements) >= 1<<20 {
					return nil, fmt.Errorf("scrolling: bitmap page exceeds glyph budget")
				}
				placements = append(placements, BitmapPageGlyph{
					Index: glyph, X: position, Y: line.Y, ScaleX: line.ScaleX, ScaleY: line.ScaleY,
				})
			}
			position += line.Advance
			if math.IsNaN(position) || math.IsInf(position, 0) {
				return nil, fmt.Errorf("scrolling: bitmap page pen overflow")
			}
		}
	}
	return placements, nil
}
