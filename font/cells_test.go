package font

import (
	"image"
	"image/color"
	"testing"
)

func TestCellBankCachesOrderedAliasesAndUnicode(t *testing.T) {
	calls := 0
	b, err := NewCellBank(CellBankConfig{Width: 3, Height: 2, Characters: "Aa雪 ",
		Key: func(r rune) int {
			if r == 'a' {
				return int('A')
			}
			return int(r)
		},
		Pixel: func(r rune, x, y int) bool { calls++; return r != ' ' && (x+y)%2 == 0 },
		Color: func(r rune, x, y int) color.NRGBA { return color.NRGBA{R: uint8(x * 20), G: uint8(y * 80), A: 255} },
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 18 {
		t.Fatalf("sampled aliases more than once: %d calls", calls)
	}
	want := []image.Point{{0, 0}, {2, 0}, {1, 1}}
	for i, cell := range b.Glyph('A') {
		if image.Pt(cell.X, cell.Y) != want[i] {
			t.Fatal("cell order changed", b.Glyph('A'))
		}
	}
	if len(b.Glyph('雪')) != 3 || len(b.Glyph(' ')) != 0 || len(b.Glyph('?')) != 0 {
		t.Fatal("cell lookup changed blank or Unicode glyphs")
	}
	if &b.Glyph('A')[0] != &b.Glyph('a')[0] {
		t.Fatal("aliases did not share cached artwork")
	}
	if b.Glyph('A')[2].Color.G != 80 {
		t.Fatal("cell colors were not cached")
	}
	if got := testing.AllocsPerRun(100, func() { _ = b.Glyph('雪') }); got != 0 {
		t.Fatalf("glyph lookup allocates: %v", got)
	}
}

func TestImageCellBankKeepsMetricsColorsAndCpuSnapshot(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(3, 4, 10, 9))
	pixels.SetNRGBA(4, 5, color.NRGBA{R: 200, B: 60, A: 255})
	pixels.SetNRGBA(5, 5, color.NRGBA{G: 220, A: 64})
	m, err := New(Config{Bounds: pixels.Bounds(), Glyphs: map[rune]Glyph{
		'A': {Rect: image.Rect(4, 5, 7, 7), Advance: 4},
	}, Aliases: map[rune]rune{'a': 'A'}, LineHeight: 3, SpaceAdvance: 2})
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewImageCellBank(pixels, m, "Aa ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if b.Width() != 3 || b.Height() != 2 || len(b.Glyph('A')) != 1 || b.Glyph('A')[0].Color.R != 200 {
		t.Fatal("atlas crop, threshold or colors changed", b.Glyph('A'))
	}
	pixels.SetNRGBA(4, 5, color.NRGBA{})
	if b.Glyph('a')[0].Color.R != 200 {
		t.Fatal("CPU image mutation changed compiled font")
	}
}

func TestCellBankRejectsInvalidInput(t *testing.T) {
	for _, c := range []CellBankConfig{{}, {Width: 1, Height: 1, Characters: "AA", Pixel: func(rune, int, int) bool { return false }},
		{Width: 4096, Height: 4096, Characters: "A", Pixel: func(rune, int, int) bool { return false }},
		{Width: 1, Height: 1, Characters: string([]byte{255}), Pixel: func(rune, int, int) bool { return false }}} {
		if _, err := NewCellBank(c); err == nil {
			t.Fatal("accepted invalid cell bank")
		}
	}
}
