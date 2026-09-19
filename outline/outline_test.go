package outline

import (
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestOutlineTextHasClosedContoursAndWhitespaceAdvance(t *testing.T) {
	f, err := New(goregular.TTF, 32, 8)
	if err != nil {
		t.Fatal(err)
	}
	g, err := f.Glyph('O')
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Segments) < 16 || g.Advance <= 0 {
		t.Fatal(g)
	}
	// Every contour endpoint must have a matching outgoing segment.
	for _, s := range g.Segments {
		found := false
		for _, other := range g.Segments {
			if s.B == other.A {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("open contour at %+v", s.B)
		}
	}
	a, _, err := f.Text("AA", 0)
	if err != nil {
		t.Fatal(err)
	}
	b, edges, err := f.Text("A A", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) || b[len(b)-1].X <= a[len(a)-1].X {
		t.Fatal("space did not advance")
	}
	for _, e := range edges {
		if e[0] < 0 || e[1] >= len(b) {
			t.Fatal(e)
		}
	}
}
