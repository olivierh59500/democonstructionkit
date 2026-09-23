package presets

import (
	"image"
	"testing"
)

func TestSourceFontMappings(t *testing.T) {
	tests := []struct {
		id      string
		bounds  image.Rectangle
		char    rune
		rect    image.Rectangle
		advance float64
	}{
		{"bilizir-demo", image.Rect(0, 0, 320, 200), 'a', image.Rect(0, 0, 32, 32), 32},
		{"teamg1-demo", image.Rect(0, 0, 480, 216), 'I', image.Rect(48, 144, 64, 180), 16},
		{"teamg1-demo", image.Rect(0, 0, 480, 216), '#', image.Rect(432, 180, 480, 216), 48},
		{"phenomena-dna-scroll-intro", image.Rect(0, 0, 720, 26), '@', image.Rect(704, 0, 720, 26), 16},
		{"tcb-multi-plane-3d-scroller", image.Rect(0, 0, 320, 200), 'A', image.Rect(96, 99, 128, 132), 32},
		{"grodan-kvack-kvack-demo", image.Rect(0, 0, 240, 200), '\'', image.Rect(120, 0, 144, 33), 24},
		{"grodan-up", image.Rect(0, 0, 330, 227), '#', image.Rect(165, 58, 198, 87), 33},
		{"dma-3d", image.Rect(0, 0, 640, 300), ':', image.Rect(448, 100, 512, 150), 64},
	}
	for _, tt := range tests {
		s, ok := FindFont(tt.id)
		if !ok {
			t.Fatal(tt.id)
		}
		f, err := s.Build(tt.bounds)
		if err != nil {
			t.Fatal(err)
		}
		g, ok := f.Glyph(tt.char)
		if !ok || g.Rect != tt.rect || g.Advance != tt.advance {
			t.Fatalf("%s %q: %+v", tt.id, tt.char, g)
		}
	}
}
func TestMissingVerticalDigitsAreBlank(t *testing.T) {
	s, _ := FindFont("grodan-up")
	f, err := s.Build(image.Rect(0, 0, 330, 227))
	if err != nil {
		t.Fatal(err)
	}
	g, ok := f.Glyph('5')
	if !ok || !g.Rect.Empty() || g.Advance != 33 {
		t.Fatal(g, ok)
	}
}
