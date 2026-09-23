package scrolling

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
)

func TestAtlasSharesViewsAndPreservesOffsetsAliasesAndFallback(t *testing.T) {
	root := ebiten.NewImage(64, 32)
	defer root.Deallocate()
	img := root.SubImage(image.Rect(7, 5, 47, 21)).(*ebiten.Image)
	metrics, err := font.New(font.Config{Bounds: img.Bounds(), LineHeight: 16, SpaceAdvance: 4, Uppercase: true, Fallback: '?', Aliases: map[rune]rune{'@': 'A'}, Glyphs: map[rune]font.Glyph{
		'A': {Rect: image.Rect(9, 6, 17, 19), Advance: 10, OffsetX: -2, OffsetY: 3},
		'?': {Rect: image.Rect(23, 5, 31, 21), Advance: 9},
	}})
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewAtlas(img, metrics)
	if err != nil {
		t.Fatal(err)
	}
	x, g, ok := a.Glyph('a')
	if !ok || x.Bounds() != image.Rect(9, 6, 17, 19) || g.Advance != 10 || g.OffsetX != -2 {
		t.Fatal(x.Bounds(), g, ok)
	}
	y, _, _ := a.Glyph('@')
	if x != y {
		t.Fatal("alias recreated its image view")
	}
	_, _, ok = a.ExactGlyph('a')
	if ok {
		t.Fatal("literal lookup unexpectedly folded case")
	}
	blank, space, ok := a.Glyph(' ')
	if !ok || blank != nil || space.Advance != 4 {
		t.Fatal(blank, space, ok)
	}
	fallback, _, ok := a.Glyph('!')
	known, _, _ := a.Glyph('?')
	if ok || fallback != known {
		t.Fatal("fallback lost shared pixels")
	}
	glyphs := a.Glyphs("a @!")
	if len(glyphs) != 4 || glyphs[0].Image != x || glyphs[0].X != -2 || glyphs[0].Y != 3 || glyphs[1].Advance != 4 || glyphs[3].Image != fallback {
		t.Fatal(glyphs)
	}
	if got := testing.AllocsPerRun(100, func() { a.Glyph('a'); a.Glyph('!'); a.ExactGlyph('A') }); got != 0 {
		t.Fatal("lookup allocations:", got)
	}
	horizontal := a.Layout("A! A", AtlasText{Literal: true, SkipMissing: true})
	if len(horizontal) != 3 || horizontal[1].Image != nil || horizontal[2].Offset != 14 {
		t.Fatal("literal horizontal layout changed supported spacing", horizontal)
	}
	vertical := a.Layout("A! A", AtlasText{Literal: true, Vertical: true})
	if len(vertical) != 4 || vertical[1].Image != nil || vertical[1].Advance != 16 || vertical[3].Offset != 48 {
		t.Fatal("vertical layout did not preserve missing-character slots", vertical)
	}
}

func TestAtlasValidatesImageBoundsAndGridOrigins(t *testing.T) {
	root := ebiten.NewImage(32, 32)
	defer root.Deallocate()
	img := root.SubImage(image.Rect(4, 8, 28, 24)).(*ebiten.Image)
	views, err := GridImages(img, image.Pt(8, 8), 3, 6)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range views {
		want := image.Rect(4+i%3*8, 8+i/3*8, 12+i%3*8, 16+i/3*8)
		if v.Bounds() != want {
			t.Fatal(i, v.Bounds(), want)
		}
	}
	if _, err = GridImages(img, image.Pt(8, 8), 3, 7); err == nil {
		t.Fatal("accepted overflowing tile count")
	}
	metrics, _ := font.NewGrid(font.Grid{Bounds: root.Bounds(), Cell: image.Pt(8, 8), Columns: 1, Order: "A"})
	if _, err = NewAtlas(img, metrics); err == nil {
		t.Fatal("accepted metrics outside the atlas")
	}
	if _, err = NewAtlas(root, nil); err == nil {
		t.Fatal("accepted nil metrics")
	}
}

func TestAtlasCachesLiteralGlyphsBeforeCaseFolding(t *testing.T) {
	img := ebiten.NewImage(16, 8)
	defer img.Deallocate()
	metrics, err := font.New(font.Config{Bounds: img.Bounds(), LineHeight: 8, SpaceAdvance: 4, Uppercase: true, Glyphs: map[rune]font.Glyph{
		'A': {Rect: image.Rect(0, 0, 8, 8), Advance: 8}, 'a': {Rect: image.Rect(8, 0, 16, 8), Advance: 8},
	}})
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewAtlas(img, metrics)
	if err != nil {
		t.Fatal(err)
	}
	literal, _, ok := a.ExactGlyph('a')
	if !ok || literal == nil || literal.Bounds() != image.Rect(8, 0, 16, 8) {
		t.Fatal("literal view was replaced by its folded glyph")
	}
	folded, _, _ := a.Glyph('a')
	if folded == nil || folded == literal {
		t.Fatal("folded lookup lost its independent view")
	}
}
