package font

import (
	"image"
	"testing"
)

func TestSparseUnicodeGridWithGuttersAndFallback(t *testing.T) {
	f, err := NewGrid(Grid{Bounds: image.Rect(10, 20, 50, 50), Origin: image.Pt(2, 3), Cell: image.Pt(4, 6), Stride: image.Pt(5, 7), Columns: 3, Order: "A\x00É?", Uppercase: true, Fallback: '?'})
	if err != nil {
		t.Fatal(err)
	}
	g, ok := f.Glyph('é')
	if !ok || g.Rect != image.Rect(22, 23, 26, 29) {
		t.Fatalf("glyph = %+v, %v", g, ok)
	}
	g, ok = f.Glyph('€')
	if ok || g.Rect != image.Rect(12, 30, 16, 36) {
		t.Fatalf("fallback = %+v, %v", g, ok)
	}
}

func TestProportionalLayoutAndIsolation(t *testing.T) {
	metrics := map[rune]Glyph{'I': {Rect: image.Rect(0, 0, 2, 8), Advance: 3}, 'W': {Rect: image.Rect(2, 0, 10, 8), Advance: 9, OffsetX: -1}}
	f, err := New(Config{Bounds: image.Rect(0, 0, 10, 8), Glyphs: metrics, LineHeight: 10, SpaceAdvance: 4})
	if err != nil {
		t.Fatal(err)
	}
	delete(metrics, 'I')
	l, err := f.Layout("IW\nW I", 1)
	if err != nil {
		t.Fatal(err)
	}
	if l.Width != 18 || l.Height != 20 || l.Glyphs[1].X != 3 || l.Glyphs[4].X != 15 {
		t.Fatalf("layout = %+v", l)
	}
	if _, err = f.Layout("I", -3); err == nil {
		t.Fatal("nonprogressing layout accepted")
	}
}

func TestInvalidAtlasRejected(t *testing.T) {
	for _, g := range []Grid{
		{Bounds: image.Rect(0, 0, 4, 4), Cell: image.Pt(4, 4), Columns: 1, Order: "AB"},
		{Bounds: image.Rect(0, 0, 8, 4), Cell: image.Pt(4, 4), Columns: 2, Order: "AA"},
		{Bounds: image.Rect(0, 0, 8, 4), Cell: image.Pt(4, 4), Columns: 0, Order: "A"},
	} {
		if _, err := NewGrid(g); err == nil {
			t.Fatalf("accepted %+v", g)
		}
	}
}
