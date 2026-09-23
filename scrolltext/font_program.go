package scrolltext

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// FontProgram compiles font controls into visible glyph positions. It is useful
// when several synchronized text surfaces show different scales or materials
// of the same message. Regular mixed-font text can use scrolling.New directly.
// The immutable program has no graphics resources and is safe for concurrent use.
type FontProgram struct {
	text    []rune
	runs    []fontRun
	initial string
}

type fontRun struct {
	start int
	font  string
}

// NewFontProgram accepts any control syntax through Decoder, including DomSizes
// and Braces. Font controls take no glyph space. Other control kinds are rejected
// rather than silently losing their timing or visual meaning.
func NewFontProgram(text string, decoder Decoder, initial string) (*FontProgram, error) {
	tokens, err := Parse(text, decoder)
	if err != nil {
		return nil, err
	}
	p := &FontProgram{initial: initial}
	for _, token := range tokens {
		switch token.Kind {
		case Text:
			p.text = append(p.text, []rune(token.Text)...)
		case Font:
			if token.Text == "" {
				return nil, fmt.Errorf("scrolltext: empty font selection")
			}
			p.runs = append(p.runs, fontRun{len(p.text), token.Text})
		default:
			return nil, fmt.Errorf("scrolltext: font program does not support control kind %d", token.Kind)
		}
	}
	return p, nil
}

func (p *FontProgram) Len() int { return len(p.text) }

// FontAt reports the selection at a visible glyph index. Negative positions
// use the initial font; positions beyond the message keep its final selection.
func (p *FontProgram) FontAt(glyph int) string {
	i := sort.Search(len(p.runs), func(i int) bool { return p.runs[i].start > glyph })
	if i == 0 {
		return p.initial
	}
	return p.runs[i-1].font
}

// MaskedText keeps glyphs belonging to font and substitutes blank elsewhere.
// Controls are removed; all returned strings retain the same glyph positions.
// This is a construction-time operation; cache the result for rendering.
func (p *FontProgram) MaskedText(font string, blank rune) string {
	var result strings.Builder
	result.Grow(len(p.text))
	active, next := p.initial, 0
	if !utf8.ValidRune(blank) {
		blank = utf8.RuneError
	}
	for i, r := range p.text {
		for next < len(p.runs) && p.runs[next].start <= i {
			active = p.runs[next].font
			next++
		}
		if active != font {
			r = blank
		}
		result.WriteRune(r)
	}
	return result.String()
}
