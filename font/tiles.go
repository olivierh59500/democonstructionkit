package font

import (
	"fmt"
	"strings"
	"unicode"
)

// TileAlphabet describes glyphs occupying a variable number of adjacent tiles.
// Order is explicit; First and Widths support contiguous character sets instead.
// Fallback is an index in Widths. Ignore lists characters omitted from the stream.
type TileAlphabet struct {
	Order            string
	First            rune
	Widths           []int
	Stride, Fallback int
	Ignore           string
	Uppercase        bool
}

// Compile resolves a text once into tile indices, independent of any renderer.
func (a TileAlphabet) Compile(text string) ([]int, error) {
	order := []rune(a.Order)
	if a.Stride <= 0 || len(a.Widths) == 0 || a.Fallback < 0 || a.Fallback >= len(a.Widths) || (len(order) != 0 && len(order) != len(a.Widths)) {
		return nil, fmt.Errorf("font: invalid tile alphabet")
	}
	indices := make(map[rune]int, len(order))
	for i, r := range order {
		if _, exists := indices[r]; exists {
			return nil, fmt.Errorf("font: duplicate tile character %q", r)
		}
		indices[r] = i
	}
	for _, width := range a.Widths {
		if width < 0 || width > a.Stride {
			return nil, fmt.Errorf("font: tile width exceeds stride")
		}
	}
	result := make([]int, 0, len(text))
	for _, r := range text {
		if strings.ContainsRune(a.Ignore, r) {
			continue
		}
		if a.Uppercase {
			r = unicode.ToUpper(r)
		}
		index := int(r - a.First)
		if len(order) != 0 {
			var ok bool
			index, ok = indices[r]
			if !ok {
				index = a.Fallback
			}
		}
		if index < 0 || index >= len(a.Widths) {
			index = a.Fallback
		}
		for j := 0; j < a.Widths[index]; j++ {
			result = append(result, index*a.Stride+j)
		}
	}
	return result, nil
}
