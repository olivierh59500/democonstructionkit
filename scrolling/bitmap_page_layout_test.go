package scrolling

import (
	"math"
	"strings"
	"testing"
)

func TestPhenomenaIntroPageLayoutMatchesAuthoredGlyphPlacement(t *testing.T) {
	const order = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"
	for _, lines := range [][]struct {
		y    float64
		text string
	}{
		{{18, "   FOR HOT VHS"}, {75, "  AND SOFTWARE"}, {133, "SWAPPING, CONTACT"}, {219, " THE PUNISHER "}, {291, "      AT..."}},
		{{78, "    PHENOMENA"}, {158, "   SKALDEV. 69"}, {238, "  16142 BROMMA"}, {334, "     SWEDEN!"}},
	} {
		program := make([]BitmapPageLine, len(lines))
		for i, line := range lines {
			program[i] = BitmapPageLine{Text: line.text, X: 48, Y: line.y, Advance: 32, ScaleX: 2, ScaleY: 2}
		}
		got, err := CompileBitmapPageLayout(order, program, false)
		if err != nil {
			t.Fatal(err)
		}
		var want []BitmapPageGlyph
		for _, line := range lines {
			x := 48.0
			for _, character := range line.text {
				if index := strings.IndexRune(order, character); index >= 0 {
					want = append(want, BitmapPageGlyph{Index: index, X: x, Y: line.y, ScaleX: 2, ScaleY: 2})
				}
				x += 32
			}
		}
		if len(got) != len(want) {
			t.Fatalf("page has %d placed glyphs, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("glyph %d = %+v, want %+v", i, got[i], want[i])
			}
		}
	}
}

func TestBitmapPageLayoutRetainsUnsupportedSpacing(t *testing.T) {
	got, err := CompileBitmapPageLayout("AB", []BitmapPageLine{{Text: "A€B", Advance: 16, ScaleX: 1, ScaleY: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].X != 0 || got[1].X != 32 {
		t.Fatalf("unsupported rune changed pen spacing: %+v", got)
	}
	for _, config := range []struct {
		order string
		line  BitmapPageLine
	}{
		{"AA", BitmapPageLine{Text: "A", Advance: 1, ScaleX: 1, ScaleY: 1}},
		{"A", BitmapPageLine{Text: "A", Advance: 1, ScaleX: math.NaN(), ScaleY: 1}},
		{"A", BitmapPageLine{Text: "A", Advance: 0, ScaleX: 1, ScaleY: 1}},
	} {
		if _, err := CompileBitmapPageLayout(config.order, []BitmapPageLine{config.line}, false); err == nil {
			t.Fatalf("accepted invalid bitmap page layout: %+v", config)
		}
	}
}
