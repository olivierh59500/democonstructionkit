package scrolling

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"testing"
)

func TestBitmapRecipeRetainsFractionalAtlasArithmetic(t *testing.T) {
	image := ebiten.NewImage(1000, 82)
	defer image.Deallocate()
	g, err := (BitmapSpec{Width: 83.25, Height: 41, FractionalColumns: true, First: 32}).Grid(image, ebiten.FilterLinear)
	if err != nil {
		t.Fatal(err)
	}
	region, ok := g.Region(rune(32 + 13))
	if !ok {
		t.Fatal("missing fractional cell")
	}
	if region.X != math.Floor(math.Mod(13, 1000/83.25))*83.25 || region.Y != 41 || region.Width != 83.25 || g.Filter != ebiten.FilterLinear {
		t.Fatalf("fractional recipe changed: %+v", region)
	}
	if _, err := g.Scrolling("A"); err == nil {
		t.Fatal("fractional grid must not silently round into integer glyphs")
	}
}
func TestBitmapTextCustomOrderAndBlankAdvances(t *testing.T) {
	atlas := ebiten.NewImage(48, 16)
	defer atlas.Deallocate()
	grid, err := (BitmapSpec{Width: 16, Height: 16, Order: "Z A"}).Grid(atlas, ebiten.FilterNearest)
	if err != nil {
		t.Fatal(err)
	}
	text, err := NewBitmapText(grid, "AZ?", 20)
	if err != nil {
		t.Fatal(err)
	}
	if text.Len() != 3 || text.Width() != 60 || text.regions[0].X != 32 || text.regions[1].X != 0 || text.valid[2] {
		t.Fatalf("wrong compiled layout: %+v", text)
	}
	scroll, err := grid.Scrolling("AZ?")
	if err != nil {
		t.Fatal(err)
	}
	if scroll.Length() != 48 {
		t.Fatalf("blank advance lost: %g", scroll.Length())
	}
	if scroll.glyphs[0].Image.Bounds() != image.Rect(32, 0, 48, 16) || scroll.glyphs[2].Image != nil {
		t.Fatal("wrong atlas glyph resolution")
	}
}
func TestBitmapParagraphAlignsBeforeClipping(t *testing.T) {
	atlas := ebiten.NewImage(1024, 64)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 17, Height: 11, First: 32}).Grid(atlas, ebiten.FilterLinear)
	p, err := NewBitmapParagraph(BitmapParagraphConfig{Font: grid, Lines: []string{"HELLO", "LONGER THAN A WINDOW"}, Width: 300, LineAdvance: 17, GlyphAdvance: 20, AlignmentAdvance: 16, OffsetX: -8, Align: AlignCenter, MaxCharacters: 12})
	if err != nil {
		t.Fatal(err)
	}
	if p.origins[0] != 102 || p.origins[1] != (300-20*16)/2.0-8 {
		t.Fatalf("alignment changed: %v", p.origins)
	}
	if p.lines[1].Len() != 20 {
		t.Fatal("long line truncated before centering")
	}
}
func TestRevealOrderingAndIndependentStart(t *testing.T) {
	atlas := ebiten.NewImage(512, 64)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 16, Height: 16, First: 32}).Grid(atlas, ebiten.FilterNearest)
	y := 500.0
	r, err := NewReveal(RevealConfig{Font: grid, Lines: []string{"AB", "C"}, Columns: 2, X: 154, Y: 358, AdvanceX: 16, AdvanceY: 16, UniformStartY: &y, Delay: 30, Duration: 50, Order: RevealColumnsBottomFirst})
	if err != nil {
		t.Fatal(err)
	}
	want := []int{2, 0, 3, 1}
	for i, g := range r.glyphs {
		if g.index != want[i] || g.startY != 500 {
			t.Fatalf("glyph %d: %+v", i, g)
		}
	}
	y = 0
	if r.glyphs[0].startY != 500 {
		t.Fatal("reveal retained mutable start pointer")
	}
	if r.glyphs[0].y != 374 || r.glyphs[1].y != 358 {
		t.Fatal("bottom-first order changed layout")
	}
}

func TestBitmapScrollingCachesImagesAndRetainsFilter(t *testing.T) {
	atlas := ebiten.NewImage(32, 16)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 16, Height: 16, Order: "AB"}).Grid(atlas, ebiten.FilterLinear)
	scrolling, err := grid.Scrolling("AABA")
	if err != nil {
		t.Fatal(err)
	}
	if scrolling.glyphs[0].Image != scrolling.glyphs[1].Image || scrolling.glyphs[0].Image != scrolling.glyphs[3].Image {
		t.Fatal("repeated characters created duplicate image wrappers")
	}
	state := IdentityState()
	count := 0
	state.Paint = func(_ *ebiten.Image, _ Sample, op ebiten.DrawImageOptions) {
		count++
		if op.Filter != ebiten.FilterLinear {
			t.Fatal("font filter lost")
		}
	}
	scrolling.DrawAt(nil, state)
	if count != 4 {
		t.Fatalf("painted %d glyphs", count)
	}
}
